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
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
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
	CommitHash              string `json:"commitHash"`
	ExcludedFromReleaseNote bool   `json:"excludedFromReleaseNote"`
}

// ReleaseFile è la struttura per il file JSON che contiene tutte le release.
type ReleaseFile struct {
	Releases []Release `json:"releases"`
}

// Global context per web server
type AppContext struct {
	modules                []string
	commitAuthor           string
	commitDesc             string
	commitDateRaw          string
	commitHash             string
	releasesFilePath       string
	noteSaved              bool
	modulesFilePath        string
}

// ===== FUNZIONI PURE (riutilizzabili in Fyne e Web) =====

// loadModules carica i moduli da un file JSON esterno
func loadModules(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File moduli non trovato: %s. Utilizzo moduli di default.\n", filePath)
			return []string{"Default Module"}, nil
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

// getGitData recupera i dati rilevanti per il commit attuale
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

	cmdDate := exec.Command("git", "log", "-1", "--format=%ci")
	dateBytes, errDate := cmdDate.Output()
	if errDate == nil {
		date = strings.TrimSpace(string(dateBytes))
	} else {
		date = time.Now().Format("2006-01-02 15:04:05 -0700")
	}

	cmdHash := exec.Command("git", "rev-parse", "HEAD")
	outHash, errHash := cmdHash.Output()
	if errHash != nil {
		commitHash = "Sconosciuto"
	} else {
		commitHash = strings.TrimSpace(string(outHash))
	}

	return author, description, date, commitHash, nil
}

// ValidateAndSaveRelease valida i dati e salva in release_notes.json
func ValidateAndSaveRelease(release Release, releasesFilePath string) error {
	isExcluded := release.ExcludedFromReleaseNote

	if !isExcluded {
		if release.Tipo == "" {
			return fmt.Errorf("Tipo non selezionato")
		}
		if release.Modulo == "" {
			return fmt.Errorf("Modulo non selezionato")
		}
		if len(strings.TrimSpace(release.Descrizione)) < 20 {
			return fmt.Errorf("Descrizione troppo corta (minimo 20 caratteri, spazi esclusi)")
		}
	} else {
		if strings.TrimSpace(release.Descrizione) == "" {
			return fmt.Errorf("La Descrizione non può essere vuota")
		}
	}

	var relFile ReleaseFile
	if content, err := os.ReadFile(releasesFilePath); err == nil {
		_ = json.Unmarshal(content, &relFile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("errore nella lettura del file release notes: %w", err)
	}

	relFile.Releases = append(relFile.Releases, release)

	out, err := json.MarshalIndent(relFile, "", "  ")
	if err != nil {
		return fmt.Errorf("errore nella serializzazione JSON: %w", err)
	}

	// Atomic write: scrivi in temp, poi rinomina
	tmpFile := releasesFilePath + ".tmp"
	err = os.WriteFile(tmpFile, out, 0644)
	if err != nil {
		return fmt.Errorf("errore nella scrittura temporanea: %w", err)
	}

	err = os.Rename(tmpFile, releasesFilePath)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("errore nel rinominare file: %w", err)
	}

	return nil
}

// openBrowser apre un URL nel browser di default
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch os.Getenv("GOOS") {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// fallback
		if os.Getenv("OS") == "Windows_NT" {
			cmd = exec.Command("cmd", "/c", "start", url)
		} else {
			cmd = exec.Command("xdg-open", url)
		}
	}
	return cmd.Start()
}

// ===== WEB UI EMBEDDED (HTML/CSS/JS) =====

