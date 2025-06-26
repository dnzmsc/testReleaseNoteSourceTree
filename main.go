package main

import (
	"encoding/json"
	"fmt"
	"os" // Usiamo os.ReadFile e os.WriteFile
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Release rappresenta la struttura di una singola nota di rilascio.
type Release struct {
	Data                    string `json:"data"`
	Tipo                    string `json:"tipo"`
	Modulo                  string `json:"modulo"`
	Titolo                  string `json:"titolo,omitempty"`
	Descrizione             string `json:"descrizione"`
	InternalTicket          string `json:"internalTicket,omitempty"`
	ClientTicket            string `json:"clientTicket,omitempty"`
	CommitAuthor            string `json:"commitAuthor"`
	CommitDesc              string `json:"commitDesc"`
	CommitDate              string `json:"commitDate"`
	ExcludedFromReleaseNote bool   `json:"excludedFromReleaseNote"`
}

// ReleaseFile è la struttura per il file JSON che contiene tutte le release.
type ReleaseFile struct {
	Releases []Release `json:"releases"`
}

// loadModules carica i moduli da un file JSON esterno.
func loadModules(filePath string) ([]string, error) {
	// Sostituito ioutil.ReadFile con os.ReadFile
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file moduli: %w", err)
	}

	var modules []string
	err = json.Unmarshal(content, &modules)
	if err != nil {
		return nil, fmt.Errorf("errore nel parsing del file moduli: %w", err)
	}
	return modules, nil
}

