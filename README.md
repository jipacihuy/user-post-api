# User Post API

RESTful API untuk manajemen Users dan Posts menggunakan **Go**, **PostgreSQL**, **Redis**, dan **JWT Authentication**.

## ✨ Fitur

- **Autentikasi JWT** — Register & Login dengan token-based auth
- **Manajemen Users** — CRUD users (get all, get by ID, create, update, delete)
- **Manajemen Posts** — CRUD posts (get all, get by ID, create, delete)
- **Caching dengan Redis** — Cache data untuk performa lebih cepat
- **Validasi Input** — Menggunakan `go-playground/validator`
- **Password Hashing** — Menggunakan `bcrypt`
- **Docker Support** — Mudah dijalankan dengan Docker Compose

## 🛠️ Tech Stack

| Teknologi | Kegunaan |
|-----------|----------|
| **Go 1.26.5** | Bahasa pemrograman utama |
| **Gorilla Mux** | HTTP router |
| **PostgreSQL** | Database utama |
| **Redis** | Caching |
| **JWT (golang-jwt)** | Autentikasi token |
| **bcrypt** | Hashing password |
| **go-playground/validator** | Validasi input |
| **Docker** | Containerization |

## 📁 Struktur Proyek

```
user-post-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── config/
│   │   └── config.go            # Koneksi DB & Redis
│   ├── handler/
│   │   ├── post_handler.go      # HTTP handler untuk posts
│   │   └── user_handler.go      # HTTP handler untuk users
│   ├── middleware/
│   │   └── auth.go              # Middleware JWT auth
│   ├── model/
│   │   └── model.go             # Struct User, Post, dll
│   ├── redis/
│   │   └── redis.go             # Service caching Redis
│   ├── repository/
│   │   ├── post_repository.go   # Query database posts
│   │   └── user_repository.go   # Query database users
│   ├── service/
│   │   ├── post_service.go      # Business logic posts
│   │   └── user_service.go      # Business logic users
│   └── utils/
│       └── response.go          # Helper response JSON
├── pkg/
│   └── jwt/
│       └── jwt.go               # Generate & verify JWT token
├── .env                         # Environment variables (jangan di-commit)
├── .env.example                 # Template env variables (copy lalu isi)
├── .gitignore
├── docker-compose.yml           # Orchestrasi container
├── dockerfile                   # Build image Go
├── go.mod
├── go.sum
└── test.rest                    # Testing endpoints (VS Code REST Client)
```

## 🚀 Cara Menjalankan

### Prasyarat

- Go 1.26.5+
- PostgreSQL
- Redis (opsional, cache tetap berjalan tanpa Redis)
- Docker & Docker Compose (opsional)

### 1. Tanpa Docker (Langsung)

```bash
# Clone repo
git clone https://github.com/<username>/user-post-api.git
cd user-post-api

# Copy & edit environment variables
cp .env.example .env
# Sesuaikan isi .env dengan konfigurasi lokal kamu

# Jalankan aplikasi
go run cmd/api/main.go
```

### 2. Dengan Docker Compose (Rekomendasi ✅)

```bash
# Copy & edit environment variables
cp .env.example .env

# Jalankan semua service (postgres, redis, api)
docker-compose up -d

# Cek logs
docker-compose logs -f api
```

Aplikasi akan berjalan di **http://localhost:8080**

### Setup Database

Buat tabel di PostgreSQL:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
```

## 📬 API Endpoints

### Public Routes (Tanpa Token)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `POST` | `/register` | Registrasi user baru |
| `POST` | `/login` | Login dan dapatkan token JWT |
| `GET` | `/posts` | Lihat semua posts (public) |
| `GET` | `/posts/{id}` | Lihat detail post (public) |

### Protected Routes (Perlu Bearer Token)

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `GET` | `/users` | Lihat semua users |
| `GET` | `/users/{id}` | Lihat detail user |
| `POST` | `/posts` | Buat post baru |
| `DELETE` | `/posts/{id}` | Hapus post |
| `PUT` | `/users/{id}` | Update user |
| `DELETE` | `/users/{id}` | Hapus user |

### Contoh Request

**Register:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "email": "john@test.com", "password": "rahasia123"}'
```

**Login:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@test.com", "password": "rahasia123"}'
```

**Akses Protected Route:**
```bash
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer <token-dari-login>"
```

### Testing dengan REST Client

File `test.rest` sudah berisi semua endpoint. Bisa langsung di-test menggunakan VS Code extension **REST Client**.

## 🐳 Docker

```bash
# Build & jalankan
docker-compose up -d

# Hentikan
docker-compose down

# Hapus volume (data postgres hilang)
docker-compose down -v

# Lihat logs
docker-compose logs -f api

# Build ulang image
docker-compose up -d --build
```

## ⚙️ Environment Variables

Semua konfigurasi melalui file `.env`:

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `POSTGRES_USER` | `postgres` | User database |
| `POSTGRES_PASSWORD` | — | Password database |
| `POSTGRES_DB` | `user_post_api` | Nama database |
| `POSTGRES_PORT` | `5432` | Port PostgreSQL |
| `REDIS_PORT` | `6379` | Port Redis |
| `API_PORT` | `8080` | Port aplikasi |
| `JWT_SECRET` | — | Secret key untuk JWT |
| `DATABASE_URL` | — | URL koneksi DB (localhost) |
| `DATABASE_URL_DOCKER` | — | URL koneksi DB (Docker) |
| `REDIS_URL` | `redis://redis:6379` | URL koneksi Redis |

## 📄 Lisensi

[MIT](license.txt)
