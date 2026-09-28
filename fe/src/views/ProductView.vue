<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import { ArrowLeft, MessageCircle, Share2 } from 'lucide-vue-next'
import { catalog, ApiError } from '@/lib/api'
import type { ProductDetail } from '@/types'
import BrandHeader from '@/components/BrandHeader.vue'
import PhotoCarousel from '@/components/PhotoCarousel.vue'
import PriceBlock from '@/components/PriceBlock.vue'

const route = useRoute()
const product = ref<ProductDetail | null>(null)
const notFound = ref(false)
const loading = ref(true)

onMounted(async () => {
  try {
    product.value = await catalog.getBySlug(route.params.slug as string)
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound.value = true
    else toast.error(e instanceof Error ? e.message : 'Erro ao carregar o anúncio')
  } finally {
    loading.value = false
  }
})

const productUrl = computed(() => window.location.href)

const whatsappLink = computed(() => {
  if (!product.value) return '#'
  const text = `Olá! Vi o anúncio "${product.value.title}" no fechou o brick e tenho interesse: ${productUrl.value}`
  const sep = product.value.whatsapp_url.includes('?') ? '&' : '?'
  return `${product.value.whatsapp_url}${sep}text=${encodeURIComponent(text)}`
})

async function share() {
  if (!product.value) return
  const data = { title: product.value.title, url: productUrl.value }
  if (navigator.share) {
    try {
      await navigator.share(data)
    } catch {
      /* usuário cancelou o share sheet */
    }
    return
  }
  await navigator.clipboard.writeText(productUrl.value)
  toast.success('Link copiado!')
}
</script>

<template>
  <div class="min-h-dvh pb-28">
    <BrandHeader>
      <RouterLink to="/" class="ml-auto flex items-center gap-1 text-sm font-semibold text-navy/60">
        <ArrowLeft class="h-4 w-4" /> ver tudo
      </RouterLink>
    </BrandHeader>

    <main v-if="product" class="mx-auto max-w-3xl">
      <div class="relative sm:mt-4 sm:overflow-hidden sm:rounded-2xl sm:px-4">
        <PhotoCarousel :photos="product.photos" :title="product.title" />
        <div
          v-if="product.status === 'VENDIDO'"
          class="absolute inset-x-0 top-1/2 -translate-y-1/2 bg-navy/85 py-3 text-center text-lg font-extrabold uppercase tracking-widest text-white"
        >
          Vendido 🤝
        </div>
      </div>

      <div class="px-4 py-4">
        <h1 class="mb-2 text-xl font-extrabold leading-tight">{{ product.title }}</h1>
        <PriceBlock
          :price-pix="product.price_pix"
          :price-card-total="product.price_card_total"
          :card-installments="product.card_installments"
        />
        <p v-if="product.description" class="mt-4 whitespace-pre-line text-[15px] leading-relaxed text-navy/80">
          {{ product.description }}
        </p>
      </div>
    </main>

    <main v-else-if="notFound" class="flex flex-col items-center gap-3 px-4 py-24 text-center">
      <p class="text-4xl">🧱</p>
      <h1 class="text-lg font-extrabold">Anúncio não encontrado</h1>
      <p class="text-sm text-navy/60">Pode ter sido removido ou o link está errado.</p>
      <RouterLink to="/" class="mt-2 rounded-full bg-brick px-5 py-2.5 font-bold text-white">
        Ver todos os anúncios
      </RouterLink>
    </main>

    <main v-else class="mx-auto max-w-3xl animate-pulse">
      <div class="aspect-[4/3] bg-navy/10 sm:mx-4 sm:mt-4 sm:rounded-2xl" />
      <div class="space-y-3 px-4 py-4">
        <div class="h-6 w-2/3 rounded bg-navy/10" />
        <div class="h-7 w-1/2 rounded bg-navy/10" />
        <div class="h-4 w-full rounded bg-navy/10" />
      </div>
    </main>

    <!-- Barra fixa: WhatsApp + compartilhar sempre à mão no mobile -->
    <div
      v-if="product && product.status !== 'VENDIDO'"
      class="fixed inset-x-0 bottom-0 z-30 border-t border-navy/5 bg-white/95 pb-[max(env(safe-area-inset-bottom),0.75rem)] pt-3 backdrop-blur"
    >
      <div class="mx-auto flex max-w-3xl gap-2 px-4">
        <a
          :href="whatsappLink"
          target="_blank"
          rel="noopener"
          class="flex flex-1 items-center justify-center gap-2 rounded-full bg-wa px-4 py-3 font-extrabold text-white transition active:bg-wa-dark"
        >
          <MessageCircle class="h-5 w-5" /> Chamar no WhatsApp
        </a>
        <button
          type="button"
          class="flex items-center justify-center gap-2 rounded-full border-2 border-navy/15 bg-white px-4 py-3 font-bold text-navy transition active:bg-cream"
          aria-label="Compartilhar"
          @click="share"
        >
          <Share2 class="h-5 w-5" />
        </button>
      </div>
    </div>
  </div>
</template>
