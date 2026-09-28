<script setup lang="ts">
import { ref } from 'vue'
import { ImageOff, X } from 'lucide-vue-next'
import { photoUrl } from '@/lib/api'
import type { PhotoMeta } from '@/types'

const props = defineProps<{ photos: PhotoMeta[]; title: string }>()

const activeIndex = ref(0)
const fullscreen = ref<PhotoMeta | null>(null)

// scroll-snap nativo; o indicador acompanha pela posição do scroll
function onScroll(e: Event) {
  const el = e.target as HTMLElement
  activeIndex.value = Math.round(el.scrollLeft / el.clientWidth)
}
</script>

<template>
  <div v-if="props.photos.length" class="relative">
    <div
      class="no-scrollbar flex snap-x snap-mandatory overflow-x-auto"
      @scroll.passive="onScroll"
    >
      <button
        v-for="photo in props.photos"
        :key="photo.id"
        type="button"
        class="aspect-[4/3] w-full flex-none snap-center bg-brick-soft"
        @click="fullscreen = photo"
      >
        <img :src="photoUrl(photo.id)" :alt="props.title" class="h-full w-full object-cover" />
      </button>
    </div>

    <div v-if="props.photos.length > 1" class="absolute bottom-2 left-1/2 flex -translate-x-1/2 gap-1.5">
      <span
        v-for="(photo, i) in props.photos"
        :key="photo.id"
        class="h-1.5 rounded-full transition-all"
        :class="i === activeIndex ? 'w-5 bg-white' : 'w-1.5 bg-white/60'"
      />
    </div>

    <Teleport to="body">
      <div
        v-if="fullscreen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-navy-dark/95 p-2"
        @click="fullscreen = null"
      >
        <button type="button" class="absolute right-4 top-4 rounded-full bg-white/10 p-2 text-white" aria-label="Fechar">
          <X class="h-5 w-5" />
        </button>
        <img :src="photoUrl(fullscreen.id)" :alt="props.title" class="max-h-full max-w-full rounded-lg object-contain" />
      </div>
    </Teleport>
  </div>

  <div v-else class="flex aspect-[4/3] w-full items-center justify-center bg-brick-soft text-brick/40">
    <ImageOff class="h-12 w-12" />
  </div>
</template>
