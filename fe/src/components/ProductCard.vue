<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { ImageOff } from 'lucide-vue-next'
import { photoUrl } from '@/lib/api'
import type { ProductListItem } from '@/types'
import PriceBlock from '@/components/PriceBlock.vue'

defineProps<{ product: ProductListItem }>()
</script>

<template>
  <RouterLink
    :to="{ name: 'product', params: { slug: product.slug } }"
    class="block overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-navy/5 transition active:scale-[0.98]"
  >
    <div class="relative aspect-[4/3] bg-brick-soft">
      <img
        v-if="product.primary_photo_id"
        :src="photoUrl(product.primary_photo_id)"
        :alt="product.title"
        loading="lazy"
        class="h-full w-full object-cover"
      />
      <div v-else class="flex h-full w-full items-center justify-center text-brick/40">
        <ImageOff class="h-10 w-10" />
      </div>
    </div>
    <div class="p-3">
      <h2 class="mb-1 line-clamp-2 text-sm font-bold leading-snug">{{ product.title }}</h2>
      <PriceBlock
        :price-pix="product.price_pix"
        :price-card-total="product.price_card_total"
        :card-installments="product.card_installments"
        compact
      />
    </div>
  </RouterLink>
</template>
