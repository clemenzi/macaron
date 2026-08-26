---
title: Riferimento CLI
description: Sintassi completa dei comandi Macaron, opzioni di installazione e codici di uscita.
---

## Comandi

| Comando | Descrizione |
| --- | --- |
| `macaron` | Avvia Macaron |
| `macaron start` | Avvia Macaron esplicitamente |
| `macaron doctor` | Controlla Tailscale e i servizi abilitati |
| `macaron install [opzioni] <origine>` | Installa un servizio |
| `macaron update` | Aggiorna e ricompila i servizi Git abilitati |
| `macaron disable <servizio>` | Sposta un servizio tra quelli disabilitati |
| `macaron enable <servizio>` | Sposta un servizio tra quelli abilitati |
| `macaron delete <servizio>` | Rimuove definitivamente la directory di un servizio |
| `macaron list` | Elenca i servizi abilitati e disabilitati |
| `macaron self-update` | Installa l'ultima release di Macaron |
| `macaron help` | Mostra la guida dei comandi |

## Opzioni di installazione

| Opzione | Descrizione |
| --- | --- |
| `--name <nome>` | Imposta il nome di destinazione del servizio |
| `--branch <branch>` | Clona un branch Git specifico |
| `--skip-build` | Non esegue `.macaron/build` |
| `--skip-doctor` | Non esegue `.macaron/doctor` dopo l'installazione |
| `-y`, `--yes` | Approva lo script di build senza chiedere conferma |
| `-h`, `--help` | Mostra la guida di installazione |

Le opzioni supportano la gestione delle forme brevi quando esiste un alias. `install` accetta esattamente un'origine; i comandi che cambiano stato accettano esattamente un nome di servizio.

## Codici di uscita

| Codice | Significato |
| --- | --- |
| `0` | Comando completato correttamente |
| `1` | Errore di esecuzione, dipendenza, servizio o controllo diagnostico |
| `2` | Argomenti o nome del servizio non validi, oppure comando sconosciuto |
