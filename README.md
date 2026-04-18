# Solekita

SaaS manajemen laundry sepatu untuk UMKM Indonesia.

---

## Tech Stack

| Komponen     | Teknologi                        |
| ------------ | -------------------------------- |
| Backend API  | Golang + Gin                     |
| Web Dashboard | Next.js 16 + Tailwind CSS       |
| Mobile App   | Flutter (Android)                |
| Database     | PostgreSQL 16                    |
| File Storage | Cloudflare R2                    |

---

## Prasyarat

Pastikan tools berikut sudah terinstall:

- [Go 1.26+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) + [pnpm](https://pnpm.io/)
- [Flutter 3+](https://docs.flutter.dev/get-started/install)
- [Docker + Docker Compose](https://docs.docker.com/get-docker/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [Air](https://github.com/air-verse/air) — hot reload untuk Go

---

## Setup Development Lokal

### 1. Clone repository

```bash
git clone https://github.com/stefanuspet/solekita.git
cd solekita
```

### 2. Jalankan PostgreSQL

```bash
cd docker
cp .env.example .env        # isi DB_PASSWORD
docker compose up db -d
```

### 3. Setup Backend

```bash
cd backend
cp .env.example .env        # isi semua env var yang dibutuhkan
go mod tidy
```

Generate JWT secret:

```bash
openssl rand -hex 32    # pakai output untuk JWT_SECRET
openssl rand -hex 32    # pakai output untuk JWT_REFRESH_SECRET
```

Jalankan migration:

```bash
make migrate-up
```

Jalankan backend dengan hot reload:

```bash
make dev-be
```

Backend berjalan di `http://localhost:8080`.

### 4. Setup Web Dashboard

```bash
cd web
cp .env.local.example .env.local    # isi env var
pnpm install
pnpm dev
```

Web berjalan di `http://localhost:3000`.

### 5. Verifikasi

```bash
curl http://localhost:8080/v1/health
# Expected: {"success":true,"message":"OK","data":{"status":"healthy"}}
```

---

## Struktur Repository

```
solekita/
├── backend/        # Golang + Gin REST API
├── web/            # Next.js 16 Web Dashboard
├── mobile/         # Flutter Android App
├── docker/         # Docker Compose configs
└── docs/           # Dokumentasi produk & teknis
```

---

## Environment Variables

| File                  | Keterangan                        |
| --------------------- | --------------------------------- |
| `backend/.env`        | Konfigurasi backend               |
| `web/.env.local`      | Konfigurasi web dashboard         |
| `docker/.env`         | Password database untuk Docker    |

Jangan pernah commit file `.env` ke Git.

---

## Useful Commands

```bash
# Backend
make dev-be         # jalankan backend dengan hot reload
make migrate-up     # jalankan semua migration
make migrate-down   # rollback 1 migration
make migrate-reset  # rollback semua lalu migrate up
make build-be       # build binary backend
make tidy           # go mod tidy
make lint           # jalankan golangci-lint

# Web
make dev-web        # jalankan web dashboard
make build-web      # build Next.js production

# Database
make dev-db         # jalankan PostgreSQL via Docker
```

---

## API

| Environment | URL                          |
| ----------- | ---------------------------- |
| Production  | `https://api.solekita.id/v1` |
| Development | `http://localhost:8080/v1`   |
