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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Release represents a single release note entry.
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
	CommitHash              string `json:"commitHash"`
	ExcludedFromReleaseNote bool   `json:"excludedFromReleaseNote"`
}

// ReleaseFile is the JSON structure containing all release notes.
type ReleaseFile struct {
	Releases []Release `json:"releases"`
}

func main() {
	commitMsgFile := ""
	outputFile := "release_notes.json"
	modulesFile := "modules.json"

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--commit-msg":
			if i+1 < len(args) {
				commitMsgFile = args[i+1]
				i++
			}
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--modules":
			if i+1 < len(args) {
				modulesFile = args[i+1]
				i++
			}
		case "--version":
			fmt.Printf("release-notes %s\n", Version)
			os.Exit(0)
		}
	}

	modules, err := loadModules(modulesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errore caricamento moduli: %v\n", err)
		os.Exit(1)
	}

	author, commitDesc, commitDate := getGitInfo(commitMsgFile)

	exitCode := showReleaseForm(modules, author, commitDesc, commitDate, outputFile)
	os.Exit(exitCode)
}

// showReleaseForm displays the GUI. Returns 0 if at least one note was saved, 1 otherwise.
func showReleaseForm(modules []string, author, commitDesc, commitDate, outputFile string) int {
	noteCount := 0

	a := app.New()
	w := a.NewWindow("📝 Release Notes")
	w.Resize(fyne.NewSize(620, 750))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	// Closing the window without having saved at least one note = abort
	w.SetCloseIntercept(func() {
		if noteCount == 0 {
			// No notes saved yet — confirm abort
			dialog.ShowConfirm(
				"Annullare il commit?",
				"Non hai salvato nessuna nota di rilascio.\nIl commit verrà annullato.",
				func(confirmed bool) {
					if confirmed {
						w.Close()
					}
				},
				w,
			)
		} else {
			w.Close()
		}
	})

	// --- Header ---
	header := widget.NewLabelWithStyle(
		"Nuova Nota di Rilascio",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// --- Git info card ---
	gitInfoCard := widget.NewCard("Dettagli Commit", "", container.NewVBox(
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**Autore:** %s", author)),
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**Data:** %s", commitDate)),
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**Messaggio:** %s", truncate(commitDesc, 80))),
	))

	// --- Status label (shows how many notes saved so far) ---
	statusLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// --- Form fields ---
	excludedCheck := widget.NewCheck("Escludi dalla Nota di Rilascio", nil)

	tipiOptions := []string{"Funzionalità", "Correzione Bug", "Refactoring", "Documentazione", "Generico"}
	tipoSelect := widget.NewSelect(tipiOptions, nil)
	tipoSelect.PlaceHolder = "Seleziona Tipo..."

	moduloSelect := widget.NewSelect(modules, nil)
	moduloSelect.PlaceHolder = "Seleziona Modulo..."

	titoloEntry := widget.NewEntry()
	titoloEntry.PlaceHolder = "Titolo breve (facoltativo)"

	descEntry := widget.NewMultiLineEntry()
	descEntry.PlaceHolder = "Descrizione dettagliata (min. 20 caratteri)..."
	descEntry.SetMinRowsVisible(4)

	internalTicketEntry := widget.NewEntry()
	internalTicketEntry.PlaceHolder = "Es: PROJ-123"

	clientTicketEntry := widget.NewEntry()
	clientTicketEntry.PlaceHolder = "Es: CLI-456"

	errorLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// --- Reset form helper ---
	resetForm := func() {
		excludedCheck.SetChecked(false)
		tipoSelect.ClearSelected()
		moduloSelect.ClearSelected()
		titoloEntry.SetText("")
		descEntry.SetText("")
		internalTicketEntry.SetText("")
		clientTicketEntry.SetText("")
		errorLabel.SetText("")
		tipoSelect.Enable()
		moduloSelect.Enable()
		titoloEntry.Enable()
		internalTicketEntry.Enable()
		clientTicketEntry.Enable()
		descEntry.PlaceHolder = "Descrizione dettagliata (min. 20 caratteri)..."
	}

	// --- Excluded checkbox logic ---
	excludedCheck.OnChanged = func(checked bool) {
		if checked {
			tipoSelect.Disable()
			moduloSelect.Disable()
			titoloEntry.Disable()
			internalTicketEntry.Disable()
			clientTicketEntry.Disable()
			descEntry.PlaceHolder = "Motivo dell'esclusione..."
		} else {
			tipoSelect.Enable()
			moduloSelect.Enable()
			titoloEntry.Enable()
			internalTicketEntry.Enable()
			clientTicketEntry.Enable()
			descEntry.PlaceHolder = "Descrizione dettagliata (min. 20 caratteri)..."
		}
	}

	// --- Buttons ---
	// Declared here so saveAndAdd can reference finishBtn and vice versa
	var saveAndCommitBtn *widget.Button
	var saveAndAddBtn *widget.Button

	saveNote := func() bool {
		isExcluded := excludedCheck.Checked
		description := strings.TrimSpace(descEntry.Text)

		if errMsg := validateForm(isExcluded, tipoSelect.Selected, moduloSelect.Selected, description); errMsg != "" {
			errorLabel.SetText("❌ " + errMsg)
			return false
		}

		release := Release{
			Data:                    time.Now().Format("02/01/2006 15:04:05"),
			Tipo:                    tipoSelect.Selected,
			Modulo:                  moduloSelect.Selected,
			Titolo:                  strings.TrimSpace(titoloEntry.Text),
			Descrizione:             description,
			InternalTicket:          strings.TrimSpace(internalTicketEntry.Text),
			ClientTicket:            strings.TrimSpace(clientTicketEntry.Text),
			CommitAuthor:            author,
			CommitDesc:              commitDesc,
			CommitDate:              commitDate,
			CommitHash:              "PENDING",
			ExcludedFromReleaseNote: isExcluded,
		}

		if err := saveRelease(release, outputFile); err != nil {
			dialog.ShowError(fmt.Errorf("Errore salvataggio: %v", err), w)
			return false
		}

		noteCount++
		statusLabel.SetText(fmt.Sprintf("✅ %d nota/e salvata/e per questo commit", noteCount))
		return true
	}

	// "Save and add another note" button
	saveAndAddBtn = widget.NewButtonWithIcon("Salva e Aggiungi Altra Nota", theme.ContentAddIcon(), func() {
		if saveNote() {
			resetForm()
		}
	})

	// "Save and commit" button (saves current note + closes)
	saveAndCommitBtn = widget.NewButtonWithIcon("Salva e Committa", theme.DocumentSaveIcon(), func() {
		if saveNote() {
			w.Close()
		}
	})
	saveAndCommitBtn.Importance = widget.HighImportance

	// "Finish" button — only visible after at least one note is saved
	finishBtn := widget.NewButtonWithIcon("Completa Commit", theme.ConfirmIcon(), func() {
		w.Close()
	})
	finishBtn.Importance = widget.HighImportance
	finishBtn.Hide()

	cancelBtn := widget.NewButtonWithIcon("Annulla Commit", theme.CancelIcon(), func() {
		if noteCount == 0 {
			w.Close()
		} else {
			dialog.ShowConfirm(
				"Annullare il commit?",
				fmt.Sprintf("Hai già salvato %d nota/e.\nSe annulli, le note resteranno nel file ma il commit non verrà eseguito.", noteCount),
				func(confirmed bool) {
					if confirmed {
						noteCount = 0 // Reset so exit code is 1
						w.Close()
					}
				},
				w,
			)
		}
	})

	// --- Layout ---
	buttonRow := container.NewGridWithColumns(2, cancelBtn, saveAndCommitBtn)
	addAnotherRow := container.NewGridWithColumns(2, saveAndAddBtn, finishBtn)

	form := container.NewVBox(
		gitInfoCard,
		widget.NewSeparator(),
		statusLabel,
		excludedCheck,
		widget.NewLabel("Tipo *"), tipoSelect,
		widget.NewLabel("Modulo *"), moduloSelect,
		widget.NewLabel("Titolo"), titoloEntry,
		widget.NewLabel("Descrizione / Motivo Esclusione *"), descEntry,
		widget.NewLabel("Internal Ticket"), internalTicketEntry,
		widget.NewLabel("Client Ticket"), clientTicketEntry,
		errorLabel,
		layout.NewSpacer(),
		addAnotherRow,
		buttonRow,
	)

	scrollable := container.NewVScroll(form)

	content := container.NewPadded(container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		scrollable,
	))

	w.SetContent(content)

	// Use a goroutine-safe way to track note count for button visibility.
	// We wrap saveNote to also update button visibility.
	origSaveNote := saveNote
	saveNote = func() bool {
		result := origSaveNote()
		if result && noteCount > 0 {
			finishBtn.Show()
		}
		return result
	}
	// Re-bind buttons with the wrapped saveNote
	saveAndAddBtn.OnTapped = func() {
		if saveNote() {
			resetForm()
		}
	}
	saveAndCommitBtn.OnTapped = func() {
		if saveNote() {
			w.Close()
		}
	}

	w.ShowAndRun()

	if noteCount > 0 {
		return 0
	}
	return 1
}

