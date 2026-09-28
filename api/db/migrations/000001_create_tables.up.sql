CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE admin_users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username      text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug              text NOT NULL UNIQUE,
    title             text NOT NULL,
    description       text,
    price_pix         numeric(10,2) NOT NULL,
    price_card_total  numeric(10,2),
    card_installments int CHECK (card_installments BETWEEN 1 AND 12),
    whatsapp_url      text NOT NULL,
    status            text NOT NULL DEFAULT 'ATIVO'
                      CHECK (status IN ('ATIVO', 'VENDIDO', 'OCULTO')),
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

-- Listagem alfabética + keyset pagination do infinite scroll
CREATE INDEX idx_products_alpha ON products (lower(title), id) WHERE status = 'ATIVO';

CREATE TABLE product_photos (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id   uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    position     int  NOT NULL DEFAULT 0,
    content_type text NOT NULL,
    size_bytes   int  NOT NULL,
    data         bytea NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_photos_product ON product_photos (product_id, position);
