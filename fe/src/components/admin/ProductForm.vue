<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { ArrowLeft, ArrowRight, ImagePlus, Star, Trash2, X } from 'lucide-vue-next'
import { admin, photoUrl, type ProductPayload } from '@/lib/api'
import { compressImage } from '@/lib/image'
import type { PhotoMeta, Product } from '@/types'
import PriceBlock from '@/components/PriceBlock.vue'

const props = defineProps<{
  // null = criar novo anúncio
  productId: string | null
}>()

const emit = defineEmits<{ close: []; saved: [] }>()

const title = ref('')
const description = ref('')
const pricePix = ref<number | null>(null)
const priceCardTotal = ref<number | null>(null)
const cardInstallments = ref(1)
const whatsappUrl = ref('https://wa.me/55')

const photos = ref<PhotoMeta[]>([])
const savedId = ref<string | null>(props.productId)
const saving = ref(false)
const uploading = ref(false)
const loading = ref(false)

watch(
  () => props.productId,
  async (id) => {
    savedId.value = id
    if (!id) return
    loading.value = true
    try {
      const detail = await admin.getProduct(id)
      title.value = detail.title
      description.value = detail.description ?? ''
      pricePix.value = detail.price_pix
      priceCardTotal.value = detail.price_card_total
      cardInstallments.value = detail.card_installments ?? 1
      whatsappUrl.value = detail.whatsapp_url
      photos.value = detail.photos
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Erro ao carregar anúncio')
    } finally {
      loading.value = false
    }
  },
  { immediate: true }
)

const payload = computed<ProductPayload>(() => ({
  title: title.value,
  description: description.value.trim() || null,
  price_pix: pricePix.value ?? 0,
  price_card_total: priceCardTotal.value,
  card_installments: priceCardTotal.value ? cardInstallments.value : null,
  whatsapp_url: whatsappUrl.value.trim(),
}))

async function save() {
  saving.value = true
  try {
    if (savedId.value) {
      await admin.updateProduct(savedId.value, payload.value)
      toast.success('Anúncio atualizado!')
    } else {
      const created: Product = await admin.createProduct(payload.value)
      savedId.value = created.id
      toast.success('Anúncio criado! Agora adicione as fotos.')
    }
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao salvar')
  } finally {
    saving.value = false
  }
}

async function onFilesSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  if (!files.length || !savedId.value) return

  uploading.value = true
  try {
    const compressed = await Promise.all(files.map(compressImage))
    photos.value = await admin.uploadPhotos(savedId.value, compressed)
    toast.success(files.length === 1 ? 'Foto adicionada!' : `${files.length} fotos adicionadas!`)
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro no upload')
  } finally {
    uploading.value = false
  }
}

async function applyOrder(orderedIds: string[]) {
  if (!savedId.value) return
  try {
    photos.value = await admin.reorderPhotos(savedId.value, orderedIds)
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao reordenar')
  }
}

function makePrimary(photo: PhotoMeta) {
  const rest = photos.value.filter((p) => p.id !== photo.id).map((p) => p.id)
  applyOrder([photo.id, ...rest])
}

function move(photo: PhotoMeta, delta: number) {
  const ids = photos.value.map((p) => p.id)
  const from = ids.indexOf(photo.id)
  const to = from + delta
  if (to < 0 || to >= ids.length) return
  ids.splice(from, 1)
  ids.splice(to, 0, photo.id)
  applyOrder(ids)
}

async function removePhoto(photo: PhotoMeta) {
  try {
    await admin.deletePhoto(photo.id)
    photos.value = photos.value.filter((p) => p.id !== photo.id)
    emit('saved')
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'Erro ao remover foto')
  }
}

const inputClass =
  'w-full rounded-xl border-2 border-navy/10 bg-white px-3 py-2.5 font-semibold outline-none focus:border-brick'
</script>