// validateForm checks required fields. Returns error message or "" if valid.
func validateForm(isExcluded bool, tipo, modulo, description string) string {
	if isExcluded {
		if len(description) == 0 {
			return "Inserisci il motivo dell'esclusione nella descrizione"
		}
		return ""
	}
	if tipo == "" {
		return "Seleziona un Tipo"
	}
	if modulo == "" {
		return "Seleziona un Modulo"
	}
	if len(description) < 20 {
		return "La descrizione deve essere di almeno 20 caratteri"
	}
	return ""
}

// saveRelease appends a release entry to the JSON file using atomic write.
func saveRelease(release Release, filePath string) error {
	var relFile ReleaseFile

	if content, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(content, &relFile)
	}

	relFile.Releases = append(relFile.Releases, release)

	out, err := json.MarshalIndent(relFile, "", "  ")
	if err != nil {
		return fmt.Errorf("errore serializzazione JSON: %w", err)
	}

	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, out, 0644); err != nil {
		return fmt.Errorf("errore scrittura file temporaneo: %w", err)
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("errore rinomina file: %w", err)
	}

	return nil
}

// loadModules loads the module list from a JSON file.
func loadModules(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"Default"}, nil
		}
		return nil, fmt.Errorf("errore lettura file moduli: %w", err)
	}

	var modules []string
	if err := json.Unmarshal(content, &modules); err != nil {
		return nil, fmt.Errorf("errore parsing file moduli: %w", err)
	}

	if len(modules) == 0 {
		return []string{"Default"}, nil
	}

	return modules, nil
}

// getGitInfo extracts author, commit message, and date.
func getGitInfo(commitMsgFile string) (author, commitDesc, commitDate string) {
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		author = strings.TrimSpace(string(out))
	} else {
		author = "Sconosciuto"
	}

	commitDesc = ""
	if commitMsgFile != "" {
		if content, err := os.ReadFile(commitMsgFile); err == nil {
			var lines []string
			for _, line := range strings.Split(string(content), "\n") {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					lines = append(lines, line)
				}
			}
			commitDesc = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	commitDate = time.Now().Format("2006-01-02 15:04:05")
	return
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
