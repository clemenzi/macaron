---
title: Gestire i servizi
description: Elenca, aggiorna, disabilita, abilita ed elimina i servizi Macaron installati.
---

## Elencare i servizi

```sh
macaron list
```

I servizi abilitati e disabilitati sono mostrati separatamente. Quelli abilitati si trovano in `services/`; quelli disabilitati in `services-disabled/`.

## Aggiornare i servizi

```sh
macaron update
```

Macaron esegue `git pull` per ogni servizio Git abilitato. Dopo un pull riuscito, mostra la revisione breve del commit ed esegue automaticamente `.macaron/build`, se presente. Le directory locali che non sono worktree Git vengono ignorate. I servizi disabilitati non vengono aggiornati.

Il comando si ferma al primo errore di pull o build: risolvi il problema prima di eseguirlo nuovamente.

## Disabilitare e abilitare

```sh
macaron disable dashboard
macaron enable dashboard
```

La disabilitazione sposta la directory del servizio invece di modificare un flag. I servizi disabilitati sono esclusi da avvio, aggiornamenti, controlli doctor e cleanup. Ripetere un comando quando il servizio è già nello stato richiesto è sicuro.

L'operazione fallisce se nella destinazione esiste già un servizio con lo stesso nome.

## Eliminare

```sh
macaron delete dashboard
```

`delete` accetta il nome della directory, non l'URL del repository, e può rimuovere un servizio abilitato o disabilitato. Se il nome esiste in entrambe le posizioni, Macaron rifiuta di scegliere. Eliminare un servizio inesistente non produce errori.

:::danger
L'eliminazione rimuove ricorsivamente la directory del servizio. Macaron non conserva backup.
:::
