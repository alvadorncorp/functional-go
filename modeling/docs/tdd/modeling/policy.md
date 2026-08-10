---
type: tdd
id: tdd:modeling:policy
context: modeling
title: Modeling Policy
status: Aceito
tags: [modeling, policy, generics, optional, predicate, testing]
related: [adr:modeling:005, tdd:modeling:rules-requirements]
date: 2026-08-10
---
# TDD: Modeling Policy

## Escopo técnico

Criar o package público `modeling/policy`, independente da raiz e de outras
capabilities. Change/Reducer/Replay não são pré-requisitos nem fazem parte
desta entrega.

## Contrato público

```go
type Policy[I, O any] func(I) O
func Constant[I, O any](output O) Policy[I, O]
func Map[I, O, P any](p Policy[I, O], mapper func(O) P) Policy[I, P]
func Contramap[A, I, O any](p Policy[I, O], mapper func(A) I) Policy[A, O]
func When[I, O any](condition predicate.Predicate[I], p Policy[I, O]) func(I) optional.Optional[O]
func FirstApplicable[I, O any](first func(I) optional.Optional[O], rest ...func(I) optional.Optional[O]) func(I) optional.Optional[O]
func Fallback[I, O any](conditional func(I) optional.Optional[O], fallback Policy[I, O]) Policy[I, O]
```

Apply, Conditional nominal, Condition, multi-aridade, métodos e aliases não
fazem parte do contrato. Múltiplos dados entram por um contexto explícito em I.

## Semântica

Constant ignora o input. Map transforma apenas o output e satisfaz identidade e
composição. Contramap adapta apenas o input e satisfaz as leis correspondentes.
When avalia o predicate uma vez e só chama a policy quando true; true sempre
produz Some, inclusive para zero value. FirstApplicable é não vazio, ordenado e
para no primeiro Some. Fallback preserva Some e só chama fallback em None.

Construção com policy, mapper, predicate ou conditional nil panica com,
respectivamente, `modeling/policy: nil policy`, `modeling/policy: nil mapper`,
`modeling/policy: nil predicate` e `modeling/policy: nil conditional`.

## Dependências e arquitetura

Somente `conditional.go` importa Predicate e Optional. `modeling/policy` não
importa raiz, siblings ou internal. Gates AST limitam exports exatos; gates de
dependência mantêm a exceção funcional restrita ao package e aos dois imports.

## Testes e leis

Testes black-box fixam assinaturas, ordinary-function assignability, zero values,
contagem/ordem de chamadas, leis Map/Contramap, short-circuit, Some(zero),
all-None, fallback lazy e panics nil. Um exemplo executável mostra
When → FirstApplicable → Fallback sem efeitos.

## Compatibilidade

A API pública existente permanece intacta. Testes rodam no workspace e com
`GOWORK=off` contra functional v1.1.0. A única mudança de metadata é o bump
autorizado de functional para v1.1.0; não há mudança de go.work, migração,
config, benchmark ou publicação.

## Superfície excluída

Não há Rule/Requirement/Transition/Change/Reducer bridge, engine, registry,
dispatcher, reflection, efeito, persistência, transporte, erro de aplicabilidade,
domain type concreto ou símbolo Policy na raiz.

## Verificação

Executar testes focados, módulo, published dependency, race, vet, tidy-diff,
lint uncapped, package listing e os gates workspace definidos no plano.
