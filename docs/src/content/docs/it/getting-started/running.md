---
title: Avvio e connessione
description: Avvia Macaron, collegati tramite Tailscale e comprendi la procedura di arresto.
---

## Avviare una sessione

Entrambi i comandi avviano Macaron:

```sh
macaron
# oppure
macaron start
```

Macaron convalida prima le credenziali `sudo`, registra lo stato corrente del sistema e poi:

1. abilita Remote Login di macOS;
2. disabilita la sospensione del sistema;
3. avvia Tailscale;
4. avvia ogni servizio abilitato che contiene `.macaron/start`;
5. mostra il comando SSH e l'URL di ogni servizio disponibile.

I servizi partono in ordine alfabetico. Macaron assegna a ciascuno la prima porta TCP libera a partire dalla `49001`.

## Collegarsi da remoto

Il terminale mostra un comando SSH simile a:

```sh
ssh tuo-utente@100.x.y.z
```

Mostra inoltre un URL HTTP per ogni servizio ancora in esecuzione dopo il controllo iniziale:

```text
dashboard  http://100.x.y.z:49001
```

Il dispositivo remoto deve appartenere alla stessa rete Tailscale ed essere autorizzato dalle relative regole di accesso. Il servizio deve ascoltare sulla porta assegnata e su un indirizzo raggiungibile tramite Tailscale, come `0.0.0.0`.

## Arrestare in sicurezza

Premi `Ctrl-C` oppure invia `SIGTERM` al processo Macaron. Durante l'arresto Macaron:

1. invia `SIGTERM` a ogni servizio gestito;
2. termina forzatamente i servizi che non si fermano entro un secondo;
3. esegue lo script di cleanup di ogni servizio abilitato;
4. rimuove `active-services.json`;
5. ripristina gli stati precedenti di Remote Login, sospensione e Tailscale.

Mantieni aperta la sessione del terminale finché ti serve l'accesso remoto. Se Macaron termina a causa di un errore di avvio, la pulizia differita tenta comunque di ripristinare lo stato del sistema acquisito in precedenza.
