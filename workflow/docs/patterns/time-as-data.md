# Tempo como dado

Regra de domínio que depende de “agora” recebe o fato já decidido no input
(checkout: `Coupon.expired bool`, não `time.Now()` no pacote de domínio).
`context` e relógio ficam no adapter. Não exporte `Clock` em `workflow`.
