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

// --- IMPORTANTE: ORDINE DELLE DICHIARAZIONI ALL'INTERNO DI main() ---
// Per evitare errori "undefined" dovuti allo scope e ai riferimenti incrociati,
// le variabili e le funzioni devono essere definite *prima* di essere utilizzate.

// Sequenza logica garantita per questo codice:
// 1. Recupero dati Git.
// 2. Dichiarazione delle variabili per i campi del form (Entry, Select, Check),
//    e delle variabili per i *pulsanti* (definite come puntatori inizialmente).
// 3. Dichiarazione delle Label per i dettagli del commit.
// 4. Dichiarazione delle variabili di stato (es. isAnyNoteSavedForCurrentCommit).
// 5. Dichiarazione e inizializzazione dell'etichetta di stato della form (formStatusLabel).
// 6. Definizione delle funzioni helper (resetForm).
// 7. **CRITICO: La funzione 'saveNoteFunc' DEVE essere definita QUI,
//    prima che venga inizializzata la logica dei pulsanti che la chiamano.**
// 8. **Inizializzazione dei pulsanti (assegnando loro le istanze di widget.NewButton),
//    ORA 'saveNoteFunc' è definita e può essere usata nei loro handler.**
// 9. Definizione delle funzioni helper per le label (createMandatoryLabel, createOptionalLabel).
// 10. Creazione dei contenitori di layout dei campi (mainFormFields, commitDetailsGroup, formLayout, contentScroll).
// 11. **Creazione del 'buttonContainer' che raggruppa i pulsanti. Tutti i pulsanti devono essere inizializzati PRIMA.**
// 12. Impostazione del contenuto della finestra 'w.SetContent()'.
// 13. Chiamate di inizializzazione finali (es. excludedFromReleaseNote.OnChanged).
// ----------------------------------------------------------------------

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
	CommitHash              string `json:"commitHash"` // Nuovo campo per l'hash del commit
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
		// Se il file non esiste, ritorna un modulo di default o un errore gestibile
		if os.IsNotExist(err) {
			fmt.Printf("File moduli non trovato: %s. Utilizzo moduli di default.\n", filePath)
			return []string{"Default Module"}, nil // Puoi mettere qui un modulo di fallback
		}
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
func getGitData(commitMsgFilePath string) (author, description, date, commitHash string, err error) {
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
		return "", "", "", "", fmt.Errorf("errore nella lettura del file del messaggio di commit (%s): %w", commitMsgFilePath, err)
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

	// Recupera la data e ora del commit attuale nel formato ISO 8601 esteso
	// Questo formato include fuso orario, utile per parsing precisi.
	// Esempio: "2006-01-02 15:04:05 +0100"
	cmdDate := exec.Command("git", "log", "-1", "--format=%ci")
	dateBytes, errDate := cmdDate.Output()
	if errDate == nil {
		date = strings.TrimSpace(string(dateBytes))
	} else {
		// Fallback o un formato predefinito se git log fallisce
		date = time.Now().Format("2006-01-02 15:04:05 -0700") // Formato standard Go per fallback
		fmt.Printf("Avviso: Impossibile ottenere la data del commit da Git (%v). Usando l'ora attuale: %s\n", errDate, date)
	}

	// Recupera l'hash del commit corrente
	cmdHash := exec.Command("git", "rev-parse", "HEAD")
	outHash, errHash := cmdHash.Output()
	if errHash != nil {
		commitHash = "Sconosciuto"
		fmt.Printf("Avviso: Impossibile ottenere l'hash del commit da Git (%v). Usando 'Sconosciuto'\n", errHash)
	} else {
		commitHash = strings.TrimSpace(string(outHash))
	}

	return author, description, date, commitHash, nil
}

