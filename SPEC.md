# fechou o brick — Sugestão de Implementação

Plataforma mobile-first de vendas rápidas: catálogo público de anúncios com infinite scroll, página de detalhe com fotos, contato via WhatsApp e admin de usuário único para gerenciar os produtos.

**Stack** (mesma dos projetos tallo-planning): API em Go (stdlib `net/http`, arquitetura hexagonal, sqlx/pgx, golang-migrate, JWT) + Frontend Vue 3 / Vite / TypeScript / Tailwind CSS 4 / Pinia + PostgreSQL 17.

---

## 1. Estrutura de pastas

```
fechou-brick/
├── api/                          # Backend Go (espelha tallo-planning-api)
│   ├── cmd/fechou-brick-api/main.go
│   ├── db/migrations/            # golang-migrate, roda no start
│   ├── internal/
│   │   ├── adapters/
│   │   │   ├── handlers/         # product_handler, photo_handler, admin_handler, routes
│   │   │   ├── middlewares/      # auth (JWT local), cors, logger
│   │   │   └── repositories/     # postgres_product_repository, postgres_photo_repository
│   │   ├── config/db.go
│   │   └── core/
│   │       ├── domain/           # product, photo, admin_user + interfaces de repositório
│   │       └── services/         # product_service, auth_service
│   └── Dockerfile
├── fe/                           # Frontend Vue (espelha tallo-planning-fe)
│   ├── src/
│   │   ├── views/                # HomeView, ProductView, AdminLoginView, AdminProductsView
│   │   ├── components/           # ProductCard, PhotoCarousel, WhatsAppButton, ShareButton...
│   │   ├── stores/               # products.ts, admin.ts (Pinia)
│   │   ├── router/               # /, /p/:slug, /admin/login, /admin/produtos
│   │   └── lib/api.ts            # axios
│   ├── nginx.conf                # SPA + proxy /api → backend (evita CORS em produção)
│   └── Dockerfile
├── docker-compose.yml            # VPS: api + fe (banco externo, placeholder no .env)
├── docker-compose.local.yml      # Local: api + fe + postgres 17.10
├── .env.example
└── SPEC.md
```

Diferença importante vs. tallo: **sem nayz-auth/gRPC**. O auth aqui é local e mínimo (usuário único, bcrypt + JWT emitido pela própria API). Menos infraestrutura, adequado ao escopo.

---

## 2. Modelo de dados

```sql
-- 000001_create_tables.up.sql

CREATE EXTENSION IF NOT EXISTS pgcrypto;  -- gen_random_uuid

CREATE TABLE admin_users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username      text NOT NULL UNIQUE,
    password_hash text NOT NULL,              -- bcrypt (golang.org/x/crypto/bcrypt)
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug         text NOT NULL UNIQUE,        -- gerado do título, usado no link compartilhável
    title        text NOT NULL,
    description  text,
    price_pix    numeric(10,2) NOT NULL,
    price_card_total  numeric(10,2),          -- total no cartão; NULL = anúncio "somente PIX"
    card_installments int CHECK (card_installments BETWEEN 1 AND 12),
                                              -- nº de parcelas; NULL/1 = à vista no cartão
    whatsapp_url text NOT NULL,               -- ex: https://wa.me/5551999999999
    status       text NOT NULL DEFAULT 'ATIVO'
                 CHECK (status IN ('ATIVO', 'VENDIDO', 'OCULTO')),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Listagem alfabética + keyset pagination do infinite scroll
CREATE INDEX idx_products_alpha ON products (lower(title), id) WHERE status = 'ATIVO';

CREATE TABLE product_photos (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id   uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    position     int  NOT NULL DEFAULT 0,     -- 0 = foto principal
    content_type text NOT NULL,               -- image/jpeg | image/webp
    size_bytes   int  NOT NULL,
    data         bytea NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_photos_product ON product_photos (product_id, position);
```

### Fotos no PostgreSQL: por que BYTEA

Das opções nativas do Postgres (BYTEA vs. Large Objects `lo`), **BYTEA é a escolha certa aqui**:

