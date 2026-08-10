---
type: adr
id: adr:functional:011
context: functional
title: Group public packages under primitives/ and structural/ path prefixes
status: Aceita
tags: [functional, layout, public-api, dependencies]
deciders: [alvadorn]
related: [adr:functional:007, adr:functional:008, adr:functional:009]
date: 2026-08-10
---
# ADR-011: Group public packages under primitives/ and structural/ path prefixes

## Contexto e Problema

Após a expansão do módulo (`values/*`, fold, tuple/lazy/iterator/optics), a
lista de pacotes públicos no root ficou longa. Agrupar por **papel** melhora
descoberta sem criar facade de reexport nem pasta `runtime/`.

## Fatores de Decisão

- Manter um conceito ≈ um pacote Go (sem package nos diretórios de grupo).
- Preservar o DAG e as arestas já archtestadas (só mudam os import paths).
- Evitar `runtime/` — `immutable` e `state` permanecem no root do módulo.
- Aceitar breaking change de paths (módulo ainda v0).

## Opções Consideradas

### Opção 1: Manter tudo flat no root

Zero churn de import; descoberta só por docs.

### Opção 2: `primitives/`, `structural/`, e `runtime/`

Inclui collections/state sob um prefixo; rejeitada — `immutable`/`state` não
compartilham lei de família com os leves e o prefixo `runtime` sugere
infraestrutura que o módulo não pretende unificar.

### Opção 3: Só `primitives/` e `structural/` (escolhida)

- `primitives/{function,predicate,order,fold}`
- `structural/{tuple,lazy,iterator,optics}`
- `values/*`, `immutable`, `state` inalterados em posição relativa

## Decisão

Mover os oito pacotes para os prefixos acima. Não criar package Go em
`primitives` ou `structural`. Arestas:

- `primitives/fold` → `values/optional`, `values/nonempty`
- `structural/iterator` → `structural/tuple`

## Consequências

**Positivas:** índice legível por papel; paths expressam intenção.

**Negativas:** breaking rename de oito import paths.

**Verificação:** `go list ./...`, archtest, `go test ./...`, lint.
