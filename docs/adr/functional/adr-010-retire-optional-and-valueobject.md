---
type: adr
id: adr:functional:010
context: functional
title: Retire top-level optional and valueobject packages
status: Aceita
tags: [functional, optional, valueobject, dependencies, public-api]
deciders: [alvadorn]
related: [adr:functional:001, adr:functional:003, adr:functional:006, adr:functional:007]
date: 2026-08-10
---
# ADR-010: Retire top-level optional and valueobject packages

## Contexto e Problema

O módulo `functional` expunha dois pacotes públicos sem consumidores de
produção no monorepo:

- `github.com/alvadorncorp/gopackages/functional/optional` — API legada de
  `Optional[T]`, mantida em paralelo a `values/optional` (ADR-007).
- `github.com/alvadorncorp/gopackages/functional/valueobject` — helper
  reflexivo de teste (`Check`/`Inspect`), sem uso de produção.

A presença de dois Optional e de um helper de teste no inventário público
aumentava a superfície e a documentação sem valor operacional no grafo atual
(`fold` e o restante da família usam `values/*`).

## Fatores de Decisão

- Zero imports de produção no monorepo para ambos os paths.
- `values/optional` já é a presença canônica (ADR-008 / fold).
- Remoção é breaking para consumidores externos do path legado, se existirem.
- ADRs históricos (001, 007, …) permanecem como registro; o inventário vivo
  deixa de listar os pacotes.

## Opções Consideradas

### Opção 1: Manter ambos como legacy

Preserva compatibilidade total, mas mantém duplicidade e docs de manutenção
sem consumidor interno.

### Opção 2: Deprecar e remover em major futura

Mais seguro para consumidores externos desconhecidos; atrasa a limpeza.

### Opção 3: Remover agora (escolhida)

Delete os pacotes, micro-docs e referências vivas; registre a decisão aqui.
Aceita breaking change no próximo release do grupo `functional`.

## Decisão

Remover `functional/optional` e `functional/valueobject` e toda a documentação
de produto que os descreve como pacotes atuais. Atualizar archtest, índices e
overview. Escrever este ADR como fonte da decisão de retirement.

## Consequências

### Positivas

- Um único Optional público: `values/optional`.
- Superfície pública e inventário de archtest menores.
- Docs e exemplos alinhados ao grafo real (fold → values/optional).

### Negativas / trade-offs

- Breaking change: imports de
  `github.com/alvadorncorp/gopackages/functional/optional` e
  `.../valueobject` deixam de compilar.
- Consumidores externos precisam migrar para `values/optional` ou copiar o
  helper de value-object localmente.
- ADR-001 e trechos de ADR-007 passam a ser históricos em relação ao path
  top-level `optional`.

## Verificação

- `go list ./...` não lista `functional/optional` nem `functional/valueobject`.
- `go test ./internal/archtest` e `make test-unit` / `make lint` no módulo.
- Docs vivas sem links quebrados para `packages/optional.md` /
  `packages/valueobject.md`.