- Large Objects só valem a pena acima de ~100 MB por arquivo (streaming parcial); trazem complexidade de API (`lo_*`), permissões próprias e necessidade de `vacuumlo` para órfãos.
- BYTEA é uma coluna comum: transacional, `ON DELETE CASCADE` limpa as fotos junto com o produto, backup/restore normal, suporte direto no pgx. O TOAST comprime e armazena fora da página automaticamente.
- Para manter as linhas pequenas, o **frontend redimensiona antes do upload** (canvas: lado maior ≤ 1600px, JPEG/WebP qualidade ~0.8 → ~200–500 KB por foto). A API ainda valida um teto de 10 MB por arquivo.

A tabela de fotos é separada de `products` de propósito: a listagem pública nunca carrega os bytes — só monta a URL da foto principal. Os bytes só saem do banco no endpoint dedicado de foto, que responde com `Cache-Control: public, max-age=31536000, immutable` + `ETag` (o id da foto é imutável), então cada foto é baixada uma única vez pelo navegador.

Se um dia o volume crescer, a migração para disco/S3 fica isolada no repositório de fotos — nada muda nos handlers ou no FE.

---

## 3. API — Endpoints

### Público (sem auth)

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Healthcheck |
| GET | `/api/v1/produtos?limit=20&cursor=...` | Lista `ATIVO` em ordem alfabética; keyset pagination |
| GET | `/api/v1/produtos/{slug}` | Detalhe completo (inclui lista de ids de fotos) |
| GET | `/api/v1/fotos/{id}` | Bytes da foto (Content-Type + cache imutável) |

**Infinite scroll (keyset, não offset):** o cursor codifica `(lower(title), id)` do último item em base64. Query: `WHERE status='ATIVO' AND (lower(title), id) > ($1, $2) ORDER BY lower(title), id LIMIT $3`. Estável mesmo com inserções/remoções entre páginas — offset quebraria.

Resposta da listagem (o card só precisa disso):

```json
{
  "items": [
    {
      "slug": "bicicleta-aro-29",
      "title": "Bicicleta aro 29",
      "price_pix": 850.00,
      "price_card_total": 1000.00,
      "card_installments": 4,
      "status": "ATIVO",
      "photo_url": "/api/v1/fotos/3f2a..."
    }
  ],
  "next_cursor": "YmljaWNsZXRhIGFybyAyOXwzZjJh..."
}
```

### Admin (JWT via `Authorization: Bearer`)

| Método | Rota | Descrição |
|---|---|---|
| POST | `/api/v1/admin/login` | `{username, password}` → `{token}` (JWT HS256, exp 24h) |
| GET | `/api/v1/admin/produtos` | Lista todos (inclusive VENDIDO/OCULTO) |
| POST | `/api/v1/admin/produtos` | Cria anúncio (título, descrição, preço PIX, cartão total + parcelas, whatsapp_url) |
| PUT | `/api/v1/admin/produtos/{id}` | Atualiza infos |
| PATCH | `/api/v1/admin/produtos/{id}/status` | ATIVO / VENDIDO / OCULTO |
| DELETE | `/api/v1/admin/produtos/{id}` | Remove (cascade apaga fotos) |
| POST | `/api/v1/admin/produtos/{id}/fotos` | Upload multipart (1..n arquivos) |
| PUT | `/api/v1/admin/produtos/{id}/fotos/ordem` | Reordena (define a principal = position 0) |
| DELETE | `/api/v1/admin/fotos/{foto_id}` | Remove foto |

### Auth de usuário único

- **Seed no startup**: se `admin_users` estiver vazia, a API cria o usuário com `ADMIN_USERNAME` + hash bcrypt de `ADMIN_INITIAL_PASSWORD` (env). Depois do primeiro start a env pode ser removida; trocar a senha = atualizar a linha (endpoint opcional `PATCH /api/v1/admin/senha`).
- Login compara bcrypt e emite JWT assinado com `JWT_SECRET` (mesmo padrão de middleware do tallo, sem roles — existe um perfil só).
- Delay fixo de ~300ms na resposta de login errado (mitiga força bruta sem precisar de rate-limit elaborado).

Envs da API: `DATABASE_URL` (obrigatória), `JWT_SECRET` (obrigatória), `ADMIN_USERNAME` (default `admin`), `ADMIN_INITIAL_PASSWORD` (primeira subida), `PORT` (default `8085`), `MAX_UPLOAD_MB` (default `10`).

