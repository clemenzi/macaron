---
title: Risoluzione dei problemi
description: Diagnostica problemi di Tailscale, avvio, servizi e pulizia in Macaron.
---

## Inizia da doctor

```sh
macaron doctor
```

Doctor verifica che il comando `tailscale` esista, legge lo stato JSON, rileva la necessità di autenticarsi ed esegue `.macaron/doctor` per ogni servizio abilitato che lo fornisce. Termina con codice `1` se un controllo fallisce.

## Tailscale non viene trovato

Verifica che la CLI di Tailscale sia installata e visibile nella shell che avvia Macaron:

```sh
command -v tailscale
tailscale status
```

Se Tailscale segnala `NeedsLogin`, autenticati prima di riprovare.

## Non viene trovato alcun servizio

Controlla la directory dei servizi abilitati e il file di avvio richiesto:

```sh
macaron list
ls -la ~/.config/macaron/services/<servizio>/.macaron/start
```

Vengono avviati soltanto i file regolari `.macaron/start`. I servizi disabilitati vengono ignorati intenzionalmente.

## Un servizio termina immediatamente

Esegui lo script dalla directory principale del servizio e fornisci una porta di prova:

```sh
cd ~/.config/macaron/services/<servizio>
MACARON_AVAILABLE_PORT=49001 bash .macaron/start
```

Lo script deve mantenere l'applicazione in foreground. Rimuovi l'esecuzione in background o le opzioni daemon e usa `exec` per il comando finale del server.

## L'URL non è raggiungibile

Verifica che:

- client e Mac siano connessi alla stessa rete Tailscale;
- le regole di accesso Tailscale consentano la connessione;
- l'applicazione usi `MACARON_AVAILABLE_PORT`;
- l'applicazione ascolti su `0.0.0.0` o su un altro indirizzo raggiungibile tramite Tailscale, non soltanto su `127.0.0.1`;
- autenticazione applicativa o firewall non stiano rifiutando la richiesta.

Usa `active-services.json` per confermare la porta assegnata da Macaron.

## Le impostazioni non sono state ripristinate

Macaron ripristina lo stato durante il normale arresto tramite `SIGINT` o `SIGTERM`. Un arresto forzato, un crash di sistema o una perdita di alimentazione non possono eseguire la pulizia. In quel caso, ripristina manualmente l'impostazione necessaria e riavvia Tailscale secondo le tue esigenze:

```sh
sudo systemsetup -setremotelogin off
sudo pmset -a disablesleep 0
tailscale down
```

Modifica soltanto i valori coerenti con lo stato desiderato per il tuo Mac.
