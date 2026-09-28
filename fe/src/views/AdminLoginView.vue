<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Handshake, LogIn } from 'lucide-vue-next'
import { useAdminStore } from '@/stores/admin'

const router = useRouter()
const store = useAdminStore()

const username = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  if (!username.value || !password.value) return
  loading.value = true
  try {
    await store.login(username.value, password.value)
    router.push({ name: 'admin-products' })
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Falha no login')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-dvh flex-col items-center justify-center px-6">
    <div class="mb-6 flex flex-col items-center gap-2">
      <span class="flex h-14 w-14 items-center justify-center rounded-full border-2 border-navy bg-white text-brick">
        <Handshake class="h-7 w-7" />
      </span>
      <h1 class="text-xl font-extrabold">
        <span class="italic text-navy">fechou o</span>
        <span class="ml-1 text-brick">brick</span>
      </h1>
      <p class="text-sm text-navy/50">Painel do vendedor</p>
    </div>

    <form class="w-full max-w-sm space-y-3" @submit.prevent="submit">
      <input
        v-model="username"
        type="text"
        placeholder="Usuário"
        autocomplete="username"
        class="w-full rounded-xl border-2 border-navy/10 bg-white px-4 py-3 font-semibold outline-none focus:border-brick"
      />
      <input
        v-model="password"
        type="password"
        placeholder="Senha"
        autocomplete="current-password"
        class="w-full rounded-xl border-2 border-navy/10 bg-white px-4 py-3 font-semibold outline-none focus:border-brick"
      />
      <button
        type="submit"
        :disabled="loading"
        class="flex w-full items-center justify-center gap-2 rounded-xl bg-brick px-4 py-3 font-extrabold text-white transition active:bg-brick-dark disabled:opacity-60"
      >
        <LogIn class="h-5 w-5" /> {{ loading ? 'Entrando...' : 'Entrar' }}
      </button>
    </form>
  </div>
</template>
