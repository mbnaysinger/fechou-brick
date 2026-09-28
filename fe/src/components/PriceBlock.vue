<script setup lang="ts">
import { computed } from 'vue'
import { Zap } from 'lucide-vue-next'
import { formatBRL, installmentValue } from '@/lib/money'

const props = defineProps<{
  pricePix: number
  priceCardTotal: number | null
  cardInstallments: number | null
  compact?: boolean
}>()

// Regras da spec: a linha do cartão é derivada, nunca digitada
const cardLine = computed(() => {
  const total = props.priceCardTotal
  if (total === null || total === undefined) return null
  const n = props.cardInstallments ?? 1
  if (n <= 1) {
    return { main: `ou ${formatBRL(total)} no cartão`, note: null, interestFree: false }
  }
  const parcela = formatBRL(installmentValue(total, n))
  if (total === props.pricePix) {
    return { main: `ou ${n}x de ${parcela}`, note: null, interestFree: true }
  }
  return { main: `ou ${n}x de ${parcela} no cartão`, note: `total ${formatBRL(total)}`, interestFree: false }
})
</script>

<template>
  <div>
    <p class="flex items-baseline gap-1.5 font-extrabold text-brick" :class="compact ? 'text-lg' : 'text-2xl'">
      {{ formatBRL(pricePix) }}
      <span class="inline-flex items-center gap-0.5 rounded-full bg-brick-soft px-2 py-0.5 text-[11px] font-bold uppercase tracking-wide text-brick-dark">
        <Zap class="h-3 w-3" /> no PIX
      </span>
    </p>

    <template v-if="cardLine">
      <p class="mt-0.5 text-navy/60" :class="compact ? 'truncate text-xs' : 'text-sm'">
        {{ cardLine.main }}
        <span
          v-if="cardLine.interestFree"
          class="ml-1 rounded-full bg-amber/20 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber"
        >sem juros</span>
      </p>
      <p v-if="cardLine.note && !compact" class="text-xs text-navy/40">{{ cardLine.note }}</p>
    </template>
    <p v-else class="mt-0.5 text-navy/40" :class="compact ? 'text-xs' : 'text-sm'">somente no PIX</p>
  </div>
</template>
