---
title: Servizi ufficiali
description: Esplora e installa i servizi mantenuti dal progetto Macaron.
---

I servizi ufficiali sono mantenuti nell'organizzazione GitHub [`macaron-services`](https://github.com/macaron-services). Seguono lo stesso contratto `.macaron` dei servizi personalizzati e possono essere controllati prima dell'installazione.

| Servizio | Scopo | Repository |
| --- | --- | --- |
| code-server | Esegue VS Code nel browser | [`macaron-services/code-server`](https://github.com/macaron-services/code-server) |
| dufs | Esplora e gestisce file dal browser | [`macaron-services/dufs`](https://github.com/macaron-services/dufs) |

## code-server

Il servizio code-server fornisce un ambiente VS Code nel browser, eseguito direttamente sul Mac.

Installalo con:

```sh
macaron install https://github.com/macaron-services/code-server
```

Durante l'installazione, lo script di build controlla se `code-server` è disponibile e, se manca, propone di installarlo tramite Homebrew. Il servizio ascolta sulla porta scelta da Macaron su tutte le interfacce di rete.

:::caution
Lo script di avvio ufficiale esegue code-server senza autenticazione applicativa. L'accesso deve essere limitato alla rete Tailscale fidata e alle relative regole di controllo degli accessi.
:::

- [Vedi il sorgente del servizio](https://github.com/macaron-services/code-server)
- [Scopri code-server](https://github.com/coder/code-server)

## dufs

Il servizio dufs espone una directory tramite un file manager web. Sono abilitate le operazioni di caricamento, download, modifica, ricerca, archiviazione ed eliminazione.

Installalo con:

```sh
macaron install https://github.com/macaron-services/dufs
```

Per impostazione predefinita, dufs pubblica la home dell'utente corrente. Definisci `DUFS_SERVE_PATH` quando avvii Macaron per esporre una posizione più limitata:

```sh
DUFS_SERVE_PATH="$HOME/Downloads" macaron start
```

Lo script di build controlla se `dufs` è disponibile e, se manca, propone di installarlo tramite Homebrew.

:::danger
Il servizio abilita tutte le operazioni sui file e non aggiunge autenticazione applicativa. Usa una rete Tailscale fidata e scegli un `DUFS_SERVE_PATH` limitato, a meno che tu non voglia esporre l'intera home ai membri autorizzati della rete.
:::

- [Vedi il sorgente del servizio](https://github.com/macaron-services/dufs)
- [Scopri dufs](https://github.com/sigoden/dufs)

## Trovare nuovi servizi

Questa pagina riflette l'organizzazione ufficiale al momento dell'ultimo aggiornamento della documentazione. Visita [`github.com/macaron-services`](https://github.com/macaron-services) per i nuovi servizi pubblicati e controlla gli script `.macaron` prima dell'installazione.
