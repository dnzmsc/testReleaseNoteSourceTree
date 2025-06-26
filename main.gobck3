package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	cmdAuthor := exec.Command("git", "config", "user.name")
	outAuthor, err := cmdAuthor.Output()
	if err != nil {
		author = os.Getenv("USER")
		if author == "" {
			author = os.Getenv("USERNAME")
		}
		if author == "" {
			author = "Sconosciuto"
		}
		fmt.Printf("Avviso: Impossibile ottenere user.name da Git (%v). Usando l'utente di sistema: %s\n", err, author)
	} else {
		author = strings.TrimSpace(string(outAuthor))
	}

	commitMsgContent, err := os.ReadFile(commitMsgFilePath)
	if err != nil {
		return "", "", "", fmt.Errorf("errore nella lettura del file del messaggio di commit (%s): %w", commitMsgFilePath, err)
	}
	lines := strings.Split(string(commitMsgContent), "\n")
	var cleanedDescription []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "#") && trimmedLine != "" {
			cleanedDescription = append(cleanedDescription, line)
		}
	}
	description = strings.TrimSpace(strings.Join(cleanedDescription, "\n"))

	date = time.Now().Format("2006-01-02 15:04:05 -0700")

	return author, description, date, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Release Notes Generator")
	w.Resize(fyne.NewSize(600, 700))

	if len(os.Args) < 2 {
		dialog.ShowError(fmt.Errorf("Questa applicazione deve essere eseguita come un hook 'prepare-commit-msg' e necessita del percorso del file del messaggio di commit come argomento."), w)
		return
	}
	commitMsgFilePath := os.Args[1]

	modules, err := loadModules("modules.json")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

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

	// Label per la descrizione che cambia dinamicamente
	descrizioneLabel := widget.NewLabel("Descrizione (min. 20 caratteri):")

	// Listener per il checkbox excludedFromReleaseNote
	excludedFromReleaseNote.OnChanged = func(checked bool) {
		if checked {
			descrizioneLabel.SetText("Descrizione:") // Rimuove il limite visivo
			modulo.Enable()                          // Ri-abilita il modulo se era disabilitato (nel caso di logica futura)
		} else {
			descrizioneLabel.SetText("Descrizione (min. 20 caratteri):") // Ripristina il limite visivo
		}
		// Aggiorna il layout dopo aver cambiato il testo della label
		descrizioneLabel.Refresh()
	}

	commitAuthorLabel := widget.NewLabel(fmt.Sprintf("Autore Commit: %s", commitAuthor))
	commitDescLabel := widget.NewMultiLineEntry()
	commitDescLabel.SetText(fmt.Sprintf("Messaggio Commit: %s", commitDesc))
	commitDescLabel.Disable()
	commitDateLabel := widget.NewLabel(fmt.Sprintf("Data Commit: %s", commitDate))

	var atLeastOneNoteSaved bool
	var currentNoteSaved bool

	resetForm := func() {
		tipo.SetSelected("")
		modulo.SetSelected("")
		titolo.SetText("")
		descrizione.SetText("")
		internalTicket.SetText("")
		clientTicket.SetText("")
		excludedFromReleaseNote.SetChecked(false)
		currentNoteSaved = false
	}

	saveNote := func() {
		if currentNoteSaved {
			dialog.ShowInformation("Attenzione", "Questa nota è già stata salvata. Inserisci una nuova nota o chiudi.", w)
			return
		}

		isExcluded := excludedFromReleaseNote.Checked

		// Validazioni condizionali
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
		} else { // Se è esclusa, controlla solo che la descrizione non sia vuota
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

		if content, err := os.ReadFile(filePath); err == nil {
			_ = json.Unmarshal(content, &relFile)
		}

		relFile.Releases = append(relFile.Releases, release)

		if out, err := json.MarshalIndent(relFile, "", "  "); err == nil {
			os.WriteFile(filePath, out, 0644)
			atLeastOneNoteSaved = true
			currentNoteSaved = true
			dialog.ShowInformation("Successo", "Release salvata correttamente!", w)
		} else {
			dialog.ShowError(fmt.Errorf("Errore nel salvataggio: %v", err), w)
		}
	}

	saveBtn := widget.NewButton("Salva", func() {
		saveNote()
	})

	addNoteBtn := widget.NewButton("Altra Nota", func() {
		saveNote()
		resetForm()
	})

	closeBtn := widget.NewButton("Chiudi", func() {
		if !atLeastOneNoteSaved {
			dialog.ShowConfirm("Attenzione", "Nessuna nota è stata salvata. Vuoi chiudere senza salvare e annullare il commit?", func(b bool) {
				if b {
					os.Exit(1)
				}
			}, w)
		} else {
			a.Quit()
		}
	})

	// Layout dei campi, usando la label dinamica
	formLayout := container.NewVBox(
		excludedFromReleaseNote,
		widget.NewLabel("Tipo:"), tipo,
		widget.NewLabel("Modulo:"), modulo,
		widget.NewLabel("Titolo (facoltativo):"), titolo,
		descrizioneLabel, // Usa la label dinamica qui
		descrizione,
		widget.NewLabel("Numero Ticket Interno (facoltativo):"), internalTicket,
		widget.NewLabel("Numero Ticket Cliente (facoltativo):"), clientTicket,
		widget.NewSeparator(),
		commitAuthorLabel,
		commitDateLabel,
		commitDescLabel,
		widget.NewSeparator(),
	)

	buttonLayout := container.NewGridWithColumns(3, saveBtn, addNoteBtn, closeBtn)

	w.SetContent(container.NewVScroll(container.NewVBox(
		formLayout,
		buttonLayout,
	)))

	// Imposta lo stato iniziale della label della descrizione
	// È importante farlo dopo che il widget è stato creato e aggiunto al layout,
	// ma qui lo facciamo all'inizio perché il suo valore iniziale dipende solo dallo stato del checkbox
	// che è false di default.
	excludedFromReleaseNote.OnChanged(excludedFromReleaseNote.Checked) // Esegui la funzione una volta all'avvio

	w.ShowAndRun()
}