const webUIHTML = `
<!DOCTYPE html>
<html lang="it">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Note di Rilascio</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 20px;
    }
    .container {
      background: white;
      border-radius: 12px;
      box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
      max-width: 600px;
      width: 100%;
      padding: 40px;
    }
    h1 {
      color: #333;
      margin-bottom: 10px;
      font-size: 28px;
    }
    .status {
      color: #666;
      font-size: 14px;
      margin-bottom: 30px;
    }
    .form-group {
      margin-bottom: 25px;
    }
    label {
      display: block;
      margin-bottom: 8px;
      color: #333;
      font-weight: 600;
      font-size: 14px;
    }
    .required::after {
      content: " *";
      color: #e74c3c;
    }
    input[type="text"],
    input[type="email"],
    select,
    textarea {
      width: 100%;
      padding: 12px 14px;
      border: 1px solid #e0e0e0;
      border-radius: 6px;
      font-size: 14px;
      font-family: inherit;
      transition: border-color 0.3s, box-shadow 0.3s;
    }
    input[type="text"]:focus,
    select:focus,
    textarea:focus {
      outline: none;
      border-color: #667eea;
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }
    textarea {
      resize: vertical;
      min-height: 100px;
      font-family: monospace;
      font-size: 13px;
    }
    .checkbox-group {
      display: flex;
      align-items: center;
      margin-bottom: 20px;
    }
    input[type="checkbox"] {
      width: 20px;
      height: 20px;
      cursor: pointer;
      margin-right: 10px;
    }
    .checkbox-label {
      margin: 0;
      font-weight: 500;
      color: #333;
      cursor: pointer;
    }
    .commit-details {
      background: #f8f9fa;
      border-left: 4px solid #667eea;
      padding: 15px;
      margin-bottom: 25px;
      border-radius: 4px;
      font-size: 13px;
    }
    .commit-details-item {
      margin-bottom: 10px;
    }
    .commit-details-label {
      font-weight: 600;
      color: #555;
    }
    .commit-details-value {
      color: #888;
      word-break: break-all;
      font-family: monospace;
      margin-top: 3px;
    }
    .buttons {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 15px;
      margin-top: 30px;
    }
    button {
      padding: 12px 20px;
      border: none;
      border-radius: 6px;
      font-size: 14px;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.3s;
    }
    .btn-primary {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
    }
    .btn-primary:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
    }
    .btn-secondary {
      background: #e0e0e0;
      color: #333;
    }
    .btn-secondary:hover {
      background: #d0d0d0;
    }
    .error {
      color: #e74c3c;
      font-size: 13px;
      margin-top: 6px;
    }
    .success {
      background: #d4edda;
      border: 1px solid #c3e6cb;
      color: #155724;
      padding: 15px;
      border-radius: 6px;
      margin-bottom: 20px;
    }
    .hidden {
      display: none !important;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>📝 Note di Rilascio</h1>
    <p class="status">Compila il modulo per il tuo commit</p>

    <div id="successMsg" class="success hidden">✓ Nota salvata correttamente!</div>

    <div class="commit-details">
      <div class="commit-details-item">
        <div class="commit-details-label">Autore:</div>
        <div class="commit-details-value" id="commitAuthor"></div>
      </div>
      <div class="commit-details-item">
        <div class="commit-details-label">Data:</div>
        <div class="commit-details-value" id="commitDate"></div>
      </div>
      <div class="commit-details-item">
        <div class="commit-details-label">Hash:</div>
        <div class="commit-details-value" id="commitHash"></div>
      </div>
      <div class="commit-details-item">
        <div class="commit-details-label">Messaggio:</div>
        <div class="commit-details-value" id="commitDesc"></div>
      </div>
    </div>

    <form id="releaseForm">
      <div class="checkbox-group">
        <input type="checkbox" id="excludedCheckbox">
        <label class="checkbox-label" for="excludedCheckbox">Escludi dalla Nota di Rilascio</label>
      </div>

      <div class="form-group">
        <label for="tipo" class="required">Tipo</label>
        <select id="tipo" required>
          <option value="">-- Seleziona --</option>
        </select>
        <div class="error" id="tipoError"></div>
      </div>

      <div class="form-group">
        <label for="modulo" class="required">Modulo</label>
        <select id="modulo" required>
          <option value="">-- Seleziona --</option>
        </select>
        <div class="error" id="moduloError"></div>
      </div>

      <div class="form-group">
        <label for="titolo">Titolo (facoltativo)</label>
        <input type="text" id="titolo" placeholder="Titolo breve della nota">
      </div>

      <div class="form-group">
        <label for="descrizione" id="descrizioneLabel" class="required">Descrizione (min. 20 caratteri)</label>
        <textarea id="descrizione" placeholder="Descrivi il cambiamento..."></textarea>
        <div class="error" id="descrizioneError"></div>
      </div>

      <div class="form-group">
        <label for="internalTicket">Numero Ticket Interno (facoltativo)</label>
        <input type="text" id="internalTicket" placeholder="Es: PROJ-123">
      </div>

      <div class="form-group">
        <label for="clientTicket">Numero Ticket Cliente (facoltativo)</label>
        <input type="text" id="clientTicket" placeholder="Es: CLI-456">
      </div>

      <div class="buttons">
        <button type="submit" class="btn-primary">Salva Nota</button>
        <button type="button" class="btn-secondary" onclick="window.close()">Chiudi</button>
      </div>
    </form>
  </div>

  <script>
    const tipiOptions = ["Funzionalità", "Correzione Bug", "Refactoring", "Documentazione", "Generico"];
    let moduliOptions = [];
    let gitData = {};

    // Fetch initial data
    fetch('/api/init')
      .then(r => r.json())
      .then(data => {
        gitData = data;
        moduliOptions = data.modules;

        // Populate selects
        document.getElementById('commitAuthor').textContent = data.commitAuthor;
        document.getElementById('commitDate').textContent = data.commitDate;
        document.getElementById('commitHash').textContent = data.commitHash;
        document.getElementById('commitDesc').textContent = data.commitDesc;

        const tipoSel = document.getElementById('tipo');
        tipiOptions.forEach(t => {
          const opt = document.createElement('option');
          opt.value = t;
          opt.textContent = t;
          tipoSel.appendChild(opt);
        });
        tipoSel.value = "Generico";

        const moduloSel = document.getElementById('modulo');
        moduliOptions.forEach(m => {
          const opt = document.createElement('option');
          opt.value = m;
          opt.textContent = m;
          moduloSel.appendChild(opt);
        });
      })
      .catch(err => console.error('Init error:', err));

    // Toggle excluded
    document.getElementById('excludedCheckbox').addEventListener('change', function() {
      const label = document.getElementById('descrizioneLabel');
      const tipoSel = document.getElementById('tipo');
      if (this.checked) {
        label.textContent = "Descrizione";
        label.classList.remove('required');
      } else {
        label.textContent = "Descrizione (min. 20 caratteri)";
        label.classList.add('required');
        if (tipoSel.value === '') {
          tipoSel.value = "Generico";
        }
      }
    });

    // Form submit
    document.getElementById('releaseForm').addEventListener('submit', function(e) {
      e.preventDefault();

      // Clear errors
      document.querySelectorAll('.error').forEach(el => el.textContent = '');

      const isExcluded = document.getElementById('excludedCheckbox').checked;
      const tipo = document.getElementById('tipo').value;
      const modulo = document.getElementById('modulo').value;
      const descrizione = document.getElementById('descrizione').value.trim();

      if (!isExcluded) {
        let hasError = false;
        if (!tipo) {
          document.getElementById('tipoError').textContent = 'Tipo non selezionato';
          hasError = true;
        }
        if (!modulo) {
          document.getElementById('moduloError').textContent = 'Modulo non selezionato';
          hasError = true;
        }
        if (descrizione.length < 20) {
          document.getElementById('descrizioneError').textContent = 'Minimo 20 caratteri';
          hasError = true;
        }
        if (hasError) return;
      } else {
        if (!descrizione) {
          document.getElementById('descrizioneError').textContent = 'Non può essere vuota';
          return;
        }
      }

      const payload = {
        tipo,
        modulo,
        titolo: document.getElementById('titolo').value.trim(),
        descrizione,
        internalTicket: document.getElementById('internalTicket').value.trim(),
        clientTicket: document.getElementById('clientTicket').value.trim(),
        excludedFromReleaseNote: isExcluded
      };

      fetch('/api/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      })
        .then(r => r.json())
        .then(data => {
          if (data.success) {
            document.getElementById('successMsg').classList.remove('hidden');
            document.getElementById('releaseForm').reset();
            // Auto-close dopo 2 secondi
            setTimeout(() => window.close(), 2000);
          } else {
            alert('Errore: ' + data.error);
          }
        })
        .catch(err => alert('Errore di salvataggio: ' + err.message));
    });
  </script>
</body>
</html>
`

