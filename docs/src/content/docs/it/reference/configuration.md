---
title: File e stato di esecuzione
description: Riferimento per directory, variabili d'ambiente, porte e stato generato da Macaron.
---

## Directory di configurazione

Macaron rispetta `XDG_CONFIG_HOME`. La directory principale è:

```text
$XDG_CONFIG_HOME/macaron/
```

Quando `XDG_CONFIG_HOME` non è definita, il valore predefinito è:

```text
~/.config/macaron/
```

| Percorso | Scopo |
| --- | --- |
| `services/` | Servizi abilitati |
| `services-disabled/` | Servizi disabilitati |
| `active-services.json` | Servizi che hanno superato il controllo iniziale e relative porte |

## File dei servizi attivi

Mentre Macaron è in esecuzione, `active-services.json` contiene un array simile a:

```json
[
  { "name": "dashboard", "port": 49001 },
  { "name": "notes", "port": 49002 }
]
```

Il file viene scritto atomicamente dopo l'avvio e rimosso durante l'arresto. I servizi che terminano nella finestra iniziale di un secondo non vengono inclusi. È uno stato temporaneo di esecuzione, non una configurazione permanente.

## Assegnazione delle porte

Macaron controlla le porte TCP in ordine crescente a partire dalla `49001` e salta quelle occupate. Ogni script di avvio riceve la porta scelta tramite:

```text
MACARON_AVAILABLE_PORT
```

La variabile viene fornita soltanto al processo del servizio. Macaron non configura l'indirizzo di ascolto o il protocollo dell'applicazione.

## Stato del sistema

All'avvio Macaron registra se Remote Login è abilitato, se la sospensione è già disattivata e se Tailscale è attivo. Poi abilita tutte e tre le funzionalità. Durante l'arresto tenta di ripristinare ogni valore registrato anche se la pulizia di un servizio segnala un errore.

Macaron non modifica chiavi SSH, regole di accesso Tailscale, regole firewall o configurazione delle applicazioni.