// getGitData recupera i dati rilevanti per il commit attuale.
// commitMsgFilePath è il percorso al file temporaneo di Git contenente il messaggio.
func getGitData(commitMsgFilePath string) (author, description, date string, err error) {
	// 1. Ottieni l'autore del commit dall'ambiente Git.
	cmdAuthor := exec.Command("git", "config", "user.name")
	outAuthor, err := cmdAuthor.Output()
	if err != nil {
		author = os.Getenv("USER") // Linux/macOS
		if author == "" {
			author = os.Getenv("USERNAME") // Windows
		}
		if author == "" {
			author = "Sconosciuto" // Fallback
		}
		fmt.Printf("Avviso: Impossibile ottenere user.name da Git (%v). Usando l'utente di sistema: %s\n", err, author)
	} else {
		author = strings.TrimSpace(string(outAuthor))
	}

	// 2. Ottieni la descrizione del commit dal file temporaneo passato come argomento.
	// Sostituito ioutil.ReadFile con os.ReadFile
	commitMsgContent, err := os.ReadFile(commitMsgFilePath)
	if err != nil {
		return "", "", "", fmt.Errorf("errore nella lettura del file del messaggio di commit (%s): %w", commitMsgFilePath, err)
	}
	// Pulisci il messaggio dai commenti (linee che iniziano con #)
	lines := strings.Split(string(commitMsgContent), "\n")
	var cleanedDescription []string
	for _, line := range lines {
		// Ignora le linee vuote dopo la pulizia dei commenti
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "#") && trimmedLine != "" {
			cleanedDescription = append(cleanedDescription, line)
		}
	}
	description = strings.TrimSpace(strings.Join(cleanedDescription, "\n"))

	// 3. Ottieni la data e ora attuale.
	date = time.Now().Format("2006-01-02 15:04:05 -0700")

	return author, description, date, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Release Notes Generator")
	w.Resize(fyne.NewSize(600, 700))

	// L'hook prepare-commit-msg passa il percorso del file del messaggio di commit come primo argomento.
	if len(os.Args) < 2 {
		dialog.ShowError(fmt.Errorf("Questa applicazione deve essere eseguita come un hook 'prepare-commit-msg' e necessita del percorso del file del messaggio di commit come argomento."), w)
		return
	}
	commitMsgFilePath := os.Args[1] // Il primo argomento è il percorso del file del messaggio.

	// Carica i moduli all'avvio
	modules, err := loadModules("modules.json")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	// Recupera i dati del commit attuale all'avvio dell'applicazione
	commitAuthor, commitDesc, commitDate, err := getGitData(commitMsgFilePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Impossibile recuperare i dati del commit attuale. Assicurati di eseguire l'applicazione in un repository Git valido. Errore: %v", err), w)
		return
	}

	// Campi del form
	excludedFromReleaseNote := widget.NewCheck("Escluso dalla Release Note", nil)
	tipi := []string{"Feature", "Fix", "Refactor", "Docs", "Chore"}
	tipo := widget.NewSelect(tipi, nil)
	modulo := widget.NewSelect(modules, nil)
	titolo := widget.NewEntry()
	descrizione := widget.NewMultiLineEntry()
	internalTicket := widget.NewEntry()
	clientTicket := widget.NewEntry()

	// Campi del commit (non editabili)
	commitAuthorLabel := widget.NewLabel(fmt.Sprintf("Autore Commit: %s", commitAuthor))
	commitDescLabel := widget.NewMultiLineEntry()
	commitDescLabel.SetText(fmt.Sprintf("Messaggio Commit: %s", commitDesc))
	commitDescLabel.Disable()
	commitDateLabel := widget.NewLabel(fmt.Sprintf("Data Commit: %s", commitDate))

	var atLeastOneNoteSaved bool // Indica se almeno una nota è stata salvata con successo
	var currentNoteSaved bool    // Indica se la nota attualmente visualizzata è stata salvata

	// Funzione per resettare i campi del form per una nuova nota
	resetForm := func() {
		tipo.SetSelected("")
		modulo.SetSelected("")
		titolo.SetText("")
		descrizione.SetText("")
		internalTicket.SetText("")
		clientTicket.SetText("")
		excludedFromReleaseNote.SetChecked(false)
		currentNoteSaved = false // La nuova nota non è ancora stata salvata
	}

	// Funzione per salvare la nota
	saveNote := func() {
		// Se la nota corrente è già stata salvata, non fare nulla per evitare duplicati.
		if currentNoteSaved {
			dialog.ShowInformation("Attenzione", "Questa nota è già stata salvata. Inserisci una nuova nota o chiudi.", w)
			return
		}

		isExcluded := excludedFromReleaseNote.Checked

		if !isExcluded {
			if tipo.Selected == "" {
				dialog.ShowError(fmt.Errorf("Tipo non selezionato"), w)
				return
			}
			if modulo.Selected == "" {
				dialog.ShowError(fmt.Errorf("Modulo non selezionato"), w)
				return
			}
			if len(descrizione.Text) < 20 {
				dialog.ShowError(fmt.Errorf("Descrizione troppo corta (minimo 20 caratteri)"), w)
				return
			}
		} else {
			if strings.TrimSpace(descrizione.Text) == "" {
				dialog.ShowError(fmt.Errorf("Descrizione non può essere vuota, anche se esclusa"), w)
				return
			}
		}

		release := Release{
			Data:                    time.Now().Format("2006-01-02"),
			Tipo:                    tipo.Selected,
			Modulo:                  modulo.Selected,
			Titolo:                  strings.TrimSpace(titolo.Text),
			Descrizione:             strings.TrimSpace(descrizione.Text),
			InternalTicket:          strings.TrimSpace(internalTicket.Text),
			ClientTicket:            strings.TrimSpace(clientTicket.Text),
			CommitAuthor:            commitAuthor,
			CommitDesc:              commitDesc,
			CommitDate:              commitDate,
			ExcludedFromReleaseNote: isExcluded,
		}

		filePath := "release_notes.json"
		var relFile ReleaseFile

		// Qui già usavamo os.ReadFile, quindi è a posto
		if content, err := os.ReadFile(filePath); err == nil {
			_ = json.Unmarshal(content, &relFile)
		}

		relFile.Releases = append(relFile.Releases, release)

		if out, err := json.MarshalIndent(relFile, "", "  "); err == nil {
			// Qui già usavamo os.WriteFile, quindi è a posto
			os.WriteFile(filePath, out, 0644)
			atLeastOneNoteSaved = true
			currentNoteSaved = true // Imposta a true dopo il salvataggio riuscito
			dialog.ShowInformation("Successo", "Release salvata correttamente!", w)
		} else {
			dialog.ShowError(fmt.Errorf("Errore nel salvataggio: %v", err), w)
		}
	}

	// Bottoni
	saveBtn := widget.NewButton("Salva", func() {
		saveNote()
	})

	addNoteBtn := widget.NewButton("Altra Nota", func() {
		saveNote()  // Tenta di salvare la nota corrente (se non già salvata)
		resetForm() // Resetta il form per una nuova nota
	})

	closeBtn := widget.NewButton("Chiudi", func() {
		if !atLeastOneNoteSaved {
			dialog.ShowConfirm("Attenzione", "Nessuna nota è stata salvata. Vuoi chiudere senza salvare e annullare il commit?", func(b bool) {
				if b {
					os.Exit(1) // Uscita con errore per annullare il commit
				} else {
					// L'utente ha cliccato "No", l'applicazione rimane aperta.
				}
			}, w)
		} else {
			a.Quit() // Almeno una nota è stata salvata, chiudi l'applicazione normalmente.
		}
	})

	// Layout dei campi
	formLayout := container.NewVBox(
		excludedFromReleaseNote,
		widget.NewLabel("Tipo:"), tipo,
		widget.NewLabel("Modulo:"), modulo,
		widget.NewLabel("Titolo (facoltativo):"), titolo,
		widget.NewLabel("Descrizione (min. 20 caratteri):"), descrizione,
		widget.NewLabel("Numero Ticket Interno (facoltativo):"), internalTicket,
		widget.NewLabel("Numero Ticket Cliente (facoltativo):"), clientTicket,
		widget.NewSeparator(),
		commitAuthorLabel,
		commitDateLabel,
		commitDescLabel,
		widget.NewSeparator(),
	)

	// Layout dei bottoni
	buttonLayout := container.NewGridWithColumns(3, saveBtn, addNoteBtn, closeBtn)

	w.SetContent(container.NewVScroll(container.NewVBox(
		formLayout,
		buttonLayout,
	)))

	w.ShowAndRun()
}
