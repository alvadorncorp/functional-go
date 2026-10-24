---
type: adr
id: adr:functional:014
context: functional
title: Higiene da superfície v1 — sinônimos Deprecated, núcleo enxuto, Lazy nativo e state congelado
status: Aceita
tags: [functional, v1, deprecation, lazy, hamt, core, state, iterator, predicate, tuple]
deciders: [alvadorn]
related: [adr:functional:004, adr:functional:006, adr:functional:009]
date: 2026-10-23
---
# ADR-014: Higiene da superfície v1 — sinônimos Deprecated, núcleo enxuto, Lazy nativo e state congelado

## Contexto e Problema

Depois das fases de coleções, `functional` acumulou resíduos que não mudam o
contrato público, mas confundem quem lê ou mantém o módulo:

- `structural/iterator.FromSlice`, `FromValues` e `Collect`,
  `primitives/predicate.All`/`Any`/`None` e `structural/tuple.FromValues2` são
  sinônimos finos de funções da stdlib ou de outro construtor do próprio
  pacote (`slices.Values`, `slices.Collect`, `slices.ContainsFunc`, `Of2`).
- `internal/core/sharing` (`Detach`/`Published`) e `cow.NeedsCopy` não têm
  chamador em produção: o Freeze `O(1)` por `Invalidate` (ADR-005) tornou a
  caminhada de `Detach` desnecessária, e `NeedsCopy` é apenas
  `!ownership.IsOwnedBy`.
- A HAMT interna mantém caminhos de `merge` que nenhum chamador alcança
  (hashes iguais, `shift` além do máximo) e deixa, depois de `Delete`, filhos
  com um único elemento terminal dentro de nós intermediários.
- `structural/lazy` reimplementa à mão o que `sync.OnceValue` já faz, com uma
  divergência real: um `runtime.Goexit` dentro do inicializador deixa o `Lazy`
  marcado como concluído com o valor zero, e todo `Get` seguinte devolve esse
  zero sem sinal algum.
- `state` está entregue e estável, mas nenhum documento diz se o pacote evolui,
  é substituído ou é removido.

O dono aprovou, em 2026-10-23, apenas estes itens. A mudança precisa
preservar toda a API pública v1 (todos os nomes exportados continuam
compilando) e a fronteira de dependências do ADR-003.

## Fatores de Decisão

- Nenhum nome exportado v1 é removido; remoção quebraria consumidores sem
  ganho proporcional.
- Preferir contratos nativos da stdlib à reimplementação manual (ADR-009).
- Sinalizar aos consumidores o caminho preferido sem suprimir lint nem relaxar
  gates.
- Remover código interno sem chamador e ramos inalcançáveis, com migração
  completa dos chamadores e sem shims.
- Compressão da HAMT só onde é segura: o bitmap depende da posição (`shift`) e
  não pode subir de nível.
- Não afirmar comportamento ou custo não medido.

## Opções Consideradas

### Opção 1: Manter tudo como está

Sem risco de comportamento. Os sinônimos continuam anunciados como API
preferida, o código morto continua exigindo leitura e testes, o `Lazy` mantém
o zero silencioso após `Goexit` e o futuro de `state` segue ambíguo.

### Opção 2: Remover sinônimos e `state` na linha v1

Deixa a superfície mínima. Quebra qualquer consumidor existente dos nomes
removidos e contradiz a promessa de compatibilidade v1; o dono não aprovou
remoções.

### Opção 3: Marcar sinônimos como `Deprecated`, limpar internals, adotar `sync.OnceValue` e congelar `state` (escolhida)

Mantém todos os nomes exportados, orienta consumidores para a stdlib por meio
do marcador `Deprecated:` padrão, remove somente código interno sem chamador e
troca a máquina manual do `Lazy` pela primitiva nativa.

## Decisão

**Opção escolhida: Opção 3**, porque é a única que satisfaz ao mesmo tempo
"nenhum nome v1 removido" e "um único caminho preferido por capacidade".

1. **Sinônimos `Deprecated` (sem remoção).** `iterator.FromSlice` e
   `iterator.FromValues` apontam para `slices.Values`; `iterator.Collect`
   aponta para `slices.Collect`; `predicate.All`/`Any`/`None` apontam para
   `slices.ContainsFunc` (e para os quantificadores `iterator.All`/`Any`, em
   quem já importa `iterator`); `tuple.FromValues2` aponta para `Of2`. A
   semântica não muda: os adaptadores de origem continuam sendo **visões vivas**
   do slice do chamador, sem cópia. `predicate` segue folha só-stdlib e não
   importa `iterator`. Os exemplos, benchmarks e consumidores reais migram para
   os substitutos.
2. **`state` congelado, não `Deprecated`.** A API existente é suportada e
   estável, sem novo surface, sem marcador `Deprecated` no pacote e sem
   promessa de remoção ou de v2 (emenda parcial no ADR-006, que continua
   `Aceita`).
3. **Núcleo enxuto.** O pacote `internal/core/sharing` e `cow.NeedsCopy` são
   removidos; o núcleo passa de quatro para três pacotes (`ownership`, `cow`,
   `transient`). Quem precisava de `NeedsCopy` usa `!ownership.IsOwnedBy`, com
   a mesma semântica e a mesma ordem de argumentos. As stamps de owners
   inativos permanecem como estão (Freeze `O(1)` por invalidação). É emenda
   parcial do ADR-004, que continua `Aceita`.
