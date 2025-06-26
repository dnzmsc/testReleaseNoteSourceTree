package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"
)

type Release struct {
	Data        string `json:"data"`
	Tipo        string `json:"tipo"`
	Titolo      string `json:"titolo"`
	Descrizione string `json:"descrizione"`
	Autore      string `json:"autore"`
	PR          string `json:"pr"`
	Changelog   string `json:"changelog"`
}

type ReleaseFile struct {
	Releases []Release `json:"releases"`
}

func main() {
	a := app.New()
	w := a.NewWindow("Release Notes")
	w.Resize(fyne.NewSize(500, 600))

	// Campi del form
	tipi := []string{"Feature", "Fix", "Refactor"}
	tipo := widget.NewSelect(tipi, nil)
	titolo := widget.NewEntry()
	descrizione := widget.NewMultiLineEntry()
	autore := widget.NewEntry()
	pr := widget.NewEntry()
	changelog := widget.NewMultiLineEntry()

	// Bottone
	saveBtn := widget.NewButton("Salva", func() {
		// Validazioni
		if tipo.Selected == "" {
			dialog.ShowError(fmt.Errorf("Tipo non selezionato"), w)
			return
		}
		if len(titolo.Text) < 3 {
			dialog.ShowError(fmt.Errorf("Titolo troppo corto"), w)
			return
		}
		if len(descrizione.Text) < 10 {
			dialog.ShowError(fmt.Errorf("Descrizione troppo corta"), w)
			return
		}
		if !strings.HasPrefix(pr.Text, "PR") || len(pr.Text) < 3 {
			dialog.ShowError(fmt.Errorf("PR deve iniziare con 'PR'"), w)
			return
		}

		// Crea la release
		release := Release{
			Data:        time.Now().Format("2006-01-02"),
			Tipo:        tipo.Selected,
			Titolo:      strings.TrimSpace(titolo.Text),
			Descrizione: strings.TrimSpace(descrizione.Text),
			Autore:      strings.TrimSpace(autore.Text),
			PR:          strings.TrimSpace(pr.Text),
			Changelog:   strings.TrimSpace(changelog.Text),
		}

		filePath := "release_notes.json"
		var relFile ReleaseFile

		// Carica se esiste
		if content, err := os.ReadFile(filePath); err == nil {
			_ = json.Unmarshal(content, &relFile)
		}

		// Aggiunge nuova nota
		relFile.Releases = append(relFile.Releases, release)

		// Salva
		if out, err := json.MarshalIndent(relFile, "", "  "); err == nil {
			os.WriteFile(filePath, out, 0644)
			dialog.ShowInformation("Successo", "Release salvata correttamente!", w)
			a.Quit()
		} else {
			dialog.ShowError(fmt.Errorf("Errore nel salvataggio: %v", err), w)
		}
	})

	// Layout
	form := container.NewVBox(
		widget.NewLabel("Tipo:"), tipo,
		widget.NewLabel("Titolo:"), titolo,
		widget.NewLabel("Descrizione:"), descrizione,
		widget.NewLabel("Autore:"), autore,
		widget.NewLabel("PR (es: PR1234):"), pr,
		widget.NewLabel("Changelog:"), changelog,
		saveBtn,
	)

	w.SetContent(form)
	w.ShowAndRun()
}
