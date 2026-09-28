import axios from 'axios'
import type { PhotoMeta, Product, ProductDetail, ProductListItem, PublicList } from '@/types'

// Same-origin sempre: em produção o nginx proxia /api; em dev, o vite (ver vite.config.ts)
const BASE = '/api/v1'

const TOKEN_KEY = 'brick_token'

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (token: string) => localStorage.setItem(TOKEN_KEY, token)
export const clearToken = () => localStorage.removeItem(TOKEN_KEY)

export class ApiError extends Error {
  constructor(
    public override message: string,
    public status?: number
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function handleError(error: unknown): never {
  if (axios.isAxiosError(error)) {
    const message = (error.response?.data as { error?: string } | undefined)?.error || error.message
    throw new ApiError(message, error.response?.status)
  }
  throw new ApiError('Erro desconhecido de conexão')
}

const client = axios.create({ baseURL: BASE })

client.interceptors.request.use((config) => {
  const token = getToken()
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401 && !error.config?.url?.includes('/admin/login')) {
      clearToken()
      window.dispatchEvent(new Event('brick:unauthorized'))
    }
    return Promise.reject(error)
  }
)

export const photoUrl = (photoId: string) => `${BASE}/fotos/${photoId}`

// ---------------------------------------------------------------------------
// Público
// ---------------------------------------------------------------------------

export const catalog = {
  async list(cursor?: string | null, limit = 20) {
    try {
      const params = new URLSearchParams({ limit: String(limit) })
      if (cursor) params.set('cursor', cursor)
      const { data } = await client.get(`/produtos?${params.toString()}`)
      return data as PublicList
    } catch (error) {
      handleError(error)
    }
  },

  async getBySlug(slug: string) {
    try {
      const { data } = await client.get(`/produtos/${encodeURIComponent(slug)}`)
      return data as ProductDetail
    } catch (error) {
      handleError(error)
    }
  },
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

export interface ProductPayload {
  title: string
  description: string | null
  price_pix: number
  price_card_total: number | null
  card_installments: number | null
  whatsapp_url: string
}

export const admin = {
  async login(username: string, password: string) {
    try {
      const { data } = await client.post('/admin/login', { username, password })
      return data as { token: string }
    } catch (error) {
      handleError(error)
    }
  },

  async listProducts() {
    try {
      const { data } = await client.get('/admin/produtos')
      return data as ProductListItem[]
    } catch (error) {
      handleError(error)
    }
  },

  async getProduct(id: string) {
    try {
      const { data } = await client.get(`/admin/produtos/${id}`)
      return data as ProductDetail
    } catch (error) {
      handleError(error)
    }
  },

  async createProduct(payload: ProductPayload) {
    try {
      const { data } = await client.post('/admin/produtos', payload)
      return data as Product
    } catch (error) {
      handleError(error)
    }
  },

  async updateProduct(id: string, payload: ProductPayload) {
    try {
      const { data } = await client.put(`/admin/produtos/${id}`, payload)
      return data as Product
    } catch (error) {
      handleError(error)
    }
  },

  async updateStatus(id: string, status: string) {
    try {
      const { data } = await client.patch(`/admin/produtos/${id}/status`, { status })
      return data as Product
    } catch (error) {
      handleError(error)
    }
  },

  async deleteProduct(id: string) {
    try {
      await client.delete(`/admin/produtos/${id}`)
    } catch (error) {
      handleError(error)
    }
  },

  async uploadPhotos(productId: string, files: Blob[]) {
    try {
      const form = new FormData()
      files.forEach((f, i) => form.append('fotos', f, `foto-${i}.jpg`))
      const { data } = await client.post(`/admin/produtos/${productId}/fotos`, form)
      return data as PhotoMeta[]
    } catch (error) {
      handleError(error)
    }
  },

  async reorderPhotos(productId: string, orderedIds: string[]) {
    try {
      const { data } = await client.put(`/admin/produtos/${productId}/fotos/ordem`, { ordered_ids: orderedIds })
      return data as PhotoMeta[]
    } catch (error) {
      handleError(error)
    }
  },

  async deletePhoto(photoId: string) {
    try {
      await client.delete(`/admin/fotos/${photoId}`)
    } catch (error) {
      handleError(error)
    }
  },

  async changePassword(currentPassword: string, newPassword: string) {
    try {
      await client.patch('/admin/senha', { current_password: currentPassword, new_password: newPassword })
    } catch (error) {
      handleError(error)
    }
  },
}
