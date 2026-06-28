# Horpug — ระบบจัดการหอพัก

ระบบจัดการหอพักครบวงจร พัฒนาด้วย Go (Fiber) + React + PostgreSQL รันผ่าน Docker

## Tech Stack

| Layer    | Technology                           |
|----------|--------------------------------------|
| Frontend | React 19, TypeScript, Vite, Radix UI |
| Backend  | Go 1.26, Fiber v3, JWT               |
| Database | PostgreSQL 17                        |
| Proxy    | Nginx (reverse proxy + static files) |

## โครงสร้างโปรเจกต์

```
Horpug/
├── client/           # React frontend
├── server/           # Go backend
├── nginx/            # Nginx config (reverse proxy)
├── docker-compose.yml
└── .env.example
```

## การติดตั้งและรัน

### 1. ตั้งค่า Environment

```bash
cp .env.example .env
```

แก้ไข `.env` ตั้งรหัสผ่านที่ต้องการ:

```env
POSTGRES_PASSWORD=your_strong_password
CORS_ORIGINS=http://localhost
```

### 2. รันด้วย Docker Compose

```bash
docker compose up --build -d
```

เปิด browser ที่ `http://localhost`

### 3. หยุดระบบ

```bash
# หยุด containers
docker compose down

# หยุดและลบ volume (ข้อมูล DB จะหายด้วย)
docker compose down -v
```

## อัปเดต Code

```bash
# build เฉพาะ service ที่เปลี่ยน
docker compose build --no-cache frontend
docker compose build --no-cache backend
docker compose build --no-cache backend frontend

# รีสตาร์ท
docker compose up -d
```

## Scaling Backend

Nginx จะ round-robin requests ไปยัง backend instances โดยอัตโนมัติ

```bash
docker compose up --scale backend=3 -d
```

## API Docs

Swagger UI พร้อมใช้งานที่ `http://localhost/docs` หลังจากรันระบบแล้ว

## Development (Local)

### Frontend

```bash
cd client
npm install
npm run dev        # http://localhost:5173
```

### Backend

```bash
cd server
go run ./cmd/api
```

