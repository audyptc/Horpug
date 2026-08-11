# Horpug Backend

REST API สำหรับระบบจัดการหอพัก Horpug พัฒนาด้วย Go + Fiber v3 + PostgreSQL (pgx)

## Requirements

- Go 1.26+
- PostgreSQL (รันอยู่แล้ว เช่นผ่าน `docker compose` ที่ root ของโปรเจกต์ หรือ instance ของตัวเอง)
- [swag CLI](https://github.com/swaggo/swag) (สำหรับ generate Swagger docs)

## Setup

1. ตั้งค่า environment variables โดย copy จากไฟล์ตัวอย่าง แล้วแก้ค่าตามเครื่องของตัวเอง (`.env` อยู่ในโฟลเดอร์ `backend/` นี้แล้ว):

   ```env
   APP_PORT=3009
   APP_SECRETKEY=SuperSecretKeyJwtToken
   ACCESS_TOKEN_TTL=6h
   REFRESH_TOKEN_TTL=6h

   DB_HOST=localhost
   DB_PORT=5432
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   DB_NAME=phorpug

   UPLOAD_DIR=./uploads
   UPLOAD_BASE_URL=http://localhost:3009
   ```

2. ติดตั้ง dependencies:

   ```bash
   go mod download
   ```

## รัน API

```bash
go run ./cmd/api
```

เมื่อรันสำเร็จ server จะฟังที่ `http://localhost:3009` (หรือ port ตามที่ตั้งไว้ใน `APP_PORT`)

ตอน start ระบบจะ auto migrate database และ seed ข้อมูล menu/permission เริ่มต้นให้อัตโนมัติ

### Build เป็น binary

```bash
go build -o main ./cmd/api
./main
```

## Swagger Docs

API ใช้ [swaggo/swag](https://github.com/swaggo/swag) generate เอกสารจาก comment annotation เหนือ handler function และ `main.go`

### ติดตั้ง swag CLI (ครั้งแรก)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

ตรวจสอบว่า `$GOPATH/bin` (หรือ `$HOME/go/bin`) อยู่ใน `PATH` แล้ว ถึงจะเรียกคำสั่ง `swag` ได้ตรง ๆ

### Generate/อัปเดต docs

รันคำสั่งนี้ที่ root ของ `backend/` ทุกครั้งที่แก้ไข annotation (`@Summary`, `@Router`, ฯลฯ) หรือเพิ่ม endpoint ใหม่:

```bash
swag init -g cmd/api/main.go -o docs
```

คำสั่งนี้จะ generate/อัปเดตไฟล์ในโฟลเดอร์ `docs/`:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

### ดู Swagger UI

หลังจากรัน API แล้ว เปิดดูเอกสารได้ที่:

- Swagger UI: `http://localhost:3009/docs/swagger`
- Scalar (API Reference): `http://localhost:3009/docs/scalar`
- Raw OpenAPI JSON: `http://localhost:3009/swagger/doc.json`

## โครงสร้างโปรเจกต์

```
backend/
├── cmd/api/           # entrypoint (main.go)
├── config/            # โหลด config จาก .env
├── docs/              # ไฟล์ที่ swag generate (ห้ามแก้ไขมือ)
├── internal/
│   ├── features/      # business logic แยกตาม feature (user, role, permission, menu, ...)
│   ├── http/          # route registration, error handler, response helpers
│   └── platform/      # infra เช่น database connection/migration
└── go.mod
```
