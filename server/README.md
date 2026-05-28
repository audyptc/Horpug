# Server

## Architecture

```text
[Client] -> Fiber Router / Handlers -> Usecase -> Repository -> PostgreSQL
```

## Auth Token Config

รองรับการตั้งค่าอายุ token ผ่าน environment variables:

- `ACCESS_TOKEN_TTL` ค่า default `15m`
- `REFRESH_TOKEN_TTL` ค่า default `168h`

รูปแบบค่าตาม Go duration เช่น `30s`, `15m`, `1h`, `24h`.

## Project Structure

```text
server/
├── cmd/
│   └── api/
│       └── main.go         # จุดเริ่มต้นของแอปพลิเคชัน
├── config/
│   └── config.go           # จัดการ configuration และ environment variables
├── internal/               # โค้ดหลักของระบบ
│   ├── database/
│   │   └── postgres.go     # เชื่อมต่อ PostgreSQL
│   ├── delivery/
│   │   └── http/
│   │       ├── router.go   # กำหนด routes
│   │       └── v1/
│   │           └── user_handler.go
│   ├── domain/
│   │   └── user.go         # model และ interface ของ user
│   ├── repository/
│   │   └── user_repo.go    # ติดต่อฐานข้อมูล
│   └── usecase/
│       └── user_usecase.go # business logic
├── .env                    # ค่าตัวแปรแวดล้อม
├── go.mod
├── go.sum
└── README.md
```