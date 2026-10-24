---
type: tdd
id: tdd:modeling:policy
context: modeling
title: Modeling Policy
status: Aceito
tags: [modeling, policy, generics, optional, predicate, testing]
related: [adr:modeling:005, adr:modeling:009, adr:modeling:008, tdd:modeling:specification, tdd:modeling:rules-requirements]
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
func Compose[A, B, C any](first Policy[A, B], second Policy[B, C]) Policy[A, C]
func Zip[I, A, B, O any](left Policy[I, A], right Policy[I, B], combine func(A, B) O) Policy[I, O]
func Choose[I, O any](condition predicate.Predicate[I], whenTrue Policy[I, O], whenFalse Policy[I, O]) Policy[I, O]
func OrElse[I, O any](left func(I) optional.Optional[O], right func(I) optional.Optional[O]) func(I) optional.Optional[O]
```

Apply, FlatMap, Conditional nominal, multi-aridade de input, métodos, aliases,
tuple pública e bridges não fazem parte do contrato. Múltiplos dados entram por
um contexto explícito em I.

## Semântica

Constant ignora o input. Map transforma apenas o output e satisfaz identidade e
composição. Contramap adapta apenas o input e satisfaz as leis correspondentes.
When avalia o predicate uma vez e só chama a policy quando true; true sempre
produz Some, inclusive para zero value. FirstApplicable é não vazio, ordenado e
para no primeiro Some. Fallback preserva Some e só chama fallback em None.
Compose avalia first e depois second. Zip avalia left, right e combine nessa
ordem. Choose avalia o predicate uma vez e somente o branch selecionado.
OrElse é binário, left-to-right e faz short-circuit; FirstApplicable permanece
a forma N-ária.

Construção com policy, mapper, predicate ou conditional nil panica com,
respectivamente, `modeling/policy: nil policy`, `modeling/policy: nil mapper`,
`modeling/policy: nil predicate` e `modeling/policy: nil conditional`.
Zip também rejeita combiner nil com `modeling/policy: nil combiner`.

## Dependências e arquitetura

Somente `conditional.go` importa Predicate e Optional, além do stdlib `slices`.
`modeling/policy` não importa raiz, siblings ou internal. Gates AST limitam
exports exatos; gates de dependência mantêm a exceção funcional restrita ao
package e aos dois imports functional.

## Testes e leis

Testes black-box fixam assinaturas, ordinary-function assignability, zero values,
contagem/ordem de chamadas, leis Map/Contramap, short-circuit, Some(zero),
all-None, fallback lazy, Compose associativity, Zip order, Choose branch único,
OrElse short-circuit e panics nil. Exemplos executáveis mostram todos os
combinadores sem efeitos; fuzzing bounded cobre Compose associativity.

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
