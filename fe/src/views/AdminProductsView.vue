<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { ExternalLink, ImageOff, LogOut, Pencil, Plus, Trash2 } from 'lucide-vue-next'
import { admin, photoUrl } from '@/lib/api'
import { formatBRL } from '@/lib/money'
import { useAdminStore } from '@/stores/admin'
import type { ProductListItem, ProductStatus } from '@/types'
import BrandHeader from '@/components/BrandHeader.vue'
import ProductForm from '@/components/admin/ProductForm.vue'

const router = useRouter()
const store = useAdminStore()

const products = ref<ProductListItem[]>([])
const loading = ref(true)
const formOpen = ref(false)
const editingId = ref<string | null>(null)

const statusStyle: Record<ProductStatus, string> = {
  ATIVO: 'bg-wa/15 text-wa-dark',
  VENDIDO: 'bg-amber/20 text-amber',
  OCULTO: 'bg-navy/10 text-navy/50',
}

async function load() {
  loading.value = true
  try {
    products.value = await admin.listProducts()
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao carregar anúncios')
  } finally {
    loading.value = false
  }
}

onMounted(load)

window.addEventListener('brick:unauthorized', () => router.push({ name: 'admin-login' }))

function openNew() {
  editingId.value = null
  formOpen.value = true
}

function openEdit(p: ProductListItem) {
  editingId.value = p.id
  formOpen.value = true
}

async function changeStatus(p: ProductListItem, event: Event) {
  const status = (event.target as HTMLSelectElement).value as ProductStatus
  try {
    await admin.updateStatus(p.id, status)
    p.status = status
    toast.success(status === 'VENDIDO' ? 'Fechou! 🤝' : 'Status atualizado')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao mudar status')
    ;(event.target as HTMLSelectElement).value = p.status
  }
}

async function remove(p: ProductListItem) {
  if (!window.confirm(`Excluir "${p.title}"? As fotos também serão apagadas.`)) return
  try {
    await admin.deleteProduct(p.id)
    products.value = products.value.filter((x) => x.id !== p.id)
    toast.success('Anúncio excluído')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao excluir')
  }
}

function logout() {
  store.logout()
  router.push({ name: 'admin-login' })
}
</script>

<template>
  <div class="min-h-dvh pb-24">
    <BrandHeader>
      <button type="button" class="ml-auto flex items-center gap-1 text-sm font-semibold text-navy/60" @click="logout">
        <LogOut class="h-4 w-4" /> sair
      </button>
    </BrandHeader>

    <main class="mx-auto max-w-3xl px-4 py-4">
      <h1 class="mb-4 text-lg font-extrabold">Meus anúncios</h1>

      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="h-24 animate-pulse rounded-2xl bg-navy/10" />
      </div>

      <p v-else-if="!products.length" class="py-16 text-center text-navy/50">
        Nenhum anúncio ainda. Crie o primeiro no botão abaixo!
      </p>

      <div v-else class="space-y-3">
        <div v-for="p in products" :key="p.id" class="flex gap-3 rounded-2xl bg-white p-3 shadow-sm ring-1 ring-navy/5">
          <div class="h-20 w-20 flex-none overflow-hidden rounded-xl bg-brick-soft">
            <img v-if="p.primary_photo_id" :src="photoUrl(p.primary_photo_id)" :alt="p.title" class="h-full w-full object-cover" />
            <div v-else class="flex h-full w-full items-center justify-center text-brick/40"><ImageOff class="h-6 w-6" /></div>
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex items-start justify-between gap-2">
              <h2 class="truncate font-bold">{{ p.title }}</h2>
              <a
                :href="`/p/${p.slug}`"
                target="_blank"
                rel="noopener"
                class="flex-none text-navy/40"
                aria-label="Abrir página pública"
              >
                <ExternalLink class="h-4 w-4" />
              </a>
            </div>
            <p class="text-sm font-extrabold text-brick">{{ formatBRL(p.price_pix) }} <span class="text-xs font-semibold text-navy/40">no PIX</span></p>

            <div class="mt-1.5 flex items-center gap-2">
              <select
                :value="p.status"
                class="rounded-full px-2 py-1 text-xs font-extrabold outline-none"
                :class="statusStyle[p.status]"
                @change="changeStatus(p, $event)"
              >
                <option value="ATIVO">ATIVO</option>
                <option value="VENDIDO">VENDIDO</option>
                <option value="OCULTO">OCULTO</option>
              </select>
              <button type="button" class="rounded-full p-1.5 text-navy/50 active:bg-navy/10" aria-label="Editar" @click="openEdit(p)">
                <Pencil class="h-4 w-4" />
              </button>
              <button type="button" class="rounded-full p-1.5 text-brick-dark/60 active:bg-brick-soft" aria-label="Excluir" @click="remove(p)">
                <Trash2 class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <button
      type="button"
      class="fixed bottom-5 right-5 z-30 flex items-center gap-2 rounded-full bg-brick px-5 py-3.5 font-extrabold text-white shadow-lg shadow-brick/30 transition active:bg-brick-dark"
      @click="openNew"
    >
      <Plus class="h-5 w-5" /> Novo anúncio
    </button>

    <ProductForm v-if="formOpen" :product-id="editingId" @close="formOpen = false; load()" @saved="load" />
  </div>
</template>
