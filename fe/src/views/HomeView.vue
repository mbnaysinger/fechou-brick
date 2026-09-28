<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useIntersectionObserver } from '@vueuse/core'
import { PackageOpen } from 'lucide-vue-next'
import { useProductsStore } from '@/stores/products'
import BrandHeader from '@/components/BrandHeader.vue'
import ProductCard from '@/components/ProductCard.vue'
import SkeletonCard from '@/components/SkeletonCard.vue'

const store = useProductsStore()
const sentinel = ref<HTMLElement | null>(null)

onMounted(() => {
  store.reset()
  store.fetchMore()
})

useIntersectionObserver(sentinel, ([entry]) => {
  if (entry?.isIntersecting) store.fetchMore()
}, { rootMargin: '400px' })
</script>

<template>
  <div class="min-h-dvh">
    <BrandHeader />

    <main class="mx-auto max-w-3xl px-4 py-4 pb-10">
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <ProductCard v-for="p in store.items" :key="p.id" :product="p" />
        <template v-if="store.loading">
          <SkeletonCard v-for="i in 4" :key="`sk-${i}`" />
        </template>
      </div>

      <div
        v-if="!store.loading && !store.items.length"
        class="flex flex-col items-center gap-3 py-20 text-center text-navy/50"
      >
        <PackageOpen class="h-12 w-12" />
        <p class="font-semibold">Nenhum anúncio por aqui ainda.<br />Volte em breve!</p>
      </div>

      <p v-if="store.error" class="py-6 text-center text-sm font-semibold text-brick-dark">
        {{ store.error }}
        <button type="button" class="ml-2 underline" @click="store.fetchMore()">tentar de novo</button>
      </p>

      <p v-if="store.finished && store.items.length" class="py-8 text-center text-sm text-navy/40">
        É isso! Fechou? 🤝
      </p>

      <div ref="sentinel" class="h-px" />
    </main>
  </div>
</template>
