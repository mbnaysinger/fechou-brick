const brl = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })

export const formatBRL = (value: number) => brl.format(value)

/** Valor da parcela com 2 casas (mesma regra da spec: total / parcelas) */
export const installmentValue = (total: number, installments: number) =>
  Math.round((total / installments) * 100) / 100
