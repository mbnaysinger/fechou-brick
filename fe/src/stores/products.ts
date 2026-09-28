import { defineStore } from 'pinia'
import { ref } from 'vue'
import { catalog } from '@/lib/api'
import type { ProductListItem } from '@/types'

export const useProductsStore = defineStore('products', () => {
  const items = ref<ProductListItem[]>([])
  const nextCursor = ref<string | null>(null)
  const loading = ref(false)
  const finished = ref(false)
  const error = ref<string | null>(null)

  async function fetchMore() {
    if (loading.value || finished.value) return
    loading.value = true
    error.value = null
    try {
      const page = await catalog.list(nextCursor.value)
      items.value.push(...page.items)
      nextCursor.value = page.next_cursor
      if (!page.next_cursor) finished.value = true
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Erro ao carregar anúncios'
    } finally {
      loading.value = false
    }
  }

  function reset() {
    items.value = []
    nextCursor.value = null
    finished.value = false
    error.value = null
  }

  return { items, nextCursor, loading, finished, error, fetchMore, reset }
})
