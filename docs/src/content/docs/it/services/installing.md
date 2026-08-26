---
title: Installare un servizio
description: Aggiungi servizi da repository Git o directory locali e controlla le verifiche iniziali.
---

## Installare da Git

Passa qualsiasi origine accettata da `git clone`:

```sh
macaron install https://github.com/example/macaron-service-dashboard.git
```

Il nome di destinazione deriva dal repository. Il prefisso iniziale `macaron-service-` viene rimosso, quindi l'esempio diventa `dashboard`.

Quando serve, scegli un branch o un nome esplicito:

```sh
macaron install --branch develop --name dashboard \
  https://github.com/example/dashboard.git
```

## Installare da una directory locale

```sh
macaron install ./my-service
```

Macaron copia la directory, inclusi file nascosti e collegamenti simbolici. L'origine non deve essere necessariamente un repository Git. `--branch` non può essere usato con una directory locale.

## Compilare e verificare

Se esiste `.macaron/build`, Macaron chiede conferma prima di eseguirlo. Usa `--yes` per un'installazione automatica oppure `--skip-build` per ignorarlo:

```sh
macaron install --yes <origine>
macaron install --skip-build <origine>
```

Se esiste `.macaron/doctor`, viene eseguito dopo la build. Un controllo fallito viene segnalato ma non rimuove il servizio installato. Puoi saltarlo con `--skip-doctor`.

:::caution
Gli script di build provengono dall'origine del servizio e vengono eseguiti come utente corrente. Controlla il servizio prima di approvare lo script.
:::

## Nomi e conflitti

Il nome deve essere una singola componente di percorso non vuota. Sono rifiutati i nomi contenenti `/` o `\\` e i nomi `.` e `..`. L'installazione si interrompe anche se esiste già un servizio abilitato con lo stesso nome di destinazione.