<template>
  <div class="fixed inset-0 z-40 overflow-y-auto bg-cream">
    <header class="sticky top-0 z-10 flex items-center gap-3 border-b border-navy/5 bg-cream/90 px-4 py-3 backdrop-blur">
      <button type="button" class="rounded-full p-1.5 text-navy/60 active:bg-navy/10" aria-label="Fechar" @click="emit('close')">
        <X class="h-5 w-5" />
      </button>
      <h2 class="text-lg font-extrabold">{{ savedId ? 'Editar anúncio' : 'Novo anúncio' }}</h2>
    </header>

    <div v-if="loading" class="p-8 text-center text-navy/50">Carregando...</div>

    <form v-else class="mx-auto max-w-xl space-y-5 px-4 py-5 pb-16" @submit.prevent="save">
      <div>
        <label class="mb-1 block text-sm font-bold text-navy/70">Título *</label>
        <input v-model="title" type="text" required :class="inputClass" placeholder="Ex: Bicicleta aro 29" />
      </div>

      <div>
        <label class="mb-1 block text-sm font-bold text-navy/70">Descrição</label>
        <textarea v-model="description" rows="4" :class="inputClass" placeholder="Estado, detalhes, retirada..." />
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="mb-1 block text-sm font-bold text-navy/70">Valor no PIX (R$) *</label>
          <input v-model.number="pricePix" type="number" min="0.01" step="0.01" required inputmode="decimal" :class="inputClass" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-bold text-navy/70">Total no cartão (R$)</label>
          <input
            v-model.number="priceCardTotal"
            type="number"
            min="0.01"
            step="0.01"
            inputmode="decimal"
            :class="inputClass"
            placeholder="vazio = só PIX"
          />
        </div>
      </div>

      <div v-if="priceCardTotal">
        <label class="mb-1 block text-sm font-bold text-navy/70">Parcelas no cartão</label>
        <select v-model.number="cardInstallments" :class="inputClass">
          <option v-for="n in 12" :key="n" :value="n">{{ n }}x</option>
        </select>
      </div>

      <!-- Preview ao vivo: exatamente o que o comprador verá -->
      <div v-if="pricePix" class="rounded-2xl bg-white p-4 ring-1 ring-navy/10">
        <p class="mb-2 text-xs font-bold uppercase tracking-wide text-navy/40">Como o comprador verá</p>
        <PriceBlock
          :price-pix="pricePix"
          :price-card-total="priceCardTotal"
          :card-installments="priceCardTotal ? cardInstallments : null"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-bold text-navy/70">Link do WhatsApp *</label>
        <input v-model="whatsappUrl" type="url" required :class="inputClass" placeholder="https://wa.me/5551999999999" />
      </div>

      <button
        type="submit"
        :disabled="saving"
        class="w-full rounded-xl bg-brick px-4 py-3 font-extrabold text-white transition active:bg-brick-dark disabled:opacity-60"
      >
        {{ saving ? 'Salvando...' : savedId ? 'Salvar alterações' : 'Criar anúncio' }}
      </button>

      <!-- Fotos: habilitado depois que o anúncio existe -->
      <section class="border-t border-navy/10 pt-5">
        <h3 class="mb-2 font-extrabold">Fotos</h3>
        <p v-if="!savedId" class="text-sm text-navy/50">Crie o anúncio primeiro para adicionar fotos.</p>

        <template v-else>
          <div class="grid grid-cols-3 gap-2">
            <div v-for="(photo, i) in photos" :key="photo.id" class="relative overflow-hidden rounded-xl ring-1 ring-navy/10">
              <img :src="photoUrl(photo.id)" class="aspect-square w-full object-cover" :alt="`Foto ${i + 1}`" />
              <span
                v-if="i === 0"
                class="absolute left-1 top-1 flex items-center gap-0.5 rounded-full bg-amber px-1.5 py-0.5 text-[10px] font-extrabold text-white"
              >
                <Star class="h-3 w-3" /> principal
              </span>
              <div class="absolute inset-x-0 bottom-0 flex justify-center gap-1 bg-navy-dark/70 py-1">
                <button v-if="i > 0" type="button" class="rounded p-1 text-white" aria-label="Mover antes" @click="move(photo, -1)">
                  <ArrowLeft class="h-4 w-4" />
                </button>
                <button v-if="i > 0" type="button" class="rounded p-1 text-amber" aria-label="Tornar principal" @click="makePrimary(photo)">
                  <Star class="h-4 w-4" />
                </button>
                <button v-if="i < photos.length - 1" type="button" class="rounded p-1 text-white" aria-label="Mover depois" @click="move(photo, 1)">
                  <ArrowRight class="h-4 w-4" />
                </button>
                <button type="button" class="rounded p-1 text-red-300" aria-label="Remover" @click="removePhoto(photo)">
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>

            <label
              class="flex aspect-square cursor-pointer flex-col items-center justify-center gap-1 rounded-xl border-2 border-dashed border-navy/20 text-navy/50 active:bg-white"
            >
              <ImagePlus class="h-6 w-6" />
              <span class="text-xs font-bold">{{ uploading ? 'Enviando...' : 'Adicionar' }}</span>
              <input type="file" accept="image/*" multiple class="hidden" :disabled="uploading" @change="onFilesSelected" />
            </label>
          </div>
          <p class="mt-2 text-xs text-navy/40">
            As fotos são comprimidas no navegador antes do envio. A primeira é a capa do anúncio.
          </p>
        </template>
      </section>
    </form>
  </div>
</template>
