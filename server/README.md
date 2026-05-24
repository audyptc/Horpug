# Server

## Architecture

```text
[Client] -> Fiber Router / Handlers -> Usecase -> Repository -> PostgreSQL
```

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