---

## 4. Frontend — mobile first

### Rotas

| Rota | View | Descrição |
|---|---|---|
| `/` | HomeView | Link geral: grid de cards, infinite scroll |
| `/p/:slug` | ProductView | Link único do anúncio: fotos + infos + WhatsApp + compartilhar |
| `/admin/login` | AdminLoginView | Form simples; token no localStorage (Pinia) |
| `/admin/produtos` | AdminProductsView | CRUD, upload/reordenação de fotos, status |

### Identidade visual (a partir do logo)

Tailwind 4 via `@theme` no CSS:

```css
@theme {
  --color-brick: #F05A28;        /* laranja "brick" — primário, CTAs, preço PIX */
  --color-brick-dark: #D9481C;   /* hover/pressed */
  --color-navy: #1F5673;         /* azul do traço — títulos, header */
  --color-amber: #F5A623;        /* detalhes/estrelas — badges, destaques */
  --color-cream: #FAF6F1;        /* fundo off-white quente */
  --color-wa: #25D366;           /* botão WhatsApp */
}
```

Fundo `cream`, header enxuto com o logo, tipografia arredondada (ex.: Nunito), cantos bem arredondados e sombras suaves — mesma vibe amigável do logo.

### Bloco de preço (componente `PriceBlock`, usado no card e no detalhe)

A exibição é **derivada** dos dois campos do admin (`price_card_total` + `card_installments`) — nenhum texto digitado à mão, zero chance de preço e texto divergirem. Regras:

| Dados do anúncio | Linha do cartão exibida |
|---|---|
| `price_card_total = NULL` | badge discreto "somente no PIX" |
| parcelas `NULL` ou `1` | `ou R$ 1.000,00 no cartão` |
| parcelas > 1 e total **=** PIX | `ou 4x de R$ 222,50 sem juros` |
| parcelas > 1 e total **>** PIX | `ou 4x de R$ 250,00 no cartão · total R$ 1.000,00` |

O valor da parcela é calculado (`total / parcelas`, 2 casas). O selo "sem juros" em `amber` só aparece quando é matematicamente verdade (total igual ao PIX); quando há juros, o total vem junto — transparência sem poluir. Visual:

```
┌─────────────────────────────────────┐
│  R$ 890,00              ⚡ no PIX   │   ← 2xl/bold, cor brick
│  ou 4x de R$ 250,00 no cartão       │   ← sm, cinza (navy/60)
│     total R$ 1.000,00               │   ← xs, cinza mais claro
└─────────────────────────────────────┘
```

No card da Home o bloco aparece compacto (PIX + uma linha de cartão com ellipsis); no detalhe, completo.

### Home (link geral)

- Cards em coluna única no mobile (2 colunas ≥ `sm`): **foto principal** (aspect 4:3, `object-cover`, lazy), **título curto** (1–2 linhas com ellipsis), **`PriceBlock` compacto** (PIX grande em `brick`, cartão menor em cinza).
- Infinite scroll com `IntersectionObserver` (sentinela no fim da lista) + `useInfiniteScroll` do `@vueuse/core` já disponível na stack. Skeleton cards durante o fetch.
- Produto `VENDIDO` some da listagem pública; no link direto aparece com faixa "Vendido 🤝".

### Página do produto (link único)

- Carrossel de fotos com scroll-snap horizontal (nativo, sem lib) + indicador de posição; tap abre a foto em tela cheia.
- Título, `PriceBlock` completo, descrição.
- **Barra fixa no rodapé (sticky)** com os dois botões — sempre à vista no mobile:
  - **WhatsApp** (verde `wa`): abre `"{whatsapp_url}?text=" + encodeURIComponent("Olá! Vi o anúncio \"{título}\" no fechou o brick e tenho interesse: {url do anúncio}")`.
  - **Compartilhar**: `navigator.share({title, url})` (nativo no mobile); fallback desktop copia o link + toast via `vue-sonner`.
- Meta tags OG (título, foto principal, preço) para o link ficar bonito quando compartilhado no próprio WhatsApp.

