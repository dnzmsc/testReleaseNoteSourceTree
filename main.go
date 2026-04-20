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

// Version is set at build time via ldflags
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
	// Usage: release-notes --commit-msg <path> [--output <path>] [--modules <path>]
	// Typically invoked by the prepare-commit-msg git hook.

	commitMsgFile := ""
	outputFile := "release_notes.json"
	modulesFile := "modules.json"

	// Simple flag parsing (no flag package to keep control over exit codes)
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

	// Load modules configuration
	modules, err := loadModules(modulesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errore caricamento moduli: %v\n", err)
		os.Exit(1)
	}

	// Get git data for this commit
	author, commitDesc, commitDate := getGitInfo(commitMsgFile)

	// Show the GUI and get the result
	saved := showReleaseForm(modules, author, commitDesc, commitDate, outputFile)

	if saved {
		// Stage the release_notes.json so it's included in this commit
		stageFile(outputFile)
		os.Exit(0)
	}

	// User cancelled or closed the window → abort the commit
	os.Exit(1)
}

// showReleaseForm displays the Fyne GUI and returns true if the user saved a note.
func showReleaseForm(modules []string, author, commitDesc, commitDate, outputFile string) bool {
	saved := false

	a := app.New()
	w := a.NewWindow("📝 Release Notes")
	w.Resize(fyne.NewSize(620, 720))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	// Closing the window without saving = abort
	w.SetCloseIntercept(func() {
		w.Close()
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

	// --- Error label ---
	errorLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// --- Buttons ---
	saveBtn := widget.NewButtonWithIcon("Salva e Committa", theme.DocumentSaveIcon(), func() {
		isExcluded := excludedCheck.Checked
		description := strings.TrimSpace(descEntry.Text)

		// Validation
		if errMsg := validateForm(isExcluded, tipoSelect.Selected, moduloSelect.Selected, description); errMsg != "" {
			errorLabel.SetText("❌ " + errMsg)
			return
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
			return
		}

		saved = true
		w.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Annulla Commit", theme.CancelIcon(), func() {
		w.Close()
	})

	// --- Layout ---
	form := container.NewVBox(
		gitInfoCard,
		widget.NewSeparator(),
		excludedCheck,
		widget.NewLabel("Tipo *"), tipoSelect,
		widget.NewLabel("Modulo *"), moduloSelect,
		widget.NewLabel("Titolo"), titoloEntry,
		widget.NewLabel("Descrizione / Motivo Esclusione *"), descEntry,
		widget.NewLabel("Internal Ticket"), internalTicketEntry,
		widget.NewLabel("Client Ticket"), clientTicketEntry,
		errorLabel,
		layout.NewSpacer(),
		container.NewGridWithColumns(2, cancelBtn, saveBtn),
	)

	scrollable := container.NewVScroll(form)

	content := container.NewPadded(container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		scrollable,
	))

	w.SetContent(content)
	w.ShowAndRun()

	return saved
}

// validateForm checks required fields and returns an error message, or "" if valid.
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

	// Atomic write: write to temp file, then rename
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, out, 0644); err != nil {
		return fmt.Errorf("errore scrittura file temporaneo: %w", err)
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		// Cleanup temp file on rename failure
		os.Remove(tmpFile)
		return fmt.Errorf("errore rinomina file: %w", err)
	}

	return nil
}

// loadModules loads the module list from a JSON file.
// Returns a default list if the file doesn't exist.
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

// getGitInfo extracts author, commit message, and date for the current commit.
func getGitInfo(commitMsgFile string) (author, commitDesc, commitDate string) {
	// Author from git config
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		author = strings.TrimSpace(string(out))
	} else {
		author = "Sconosciuto"
	}

	// Commit message from the file git passes to the hook
	commitDesc = ""
	if commitMsgFile != "" {
		if content, err := os.ReadFile(commitMsgFile); err == nil {
			var lines []string
			for _, line := range strings.Split(string(content), "\n") {
				// Filter out git comment lines
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					lines = append(lines, line)
				}
			}
			commitDesc = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	// Current timestamp (the commit hasn't been created yet at hook time)
	commitDate = time.Now().Format("2006-01-02 15:04:05")

	return
}

// stageFile runs "git add <file>" to include the file in the current commit.
func stageFile(filePath string) {
	cmd := exec.Command("git", "add", filePath)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Attenzione: impossibile aggiungere %s al commit: %v\n", filePath, err)
	}
}

// truncate shortens a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