4. **HAMT sem ramos mortos e com compressão terminal.** `merge` só é chamado
   com hashes distintos; os ramos de hash igual e de `shift` acima do máximo
   viram `panic("hamt: unreachable merge")`, e `joinCollision` e
   `collectEntries` saem. Em `Delete` (puro e transitório), quando um filho
   passa a conter só uma folha ou colisão, o pai guarda esse terminal
   diretamente. **Nunca** se promove um bitmap, por ser sensível ao `shift`.
   Contagem, owner e snapshots anteriores ficam preservados. Nada disso é
   visível na API pública.
5. **`Lazy` sobre `sync.OnceValue`.** O `Lazy[T]` passa a guardar um getter
   `func() T` criado por `sync.OnceValue`. `New` com inicializador nil continua
   com o mesmo panic; o `Lazy` de valor zero e o `*Lazy` nil continuam
   documentados como inválidos, sem guarda extra. Ficam preservados: execução
   única mesmo com chamadores concorrentes, cache do valor e cache do payload de
   panic comum, e a liberação da função inicializadora. A única diferença
   comportamental aprovada é o `runtime.Goexit` dentro do inicializador: o
   `Lazy` deixa de gravar o valor zero como sucesso: o próprio `Get` e todo `Get`
   seguinte entram em panic. Sob a conversão padrão de `panic(nil)` do Go
   (módulo principal com `go >= 1.21` e `GODEBUG panicnil` sem override) o
   payload é `*runtime.PanicNilError`; sob o override legado `panicnil=1`,
   `sync.OnceValue` segue o runtime e o tipo do payload não é garantido, embora
   o zero nunca seja devolvido como sucesso. Não há mudança no tratamento padrão
   de `panic(nil)`: o Go moderno já o converte em `*runtime.PanicNilError` antes
   do `recover`, e esse comportamento não é objeto desta decisão. Este ponto
   corrige a descrição do ADR-009 (emenda parcial, que continua `Aceita`).

Alternativas rejeitadas por item: remover os sinônimos (Opção 2); promover
nós bitmap de um filho para o pai na HAMT (incorreto, o bitmap depende do
`shift`); guarda explícita para `Lazy` zero ou nil e opções de `Lazy` (escopo
não aprovado); marcar `state` como `Deprecated` (o pacote não está sendo
retirado).

## Consequências

**Positivas:**

- Um caminho preferido por capacidade, anunciado pelo marcador `Deprecated:`,
  sem remover nenhum nome v1.
- O núcleo fica com três pacotes e sem API sem chamador; o grafo de
  dependências internas encolhe (`hamt` deixa de importar `sharing`).
- `Lazy` perde a máquina manual de `Once`/valor/panic e fecha o zero
  silencioso após `Goexit`.
- A estrutura da HAMT depois de `Delete` fica canônica para terminais, sem
  transientes extras.

**Negativas:**

- Os marcadores `Deprecated:` fazem o linter (SA1019) apontar os usos
  existentes nos consumidores, que precisam migrar ou aceitar o aviso. Os
  nomes continuam disponíveis, então os sinônimos ficam como superfície a
  manter.
- Testes de compatibilidade dos nomes `Deprecated` precisam ficar no pacote
  que os define, para não violar o lint, e não podem virar testes de
  tautologia da stdlib.
- O `Goexit` dentro de um inicializador de `Lazy` passa a ter efeito
  observável diferente do anterior (panic no `Get` e nos seguintes em vez do zero
  silencioso). É uma mudança de comportamento aprovada, não só interna.
- Medição observada do `Lazy` com `sync.OnceValue`, restrita aos benchmarks
  existentes (`go -C functional test -run '^$' -bench
  '^BenchmarkLazy(ColdGet|WarmGet)$' -benchmem ./structural/lazy`, Go 1.27.3,
  darwin/arm64, Apple M2 Max, **uma amostra** cada): `BenchmarkLazyColdGet`
  passou de 64 B/op e 1 alocação/op para 48 B/op e 1 alocação/op (103,7 ns/op
  depois), e `BenchmarkLazyWarmGet` ficou em 0 B/op e 0 alocações/op nos dois
  casos (7,087 ns/op depois). É uma comparação de alocações naqueles dois
  benchmarks, não uma análise estatística de latência nem uma garantia geral
  de alocação da API.
- A HAMT passa a depender do invariante "merge só recebe hashes distintos" de
  seus chamadores; uma quebra desse invariante vira panic em vez de um ramo
  silencioso.
- `state` congelado e não `Deprecated` não dá ao consumidor sinal sobre a
  direção de longo prazo do pacote.

**Neutras / Ações de acompanhamento:**

- ADR-004 e ADR-009 recebem emendas parciais e continuam `Aceita`; ADR-006
  recebe a emenda do `state` congelado. Nenhuma ADR é substituída por inteiro.
- Os guias de `iterator`, `predicate`, `tuple`, `lazy`, `fold` e `core`, o
  overview e os índices passam a descrever o estado pós-decisão.
- Fora de escopo: qualquer remoção de nome v1, novas APIs, opções do `Lazy` e
  mudanças em `modeling` ou `workflow`.

---

*ADRs relacionadas: adr:functional:004, adr:functional:006, adr:functional:009*
