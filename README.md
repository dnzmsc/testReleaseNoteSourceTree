# Release Notes Tool

Tool per raccogliere note di rilascio ad ogni commit Git tramite una finestra GUI nativa e bloccante.

## Come funziona

1. Lo sviluppatore fa un commit (da terminale, SourceTree, VS Code, ecc.)
2. Il git hook `prepare-commit-msg` lancia il tool
3. Si apre una finestra nativa con il form delle note di rilascio
4. Lo sviluppatore compila il form e clicca "Salva e Committa"
5. Il file `release_notes.json` viene aggiornato e aggiunto automaticamente al commit
6. Il commit procede normalmente

Se lo sviluppatore chiude la finestra o clicca "Annulla", il commit viene annullato.

## Requisiti

- Il binario `release-notes` deve essere nel PATH di sistema (installato dai sistemisti)
- Go 1.24+ per compilare da sorgente

## Setup per repository

Ogni repository che vuole usare le release notes deve eseguire una volta:

```bash
/path/to/setup-release-notes.sh
```

Questo installa l'hook `prepare-commit-msg` nel repository e crea i file `release_notes.json` e `modules.json` se mancanti.

## Configurazione

### modules.json

Ogni repository ha il suo `modules.json` nella root con la lista dei moduli del progetto:

```json
["CORE", "API", "FRONTEND", "DATABASE"]
```

Se il file non esiste, viene usato un modulo "Default".

### release_notes.json

File di output nella root del repository. Viene creato automaticamente e aggiornato ad ogni commit. Struttura:

```json
{
  "releases": [
    {
      "data": "20/04/2026 14:30:00",
      "tipo": "Correzione Bug",
      "modulo": "CORE",
      "titolo": "Fix login timeout",
      "descrizione": "Corretto il timeout della sessione di login...",
      "internalTicket": "PROJ-123",
      "clientTicket": "CLI-456",
      "commitAuthor": "mario.rossi",
      "commitDesc": "fix: login timeout issue",
      "commitDate": "2026-04-20 14:30:00",
      "commitHash": "PENDING",
      "excludedFromReleaseNote": false
    }
  ]
}
```

## Build

### Build per la piattaforma corrente

```bash
CGO_ENABLED=1 go build -o release-notes .
```

### Build con versione

```bash
./scripts/build-all.sh 1.0.0
```

Nota: Fyne richiede CGO, quindi la cross-compilation necessita di un cross-compiler C per ogni piattaforma target. Per build multi-piattaforma, compilare nativamente su ogni OS.

## Installazione (per sistemisti)

1. Compilare il binario sulla piattaforma target
2. Copiare il binario nel PATH di sistema:
   - macOS/Linux: `/usr/local/bin/release-notes`
   - Windows: `C:\Program Files\Git\cmd\release-notes.exe`
3. Per ogni repository, eseguire `setup-release-notes.sh`

## Struttura del progetto

```
├── main.go                          # Sorgente Go (GUI Fyne + logica)
├── go.mod / go.sum                  # Dipendenze Go
├── modules.json                     # Configurazione moduli (per-repository)
├── release_notes.json               # Output note di rilascio (per-repository)
├── .git-hooks/
│   └── prepare-commit-msg           # Hook Git che lancia il tool
├── setup-release-notes.sh           # Script setup per-repository
├── scripts/
│   └── build-all.sh                 # Script di build
└── test-suite.sh                    # Test suite
```

## Troubleshooting

### "release-notes: command not found"
Il binario non è nel PATH. Verificare con `which release-notes`.

### La finestra non si apre
Verificare che il sistema abbia un display grafico disponibile. Su server headless il tool non può funzionare.

### Commit bloccato
Se il tool crasha, il commit viene annullato. Per bypassare temporaneamente:
```bash
git commit --no-verify -m "messaggio"
```
