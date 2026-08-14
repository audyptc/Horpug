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
├── frontend/         # React frontend
├── backend/          # Go backend
├── nginx/            # Nginx config (reverse proxy)
├── docker-compose.yml
└── .env.example
```

## การติดตั้งและรัน

### 1. ตั้งค่า Database

PostgreSQL ไม่ได้รันผ่าน Docker — ติดตั้งบน host ตามปกติ แล้วสร้าง database/user ที่ต้องใช้ไว้ล่วงหน้า

### 2. ตั้งค่า Environment

```bash
cp .env.example .env
cp backend/.env.example backend/.env   # ถ้ายังไม่มี
```

แก้ไข `backend/.env` ให้ชี้ไปที่ database ที่ติดตั้งไว้:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=your_strong_password
DB_NAME=phorpug
```

(เมื่อรันผ่าน `docker compose`, `DB_HOST` จะถูก override เป็น `host.docker.internal` โดยอัตโนมัติเพื่อให้ container เชื่อมต่อ database บน host ได้)

### 3. รันด้วย Docker Compose

```bash
docker compose up --build -d
```

เปิด browser ที่ `http://localhost`

### 4. หยุดระบบ

```bash
docker compose down
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
cd frontend
npm install
npm run dev        # http://localhost:5173
```

### Backend

```bash
cd backend
go run ./cmd/api
```
