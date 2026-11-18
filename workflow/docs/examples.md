# Exemplos executáveis

Os exemplos usam a API pública e não dependem de banco, broker ou serviços.
Execute `go test ./...` a partir do módulo `workflow` para compilá-los e
conferir a saída documentada.

- [Composição e injeção de dependências](../example_test.go): a construção
  recebe as capacidades; `action.Run` só executa efeitos depois de aplicar
  o workflow à entrada.
- [Ações, composição e finalização](../action/example_test.go): use `Attempt`
  para tratar uma falha como dado e `Bracket` para delimitar um recurso.
- [Fronteira de transação em memória](../transaction/example_test.go): o
  adaptador fornece capacidades escopadas a `transaction.Within`; o mesmo
  exemplo mostra sucesso e erro de domínio.
- [Planos](../plan/example_test.go): descreva passos e transações sem executá-los.
- [Resiliência](../resilience/example_test.go): `resilience.Retry` repete a
  ação conforme uma política explícita; a posição de `Timeout` decide entre
  um orçamento total e um por tentativa.
