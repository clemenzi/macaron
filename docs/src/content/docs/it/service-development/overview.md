---
title: Struttura di un servizio
description: Comprendi il contratto delle directory e il modello di esecuzione di un servizio Macaron.
---

Un servizio è una normale directory applicativa che contiene una cartella `.macaron`:

```text
service-name/
├── .macaron/
│   ├── build       # facoltativo
│   ├── cleanup     # facoltativo
│   ├── doctor      # facoltativo
│   └── start       # necessario per l'avvio
└── ... file dell'applicazione
```

Solo i file regolari sono riconosciuti come script del ciclo di vita. Uno script può essere eseguibile e avere il proprio shebang oppure non essere eseguibile; in questo caso viene avviato con Bash. Ogni script usa la directory principale del servizio come directory di lavoro.

## Servizio minimo

Questo esempio pubblica la directory corrente con Python:

```bash title=".macaron/start"
#!/usr/bin/env bash
set -euo pipefail

exec python3 -m http.server "$MACARON_AVAILABLE_PORT" --bind 0.0.0.0
```

Rendi eseguibile lo script se vuoi che lo shebang scelga l'interprete:

```sh
chmod +x .macaron/start
```

Installa la directory e avvia Macaron:

```sh
macaron install ./service-name
macaron
```

## Contratto di esecuzione

Macaron fornisce `MACARON_AVAILABLE_PORT` allo script di avvio. Il valore è una porta TCP libera, a partire dalla `49001`. L'applicazione deve usare questo valore invece di una porta fissa.

Lo script deve rimanere in foreground. Usa preferibilmente `exec` per il comando finale, così i segnali raggiungono direttamente l'applicazione e Macaron può seguirne il processo reale:

```bash
exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Dopo aver lanciato tutti i servizi, Macaron attende un secondo. Un processo già terminato viene considerato fallito e non compare nell'elenco dei servizi attivi.

## Log e arresto

Standard output e standard error vengono mostrati nel terminale di Macaron insieme al nome del servizio. Durante l'arresto, il processo riceve `SIGTERM` e ha fino a un secondo per terminare prima di essere chiuso forzatamente. Implementa una gestione corretta di `SIGTERM` se il servizio deve salvare dati o chiudere connessioni.
