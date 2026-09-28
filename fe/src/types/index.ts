export type ProductStatus = 'ATIVO' | 'VENDIDO' | 'OCULTO'

export interface Product {
  id: string
  slug: string
  title: string
  description: string | null
  price_pix: number
  price_card_total: number | null
  card_installments: number | null
  whatsapp_url: string
  status: ProductStatus
  created_at: string
  updated_at: string
}

export interface ProductListItem extends Product {
  primary_photo_id: string | null
}

export interface PhotoMeta {
  id: string
  product_id: string
  position: number
  content_type: string
  size_bytes: number
  created_at: string
}

export interface ProductDetail extends Product {
  photos: PhotoMeta[]
}

export interface PublicList {
  items: ProductListItem[]
  next_cursor: string | null
}
