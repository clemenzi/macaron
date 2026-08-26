---
title: Installazione
description: Requisiti, funzionamento dell'installer, aggiornamenti e disinstallazione di Macaron.
---

## Requisiti

Macaron richiede:

- un Mac con processore Apple Silicon o Intel;
- [Tailscale](https://tailscale.com/download/mac) installato, disponibile nel `PATH` e autenticato;
- un account autorizzato a usare `sudo`;
- Git per installare o aggiornare servizi da repository.

Prima della prima sessione, controlla la configurazione:

```sh
macaron doctor
```

## Installare il binario

```sh
curl -L https://raw.githubusercontent.com/clemenzi/macaron/refs/heads/main/install.sh | sudo bash
```

L'installer rileva `arm64` o `x86_64`, scarica il binario corrispondente dall'ultima release GitHub e lo installa in:

```text
/usr/local/bin/macaron
```

Crea inoltre le directory dei servizi abilitati e disabilitati per l'utente che ha invocato `sudo`. Go non è necessario sul Mac di destinazione.

:::note
Il comando scarica ed esegue uno script shell con privilegi di amministratore. Se questo non è compatibile con le tue regole di sicurezza, controlla prima [`install.sh`](https://github.com/clemenzi/macaron/blob/main/install.sh).
:::

## Aggiornare Macaron

Usa l'aggiornamento integrato:

```sh
macaron self-update
```

Il comando scarica l'installer corrente in un file temporaneo e lo esegue con `sudo`. Eseguire nuovamente il comando di installazione produce lo stesso risultato.

## Disinstallare

Macaron non offre ancora un comando di disinstallazione. Rimuovi manualmente il binario se non ti serve più:

```sh
sudo rm /usr/local/bin/macaron
```

I dati dei servizi sono separati e rimangono in `~/.config/macaron/` finché non li elimini manualmente.