### Admin

- Mesma base visual, direto ao ponto: lista com miniatura/status/preços, form de criar/editar, upload com preview.
- Form de preço: valor PIX (obrigatório), valor **total** no cartão (opcional — vazio = somente PIX) e select de parcelas (1x–12x). Abaixo dos campos, um **preview ao vivo do `PriceBlock`** exatamente como o comprador verá — o admin ajusta os números até o texto ficar do jeito que quer.
- **Compressão client-side antes do upload**: canvas redimensiona para ≤ 1600px e exporta JPEG ~0.8 (função pequena em `lib/image.ts`, sem dependência nova).
- Guard de rota: sem token válido → redirect `/admin/login`. 401 da API → limpa token e redirect.

### Integração com a API

Em produção o nginx do FE faz proxy de `/api/` para o container da API — **mesma origem, zero CORS, um único portão exposto na VPS**. Em dev local, o `vite.config.ts` proxia `/api` para `http://localhost:8085`. Nenhuma env `VITE_*` obrigatória.

---

## 5. Deploy

### VPS (`docker-compose.yml`)

- Serviços `api` + `fe`. O banco é o **Postgres existente na VPS**: preencha `DATABASE_URL` no `.env` (placeholder pronto). Se o Postgres roda no host (fora de container), use `host.docker.internal` — o compose já mapeia via `extra_hosts`.
- Só o FE publica porta (`FE_PORT`, default `8084`); a API fica na rede interna do compose, acessada pelo proxy do nginx.
- Migrations rodam automaticamente no start da API (golang-migrate), igual ao tallo.
- Crie o database antes da primeira subida: `CREATE DATABASE fechou_brick;`

```bash
cp .env.example .env   # preencher DATABASE_URL, JWT_SECRET, ADMIN_INITIAL_PASSWORD
docker compose up -d --build
```

### Caddy (reverse proxy da VPS)

Como o nginx do FE já proxia `/api` internamente, o Caddy precisa de **um handle só** — mais simples que o padrão de dois handles usado nos outros apps:

```caddyfile
brick.seudominio.com.br {
    reverse_proxy localhost:8084
}
```

O HTTPS automático do Caddy resolve o requisito de contexto seguro da Web Share API (botão compartilhar nativo no mobile).

Se preferir manter o padrão de dois handles (API exposta separadamente), descomente a publicação da porta `8085` no serviço `api` do compose e use:

```caddyfile
brick.seudominio.com.br {
    handle /api/* {
        reverse_proxy localhost:8085
    }
    handle {
        reverse_proxy localhost:8084
    }
}
```

### Local (`docker-compose.local.yml`)

Stack completa com `postgres:17.10-alpine` (volume nomeado, porta 5433 no host para não conflitar com um Postgres local):

```bash
docker compose -f docker-compose.local.yml up --build
# FE: http://localhost:8084 | API: http://localhost:8085 | PG: localhost:5433
```

---

## 6. Fases de implementação

1. **F1 — API**: migrations, domain/repos/services, endpoints públicos com keyset pagination, endpoint de fotos com cache, auth (seed + login + middleware), CRUD admin, upload multipart. Testes dos services (padrão tallo).
2. **F2 — FE público**: scaffold Vite + Tailwind 4 com o tema, Home com infinite scroll, página do produto com carrossel, botões WhatsApp/compartilhar, OG tags.
3. **F3 — FE admin**: login, CRUD, upload com compressão client-side e reordenação de fotos.
4. **F4 — Deploy**: build das imagens, subida na VPS com o `.env` real, teste ponta a ponta pelo celular.

---

## 7. Decisões fechadas

1. **Proxy/HTTPS**: Caddy em container na VPS, com bloco pronto na seção 5 (um handle só, já que o nginx do FE proxia `/api`; alternativa de dois handles documentada). FE na porta `8084`.
2. **WhatsApp**: `whatsapp_url` por anúncio, como modelado.
3. **Parcelamento**: campos `price_card_total` + `card_installments` no anúncio; a exibição (parcela calculada, "sem juros" ou "total R$ X") é derivada automaticamente — regras e visual na seção 4 (`PriceBlock`), com preview ao vivo no form do admin.


