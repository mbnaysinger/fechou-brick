<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { Handshake } from 'lucide-vue-next'
</script>

<template>
  <header class="sticky top-0 z-20 border-b border-navy/5 bg-cream/90 backdrop-blur">
    <div class="mx-auto flex max-w-3xl items-center gap-2 px-4 py-3">
      <RouterLink to="/" class="flex items-center gap-2">
        <span class="flex h-9 w-9 items-center justify-center rounded-full border-2 border-navy bg-white text-brick">
          <Handshake class="h-5 w-5" />
        </span>
        <span class="text-lg font-extrabold leading-none">
          <span class="italic text-navy">fechou o</span>
          <span class="ml-1 text-brick">brick</span>
        </span>
      </RouterLink>
      <slot />
    </div>
  </header>
</template>