func main() {
	a := app.New()
	// Altezza della finestra aumentata a 900
	w := a.NewWindow("Generatore Note di Rilascio")
	w.Resize(fyne.NewSize(700, 900))

	if len(os.Args) < 2 {
		dialog.ShowError(fmt.Errorf("Questa applicazione deve essere eseguita come un hook 'prepare-commit-msg' e necessita del percorso del file del messaggio di commit come argomento."), w)
		return
	}
	commitMsgFilePath := os.Args[1]

	modules, err := loadModules("modules.json")
	if err != nil {
		dialog.ShowError(fmt.Errorf("Errore nel caricamento dei moduli: %v", err), w)
		return
	}

	// 1. Recupero dati Git
	commitAuthor, commitDesc, commitDateRaw, commitHash, err := getGitData(commitMsgFilePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Impossibile recuperare i dati del commit attuale. Assicurati di eseguire l'applicazione in un repository Git valido. Errore: %v", err), w)
		return
	}

	// 2. Dichiarazione delle variabili per i campi del form (e puntatori per i bottoni)
	excludedFromReleaseNote := widget.NewCheck("Escludi dalla Nota di Rilascio", nil)
	tipi := []string{"Funzionalità", "Correzione Bug", "Refactoring", "Documentazione", "Generico"}
	tipo := widget.NewSelect(tipi, nil)
	modulo := widget.NewSelect(modules, nil)
	titolo := widget.NewEntry()
	descrizione := widget.NewMultiLineEntry()
	internalTicket := widget.NewEntry()
	clientTicket := widget.NewEntry()

	var saveBtn *widget.Button
	var addNoteBtn *widget.Button
	var closeBtn *widget.Button

	descrizioneLabel := widget.NewLabel("Descrizione (min. 20 caratteri):")

	excludedFromReleaseNote.OnChanged = func(checked bool) {
		if checked {
			descrizioneLabel.SetText("Descrizione:")
		} else {
			descrizioneLabel.SetText("Descrizione (min. 20 caratteri):")
			if tipo.Selected == "" {
				tipo.SetSelected("Generico")
			}
		}
		descrizioneLabel.Refresh()
	}

	// 3. Dichiarazione delle Label per i dettagli del commit
	commitAuthorLabel := widget.NewLabel(fmt.Sprintf("Autore Commit: %s", commitAuthor))
	commitHashLabel := widget.NewLabel(fmt.Sprintf("Hash Commit: %s", commitHash))
	commitDescLabel := widget.NewLabel(fmt.Sprintf("Messaggio Commit: %s", commitDesc))

	var commitDateLabel *widget.Label
	// Il formato di parsing deve corrispondere esattamente al formato di output di 'git log -1 --format=%ci'
	// Che tipicamente è "YYYY-MM-DD HH:MM:SS +ZZZZ" (es. "2006-01-02 15:04:05 +0100")
	parsedTime, errParse := time.Parse("2006-01-02 15:04:05 -0700", commitDateRaw)
	if errParse != nil {
		fmt.Printf("Errore nel parsing della data di commit per la visualizzazione: %v\n", errParse)
		commitDateLabel = widget.NewLabel(fmt.Sprintf("Data Commit: %s (Formato non riconosciuto)", commitDateRaw))
	} else {
		// Formatta per la visualizzazione italiana (DD/MM/YYYY HH:MM)
		formattedDate := parsedTime.Format("02/01/2006 15:04")
		commitDateLabel = widget.NewLabel(fmt.Sprintf("Data Commit: %s", formattedDate))
	}

	// 4. Dichiarazione delle variabili di stato
	var isAnyNoteSavedForCurrentCommit bool

	// 5. Dichiarazione e inizializzazione dell'etichetta di stato della form
	formStatusLabel := widget.NewLabel("Nuova nota: inserisci i dettagli.")
	formStatusLabel.Alignment = fyne.TextAlignCenter
	formStatusLabel.TextStyle.Bold = true

	// Funzione helper per abilitare/disabilitare i campi del form
	setFormFieldsEnabled := func(enabled bool) {
		if enabled {
			tipo.Enable()
			modulo.Enable()
			titolo.Enable()
			descrizione.Enable()
			internalTicket.Enable()
			clientTicket.Enable()
			excludedFromReleaseNote.Enable()
		} else {
			tipo.Disable()
			modulo.Disable()
			titolo.Disable()
			descrizione.Disable()
			internalTicket.Disable()
			clientTicket.Disable()
			excludedFromReleaseNote.Disable()
		}
	}

	// 6. Definizione delle funzioni helper
	resetForm := func() {
		tipo.SetSelected("Generico")
		modulo.SetSelected("")
		modulo.Refresh()
		titolo.SetText("")
		descrizione.SetText("")
		internalTicket.SetText("")
		clientTicket.SetText("")
		excludedFromReleaseNote.SetChecked(false)
		formStatusLabel.SetText("Nuova nota: inserisci i dettagli.")
		formStatusLabel.Refresh()

		setFormFieldsEnabled(true)
		saveBtn.Enable()
		if addNoteBtn != nil {
			addNoteBtn.Hide()
		}
	}

	// 7. CRITICO: Definizione della funzione 'saveNoteFunc' QUI, prima dei pulsanti
	saveNoteFunc := func() {
		isExcluded := excludedFromReleaseNote.Checked

		if !isExcluded {
			if tipo.Selected == "" {
				dialog.ShowError(fmt.Errorf("Tipo non selezionato."), w)
				return
			}
			if modulo.Selected == "" {
				dialog.ShowError(fmt.Errorf("Modulo non selezionato."), w)
				return
			}
			if len(strings.TrimSpace(descrizione.Text)) < 20 {
				dialog.ShowError(fmt.Errorf("Descrizione troppo corta (minimo 20 caratteri, spazi esclusi)."), w)
				return
			}
		} else {
			if strings.TrimSpace(descrizione.Text) == "" {
				dialog.ShowError(fmt.Errorf("La Descrizione non può essere vuota, anche se la nota è esclusa dalla Release Note."), w)
				return
			}
		}

		release := Release{
			Data:                    time.Now().Format("02/01/2006 15:04:05"), // Formato italiano completo nel JSON
			Tipo:                    tipo.Selected,
			Modulo:                  modulo.Selected,
			Titolo:                  strings.TrimSpace(titolo.Text),
			Descrizione:             strings.TrimSpace(descrizione.Text),
			InternalTicket:          strings.TrimSpace(internalTicket.Text),
			ClientTicket:            strings.TrimSpace(clientTicket.Text),
			CommitAuthor:            commitAuthor,
			CommitDesc:              commitDesc,    // Usa il messaggio commit originale (non modificabile)
			CommitDate:              commitDateRaw, // Data/ora di commit grezza da Git
			CommitHash:              commitHash,    // Inserisce l'hash del commit
			ExcludedFromReleaseNote: isExcluded,
		}

		filePath := "release_notes.json"
		var relFile ReleaseFile

		if content, err := os.ReadFile(filePath); err == nil {
			_ = json.Unmarshal(content, &relFile)
		} else if !os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf("Errore nella lettura del file release notes: %w", err), w)
			return
		}

		relFile.Releases = append(relFile.Releases, release)

		if out, err := json.MarshalIndent(relFile, "", "  "); err == nil {
			err = os.WriteFile(filePath, out, 0644)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Errore nella scrittura del file release notes: %w", err), w)
				return
			}
			isAnyNoteSavedForCurrentCommit = true // Imposta a true dopo il primo salvataggio

			// Disabilita i campi e il pulsante "Salva" dopo il salvataggio
			setFormFieldsEnabled(false)
			saveBtn.Disable()

			// Mostra il pulsante "Aggiungi Nuova Nota" solo DOPO il primo salvataggio
			if addNoteBtn != nil { // Sicurezza: controlla che il puntatore non sia nil
				addNoteBtn.Show()
			}
			formStatusLabel.SetText("Nota salvata! Clicca 'Aggiungi Nuova Nota' per inserirne un'altra o 'Chiudi'.") // Aggiorna il messaggio di stato
			formStatusLabel.Refresh()
			dialog.ShowInformation("Successo", "Release salvata correttamente!", w)
		} else {
			dialog.ShowError(fmt.Errorf("Errore nella serializzazione JSON: %v", err), w)
		}
	}

	// 8. Inizializzazione dei pulsanti (Ora saveNoteFunc è definita)
	saveBtn = widget.NewButton("Salva", func() {
		saveNoteFunc()
	})

	// Modificato il comportamento di "Aggiungi Nuova Nota"
	addNoteBtn = widget.NewButton("Aggiungi Nuova Nota", func() {
		// Il suo scopo è solo resettare la form e riabilitare i campi per una nuova nota
		resetForm()
	})
	addNoteBtn.Hide() // Nasconde il pulsante all'avvio

	closeBtn = widget.NewButton("Chiudi", func() {
		if !isAnyNoteSavedForCurrentCommit {
			dialog.ShowConfirm("Attenzione", "Nessuna nota è stata salvata per questo commit. Vuoi chiudere senza salvare e annullare il commit?", func(b bool) {
				if b {
					os.Exit(1) // Esci con codice 1 per annullare il commit
				}
			}, w)
		} else {
			a.Quit() // Esci normalmente (codice 0) se almeno una nota è stata salvata
		}
	})

	// 9. Definizione delle funzioni helper per le label
	createMandatoryLabel := func(text string) *widget.Label {
		return widget.NewLabel(text + "*")
	}
	createOptionalLabel := func(text string) *widget.Label {
		return widget.NewLabel(text + " (facoltativo):")
	}

	// 10. Creazione dei contenitori di layout dei campi
	mainFormFields := container.NewVBox(
		excludedFromReleaseNote,
		createMandatoryLabel("Tipo"), tipo,
		createMandatoryLabel("Modulo"), modulo,
		createOptionalLabel("Titolo"), titolo,
		descrizioneLabel, // Usa la label dinamica qui
		descrizione,
		createOptionalLabel("Numero Ticket Interno"), internalTicket,
		createOptionalLabel("Numero Ticket Cliente"), clientTicket,
	)

	// Raggruppamento dei campi di commit in un "card"
	commitDetailsGroup := widget.NewCard(
		"Dettagli Commit",
		"", // Sottotitolo vuoto
		container.NewVBox( // Contenitore per gli elementi del gruppo
			commitAuthorLabel,
			commitDateLabel,
			commitHashLabel, // Aggiunto l'hash del commit
			commitDescLabel, // Messaggio commit come Label (non modificabile)
		),
	)

	// Layout della form con le sezioni principali e il gruppo commit
	formLayout := container.NewVBox(
		formStatusLabel, // Etichetta di stato in cima
		mainFormFields,
		widget.NewSeparator(),
		commitDetailsGroup,
		widget.NewSeparator(),
	)

	// Wrap the form in a padded container e poi in VScroll
	contentScroll := container.NewPadded(container.NewVScroll(formLayout))

	// 11. Creazione del 'buttonContainer' (TUTTI i pulsanti devono essere inizializzati PRIMA)
	buttonContainer := container.NewGridWithColumns(3,
		saveBtn,
		addNoteBtn, // Questo sarà gestito tramite Show()/Hide()
		closeBtn,
	)

	// 12. Impostazione del contenuto della finestra 'w.SetContent()'
	w.SetContent(container.NewBorder(
		nil,             // Nessun elemento in alto
		buttonContainer, // Bottoni in basso
		nil,             // Nessun elemento a sinistra
		nil,             // Nessun elemento a destra
		contentScroll,   // Contenuto principale al centro, espandibile, con padding e scroll
	))

	// 13. Chiamate di inizializzazione finali
	// La chiamata qui assicura che lo stato iniziale della descrizioneLabel sia corretto.
	// La gestione dei campi Tipo e Modulo è ora demandata a resetForm()
	excludedFromReleaseNote.OnChanged(excludedFromReleaseNote.Checked) // Inizializza la label della descrizione e i campi tipo/modulo se excluded è true
	resetForm()                                                        // La prima chiamata a resetForm() imposta i default iniziali e abilita i campi

	w.ShowAndRun()
}
