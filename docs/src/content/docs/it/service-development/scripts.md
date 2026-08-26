---
title: Script del ciclo di vita
description: Definisci build, avvio, diagnostica e pulizia di un servizio Macaron.
---

Tutti gli script si trovano in `.macaron/` e vengono eseguiti dalla directory principale del servizio.

## `start`

`start` è l'unico script necessario per avviare un servizio. Riceve `MACARON_AVAILABLE_PORT`, deve mantenere l'applicazione in foreground e deve usare un'interfaccia raggiungibile tramite Tailscale.

```bash title=".macaron/start"
#!/usr/bin/env bash
set -euo pipefail

exec ./server --host 0.0.0.0 --port "$MACARON_AVAILABLE_PORT"
```

Non aggiungere `&`, non avviare il processo come demone e non terminare dopo aver creato un processo scollegato. Macaron considererebbe il servizio arrestato e non potrebbe chiuderlo in modo affidabile.

## `build`

Usa `build` per installare dipendenze o compilare l'applicazione:

```bash title=".macaron/build"
#!/usr/bin/env bash
set -euo pipefail

npm ci
npm run build
```

Durante l'installazione, Macaron chiede conferma prima di eseguirlo. Durante `macaron update`, lo esegue automaticamente dopo ogni pull Git riuscito. Un errore interrompe l'installazione o la restante sequenza di aggiornamento.

## `doctor`

Usa `doctor` per verifiche di configurazione che non richiedono l'avvio del servizio. Termina con codice `0` quando tutto è corretto e con un valore diverso da zero quando serve un intervento:

```bash title=".macaron/doctor"
#!/usr/bin/env bash
set -euo pipefail

test -f .env || {
  echo "File .env mancante"
  exit 1
}
```

Viene eseguito dopo l'installazione, salvo esclusione, e a ogni `macaron doctor`. Il doctor globale controlla soltanto i servizi abilitati.

## `cleanup`

`cleanup` viene eseguito dopo l'arresto dei processi gestiti e prima del ripristino delle impostazioni di sistema:

```bash title=".macaron/cleanup"
#!/usr/bin/env bash
set -euo pipefail

rm -f .runtime/socket
```

Usalo per risorse temporanee appartenenti al servizio, non per fermare l'applicazione principale. Gli errori vengono segnalati ma non impediscono il tentativo di pulire gli altri servizi e ripristinare il sistema.

## Portabilità degli script

- Dichiara strumenti e configurazione richiesti in `doctor`.
- Usa percorsi relativi alla directory principale del servizio.
- Metti tra virgolette `"$MACARON_AVAILABLE_PORT"` e le altre variabili.
- Usa `set -euo pipefail` negli script Bash quando è preferibile interrompersi subito.
- Mantieni gli script non interattivi dopo la conferma della build in installazione.