// ===== HTTP HANDLERS PER WEB UI =====

func handleInit(ctx *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Parse commit date for display
		parsedTime, _ := time.Parse("2006-01-02 15:04:05 -0700", ctx.commitDateRaw)
		displayDate := parsedTime.Format("02/01/2006 15:04")
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"modules":      ctx.modules,
			"commitAuthor": ctx.commitAuthor,
			"commitDate":   displayDate,
			"commitHash":   ctx.commitHash,
			"commitDesc":   ctx.commitDesc,
		})
	}
}

func handleSave(ctx *AppContext, done chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var payload struct {
			Tipo                    string `json:"tipo"`
			Modulo                  string `json:"modulo"`
			Titolo                  string `json:"titolo"`
			Descrizione             string `json:"descrizione"`
			InternalTicket          string `json:"internalTicket"`
			ClientTicket            string `json:"clientTicket"`
			ExcludedFromReleaseNote bool   `json:"excludedFromReleaseNote"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request",
			})
			return
		}

		release := Release{
			Data:                    time.Now().Format("02/01/2006 15:04:05"),
			Tipo:                    payload.Tipo,
			Modulo:                  payload.Modulo,
			Titolo:                  payload.Titolo,
			Descrizione:             payload.Descrizione,
			InternalTicket:          payload.InternalTicket,
			ClientTicket:            payload.ClientTicket,
			CommitAuthor:            ctx.commitAuthor,
			CommitDesc:              ctx.commitDesc,
			CommitDate:              ctx.commitDateRaw,
			CommitHash:              ctx.commitHash,
			ExcludedFromReleaseNote: payload.ExcludedFromReleaseNote,
		}

		if err := ValidateAndSaveRelease(release, ctx.releasesFilePath); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		ctx.noteSaved = true
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

		// Signal completion
		select {
		case done <- struct{}{}:
		default:
		}
	}
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(webUIHTML))
}

// ===== MODE: WEB SERVER =====

func runWebMode(commitMsgFile, releasesFile string) error {
	commitAuthor, commitDesc, commitDateRaw, commitHash, err := getGitData(commitMsgFile)
	if err != nil {
		return fmt.Errorf("impossibile recuperare dati Git: %w", err)
	}

	modules, err := loadModules("modules.json")
	if err != nil {
		return fmt.Errorf("impossibile caricare moduli: %w", err)
	}

	ctx := &AppContext{
		modules:          modules,
		commitAuthor:     commitAuthor,
		commitDesc:       commitDesc,
		commitDateRaw:    commitDateRaw,
		commitHash:       commitHash,
		releasesFilePath: releasesFile,
		noteSaved:        false,
		modulesFilePath:  "modules.json",
	}

	done := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleUI)
	mux.HandleFunc("/api/init", handleInit(ctx))
	mux.HandleFunc("/api/save", handleSave(ctx, done))

	server := &http.Server{
		Addr:    "127.0.0.1:9999",
		Handler: mux,
	}

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	// Open browser
	openBrowser("http://127.0.0.1:9999")

	// Wait for save or signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		server.Close()
		return nil
	case <-sigChan:
		server.Close()
		if ctx.noteSaved {
			return nil
		}
		return fmt.Errorf("nota non salvata")
	}
}

// ===== MODE: HEADLESS (CLI) =====

func runHeadlessMode(commitMsgFile, releasesFile, tipoVal, moduloVal, descrizioneVal string, excludedVal bool) error {
	commitAuthor, commitDesc, commitDateRaw, commitHash, err := getGitData(commitMsgFile)
	if err != nil {
		return fmt.Errorf("impossibile recuperare dati Git: %w", err)
	}

	release := Release{
		Data:                    time.Now().Format("02/01/2006 15:04:05"),
		Tipo:                    tipoVal,
		Modulo:                  moduloVal,
		Descrizione:             descrizioneVal,
		CommitAuthor:            commitAuthor,
		CommitDesc:              commitDesc,
		CommitDate:              commitDateRaw,
		CommitHash:              commitHash,
		ExcludedFromReleaseNote: excludedVal,
	}

	return ValidateAndSaveRelease(release, releasesFile)
}

// ===== MODE: FYNE GUI (legacy) =====

func runFyneMode(commitMsgFilePath string) error {
	a := app.New()
	w := a.NewWindow("Generatore Note di Rilascio")
	w.Resize(fyne.NewSize(700, 900))

	modules, err := loadModules("modules.json")
	if err != nil {
		dialog.ShowError(fmt.Errorf("Errore nel caricamento dei moduli: %v", err), w)
		return err
	}

	commitAuthor, commitDesc, commitDateRaw, commitHash, err := getGitData(commitMsgFilePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Impossibile recuperare i dati del commit: %v", err), w)
		return err
	}

	// Form elements
	excludedFromReleaseNote := widget.NewCheck("Escludi dalla Nota di Rilascio", nil)
	tipi := []string{"Funzionalità", "Correzione Bug", "Refactoring", "Documentazione", "Generico"}
	tipo := widget.NewSelect(tipi, nil)
	tipo.SetSelected("Generico")
	modulo := widget.NewSelect(modules, nil)
	titolo := widget.NewEntry()
	descrizione := widget.NewMultiLineEntry()
	internalTicket := widget.NewEntry()
	clientTicket := widget.NewEntry()

	var saveBtn *widget.Button
	var addNoteBtn *widget.Button
	var closeBtn *widget.Button

	descrizioneLabel := widget.NewLabel("Descrizione (min. 20 caratteri):")

	commitAuthorLabel := widget.NewLabel(fmt.Sprintf("Autore Commit: %s", commitAuthor))
	commitHashLabel := widget.NewLabel(fmt.Sprintf("Hash Commit: %s", commitHash))
	commitDescLabel := widget.NewLabel(fmt.Sprintf("Messaggio Commit: %s", commitDesc))

	var commitDateLabel *widget.Label
	parsedTime, errParse := time.Parse("2006-01-02 15:04:05 -0700", commitDateRaw)
	if errParse != nil {
		commitDateLabel = widget.NewLabel(fmt.Sprintf("Data Commit: %s (Formato non riconosciuto)", commitDateRaw))
	} else {
		formattedDate := parsedTime.Format("02/01/2006 15:04")
		commitDateLabel = widget.NewLabel(fmt.Sprintf("Data Commit: %s", formattedDate))
	}

	var isAnyNoteSavedForCurrentCommit bool

	formStatusLabel := widget.NewLabel("Nuova nota: inserisci i dettagli.")
	formStatusLabel.Alignment = fyne.TextAlignCenter
	formStatusLabel.TextStyle.Bold = true

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

	saveNoteFunc := func() {
		release := Release{
			Data:                    time.Now().Format("02/01/2006 15:04:05"),
			Tipo:                    tipo.Selected,
			Modulo:                  modulo.Selected,
			Titolo:                  strings.TrimSpace(titolo.Text),
			Descrizione:             strings.TrimSpace(descrizione.Text),
			InternalTicket:          strings.TrimSpace(internalTicket.Text),
			ClientTicket:            strings.TrimSpace(clientTicket.Text),
			CommitAuthor:            commitAuthor,
			CommitDesc:              commitDesc,
			CommitDate:              commitDateRaw,
			CommitHash:              commitHash,
			ExcludedFromReleaseNote: excludedFromReleaseNote.Checked,
		}

		if err := ValidateAndSaveRelease(release, "release_notes.json"); err != nil {
			dialog.ShowError(err, w)
			return
		}

		isAnyNoteSavedForCurrentCommit = true
		setFormFieldsEnabled(false)
		saveBtn.Disable()
		if addNoteBtn != nil {
			addNoteBtn.Show()
		}
		formStatusLabel.SetText("Nota salvata! Clicca 'Aggiungi Nuova Nota' per inserirne un'altra o 'Chiudi'.")
		formStatusLabel.Refresh()
		dialog.ShowInformation("Successo", "Release salvata correttamente!", w)
	}

	saveBtn = widget.NewButton("Salva", func() {
		saveNoteFunc()
	})

	addNoteBtn = widget.NewButton("Aggiungi Nuova Nota", func() {
		resetForm()
	})
	addNoteBtn.Hide()

	closeBtn = widget.NewButton("Chiudi", func() {
		if !isAnyNoteSavedForCurrentCommit {
			dialog.ShowConfirm("Attenzione", "Nessuna nota è stata salvata per questo commit. Vuoi chiudere senza salvare?", func(b bool) {
				if b {
					os.Exit(1)
				}
			}, w)
		} else {
			a.Quit()
		}
	})

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

	mainFormFields := container.NewVBox(
		excludedFromReleaseNote,
		widget.NewLabel("Tipo*"), tipo,
		widget.NewLabel("Modulo*"), modulo,
		widget.NewLabel("Titolo (facoltativo):"), titolo,
		descrizioneLabel,
		descrizione,
		widget.NewLabel("Numero Ticket Interno (facoltativo):"), internalTicket,
		widget.NewLabel("Numero Ticket Cliente (facoltativo):"), clientTicket,
	)

	commitDetailsGroup := widget.NewCard(
		"Dettagli Commit",
		"",
		container.NewVBox(
			commitAuthorLabel,
			commitDateLabel,
			commitHashLabel,
			commitDescLabel,
		),
	)

	formLayout := container.NewVBox(
		formStatusLabel,
		mainFormFields,
		widget.NewSeparator(),
		commitDetailsGroup,
		widget.NewSeparator(),
	)

	contentScroll := container.NewPadded(container.NewVScroll(formLayout))

	buttonContainer := container.NewGridWithColumns(3,
		saveBtn,
		addNoteBtn,
		closeBtn,
	)

	w.SetContent(container.NewBorder(
		nil,
		buttonContainer,
		nil,
		nil,
		contentScroll,
	))

	excludedFromReleaseNote.OnChanged(excludedFromReleaseNote.Checked)
	resetForm()

	w.ShowAndRun()
	return nil
}

// ===== MAIN =====

func main() {
	// Flags
	servePtr := flag.String("serve", "", "Start web UI server on port (e.g., ':9999')")
	commitMsgPtr := flag.String("commit-msg", "", "Path to commit message file")
	releasesPtr := flag.String("json-out", "release_notes.json", "Path to output release_notes.json")
	headlessPtr := flag.Bool("headless", false, "Run in headless mode (requires other flags)")
	tipoPtr := flag.String("tipo", "", "Release type (headless mode)")
	moduloPtr := flag.String("modulo", "", "Module name (headless mode)")
	descrizionePtr := flag.String("descrizione", "", "Description (headless mode)")
	excludedPtr := flag.Bool("excluded", false, "Exclude from release notes (headless mode)")

	flag.Parse()

	// Legacy mode: no flags provided
	if *commitMsgPtr == "" && *servePtr == "" && !*headlessPtr {
		if len(os.Args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage:\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
		commitMsgFile := os.Args[1]
		if err := runFyneMode(commitMsgFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Web UI mode
	if *servePtr != "" && *commitMsgPtr != "" {
		if err := runWebMode(*commitMsgPtr, *releasesPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Headless mode
	if *headlessPtr && *commitMsgPtr != "" {
		if err := runHeadlessMode(*commitMsgPtr, *releasesPtr, *tipoPtr, *moduloPtr, *descrizionePtr, *excludedPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "Invalid combination of flags\n")
	flag.PrintDefaults()
	os.Exit(1)
}
