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

เปิด browser ที่ `https://localhost` (HTTP จะ redirect ไป HTTPS ให้อัตโนมัติ)

nginx จะออก self-signed certificate ให้เองตอน container เริ่มครั้งแรก (ดูหัวข้อ HTTPS ด้านล่าง) ดังนั้น browser จะเตือนว่า certificate ไม่น่าเชื่อถือ — กด "Advanced" → "Proceed" เพื่อเข้าใช้งานได้ตามปกติ

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

Swagger UI พร้อมใช้งานที่ `https://localhost/docs` หลังจากรันระบบแล้ว

## HTTPS

nginx เป็นจุด terminate TLS (listen 443) และ redirect HTTP → HTTPS ให้อัตโนมัติ (ยกเว้น `/health` ที่ยังเปิดผ่าน HTTP ไว้สำหรับ Docker healthcheck)

**ตอนนี้ (ยังไม่มี domain):** container `nginx` จะสร้าง self-signed certificate ให้เองตอนเริ่มครั้งแรก (`nginx/generate-cert.sh`) แล้วเก็บไว้ใน volume `certs` เพื่อไม่ต้องสร้างใหม่ทุกครั้งที่ deploy ตั้งค่า `SSL_CN` ใน `.env` เป็น public IP ของ server เพื่อให้ certificate ตรงกับที่อยู่จริง (ไม่บังคับ — self-signed จะโดน browser เตือนอยู่ดี)

**เมื่อมี domain แล้ว — ย้ายไปใช้ Let's Encrypt:**
1. ชี้ DNS ของ domain มาที่ server
2. ใช้ certbot (webroot mode ผ่าน location `/.well-known/acme-challenge/` ที่เตรียมไว้ใน `nginx/nginx.conf` แล้ว) ออก certificate จริง แล้ว mount `fullchain.pem`/`privkey.pem` ทับ volume `certs` แทนของที่ generate-cert.sh สร้าง (หรือแก้ script ให้เรียก certbot แทน openssl)
3. เปลี่ยน `server_name _;` ทั้งสอง server block ใน `nginx/nginx.conf` เป็นชื่อ domain จริง
4. เปิด `Strict-Transport-Security` header ที่ comment ไว้ใน `nginx/nginx.conf` (ปลอดภัยกับ certificate ที่ browser เชื่อถือแล้วเท่านั้น)
5. อัปเดต GitHub secret `CORS_ORIGINS` ให้เป็น `https://your-domain`

**Cookie:** refresh-token cookie จะถูกส่งเฉพาะผ่าน HTTPS เมื่อตั้ง `APP_COOKIE_SECURE=true` ใน `backend/.env` — deploy workflow (`.github/workflows/deploy.yml`) ตั้งค่านี้ให้อัตโนมัติสำหรับ production; สำหรับรัน `docker compose` ในเครื่อง ตั้งเองใน `backend/.env` ได้เช่นกัน (ค่า default คือ `false` เพื่อไม่ให้ไปกระทบ local dev แบบ `go run` ที่รันผ่าน HTTP ตรงๆ)

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