# Auth & User Modules

เอกสารนี้อธิบาย **พฤติกรรมจริงในโค้ดปัจจุบัน** ของโมดูล `internal/api/auth` และ `internal/api/user` (ไม่ใช่ design ที่ตั้งใจไว้) พร้อมส่วน [Known gaps](#known-gaps--gotchas) ที่รวบรวม bug/ช่องโหว่ที่เจอตอนรีวิว — อ่านส่วนนั้นก่อนต่อยอดโค้ด auth

## 1. Layering

ทั้งสองโมดูลใช้ pattern เดียวกัน (wire ใน `cmd/main.go`):

```
router.go → handler.go → usecase.go → repository.go
              ↑ DTO/model อยู่ใน auth.go / user.go
```

- `handler` — parse body (`c.BodyParser`), validate (`core.ValidateStruct`), เรียก usecase, ตอบผ่าน `core.SendSuccess/SendError/SendValidationError`
- `usecase` — business logic, ไม่แตะ Fiber context
- `repository` — sqlx query ตรงกับตาราง `users` (โมดูลทั้งสองใช้ตารางเดียวกัน แต่มี struct คนละตัว: `auth.UserModel` กับ `user.User`)

## 2. Auth flow

### Endpoints (`internal/api/auth/router.go`)

| Method | Path                  | Auth        | Handler        |
| ------ | --------------------- | ----------- | -------------- |
| POST   | `/api/auth/register`  | public      | `Register`     |
| POST   | `/api/auth/login`     | public      | `Login`        |
| POST   | `/api/auth/refresh`   | public      | `RefreshToken` |
| POST   | `/api/auth/logout`    | JWT         | `Logout`       |

### JWT (`pkg/core/jwt.go`)

- HS256, claims = `{ userId, email, role, exp, iat, sub }` (struct `core.Claims`)
- access + refresh แยก secret กัน (`AccessTokenSecret` / `RefreshTokenSecret`) — เป็น package-level var set จาก config ใน `main.go` ตอน startup
- `GenerateTokenPair` สร้างทั้งคู่, `ValidateAccessToken` / `ValidateRefreshToken` verify ตาม secret ของแต่ละชนิด

### Register (`usecase.go:Register`)

1. เช็คซ้ำ email + username → ถ้ามีแล้ว error
2. `common.HashPassword` (bcrypt cost 14) แล้ว `CreateUser`
3. `GenerateTokenPair` → hash refresh token เก็บลง `refresh_token_hash`
4. คืน `AuthResponse { accessToken, refreshToken, user }`

### Login (`usecase.go:Login`)

1. หา user ด้วย `FindByUsername` — ถ้าไม่เจอ fallback หาด้วย `FindByEmail` (field ใน DTO ชื่อ `username` แต่รับ email ได้ด้วย)
2. เช็ค `is_active` → เช็ค password (`common.ComparePasswords`)
3. gen token pair, hash refresh token เก็บลง DB

### Refresh (`usecase.go:RefreshToken`)

1. `ValidateRefreshToken(dto.RefreshToken)` → เอา `claims.UserID`
2. ดึง `refresh_token_hash` ที่เก็บไว้ แล้วเทียบกับ hash ของ token ที่ส่งมา
3. หา user (`FindByEmail(claims.Email)`), เช็ค `is_active`, gen token pair ใหม่, เก็บ hash ใหม่

### Logout (`usecase.go:Logout`)

- ปัจจุบัน **แค่** `ClearRefreshTokenHash(userID)` — เคลียร์ refresh hash ใน DB
- การ blacklist access token ผ่าน Redis เขียนไว้เป็น comment แต่ยังไม่ทำงาน → **access token เดิมยังใช้ได้จนกว่าจะหมดอายุ (15 นาที)** แม้ logout แล้ว

## 3. User flow

### Endpoints (`internal/api/user/router.go`)

ทั้งกลุ่ม `/api/users` อยู่หลัง `middleware.JWTMiddleware()`

| Method | Path             | Extra guard              | Handler   |
| ------ | ---------------- | ------------------------ | --------- |
| GET    | `/api/users`     | —                        | `GetAll`  |
| GET    | `/api/users/:id` | —                        | `GetByID` |
| POST   | `/api/users`     | `RequireRole("ADMIN")`   | `Create`  |
| PATCH  | `/api/users/:id` | — (ไม่มี!)               | `Update`  |
| DELETE | `/api/users/:id` | `RequireRole("ADMIN")`   | `Delete`  |

- `GetAll` ใช้ `core.PagingRequest` (default limit 0 = คืนทั้งหมด) + `core.SortingRequest` (default `created_at DESC`)
- `Update` เป็น partial update (DTO ใช้ pointer field, set เฉพาะที่ไม่ nil)
- `Delete` = soft delete (`deleted_at = NOW()`); ทุก query filter `deleted_at IS NULL`

## 4. Known gaps / gotchas

> จุดเหล่านี้เป็นพฤติกรรมจริงในโค้ดตอนนี้ ที่ควรแก้ก่อน production — บันทึกไว้เพราะไม่เห็นได้จากการอ่าน README

### 🔴 A. Register บังคับ role = ADMIN ทุกคน

`internal/api/auth/repository.go:29-30` — `INSERT INTO users (...) VALUES (..., 'ADMIN')` hardcode role เป็น `'ADMIN'` ทุกครั้ง ทำให้ **ทุกคนที่ register กลายเป็น ADMIN** (RegisterDTO ไม่มี field role ด้วยซ้ำ) ควรเปลี่ยน default เป็น `CUSTOMER` และให้ ADMIN สร้าง user ที่มี role สูงผ่าน `POST /api/users` แทน

### 🔴 B. PATCH /api/users/:id ไม่มี authz guard

`router.go` ไม่ได้ใส่ `RequireRole` และ handler/usecase ไม่ได้เช็คว่า caller เป็นเจ้าของ record หรือ ADMIN (README อ้างว่า "owner or ADMIN" แต่โค้ดไม่ได้ทำ) → **user ที่ login แล้วคนใดก็ได้ แก้ข้อมูล user อื่นได้ รวมถึงเปลี่ยน `role` ของตัวเองเป็น ADMIN** (UpdateUserDTO มี field `role`) รวมกับข้อ A แล้วเป็นช่องโหว่ยกระดับสิทธิ์เต็มรูปแบบ ต้องเพิ่ม guard: เจ้าของ record หรือ ADMIN เท่านั้น และห้าม non-ADMIN แก้ field `role`

### 🔴 C. Refresh token hashing ไม่ consistent → refresh พังบางเส้นทาง

- `Register` เก็บ refresh hash ด้วย `common.HashPassword` (bcrypt, มี salt/สุ่ม)
- `Login` เก็บด้วย `common.HashToken` (sha256, deterministic)
- `RefreshToken` **verify** ด้วยการเทียบ `HashToken(token) == storedHash` (sha256 ตรง ๆ) แล้ว re-store ด้วย `HashPassword` (bcrypt) อีก

ผลที่ตามมา:
- refresh หลัง **register** → เทียบ sha256 กับ bcrypt hash → ไม่ตรงเสมอ → `invalid refresh token`
- refresh หลัง **login** ครั้งแรก → ผ่าน (sha256 == sha256) แต่ครั้งถัดไปเก็บเป็น bcrypt → พังอีก

ต้องเลือกวิธี hash ให้เหมือนกันทั้ง 3 จุด (แนะนำ `HashToken`/sha256 เพราะ verify ต้อง deterministic — bcrypt ของ token แบบสุ่มเทียบตรง ๆ ไม่ได้อยู่แล้ว ต้องใช้ `bcrypt.CompareHashAndPassword`)

### 🟡 D. Logout ไม่ invalidate access token

ดูข้อ Logout ด้านบน — ต้องต่อ Redis blacklist (มี `gofiber/storage/redis` และ `redis/go-redis` ใน go.mod แล้ว แต่ client ถูก comment ใน `main.go`) หรือปรับ access token ให้อายุสั้นลง

### 🟡 E. SQL injection ผ่าน sort param

`user.FindAll` interpolate `sort`/`order` ลง `ORDER BY` ตรง ๆ และ `core.SortingRequest` ยังไม่ได้ทำ whitelist — ดูรายละเอียดใน [api-spec.md](./api-spec.md) จุดนี้กระทบทั้ง user module และ endpoint EMS ใหม่ที่มี sort

## 5. เกี่ยวกับ EMS docs

โมดูล EMS ใหม่ (device / telemetry ใน [api-spec.md](./api-spec.md)) จะ reuse ของเดิม:
- `middleware.JWTMiddleware()` + `middleware.RequireRole(...)` สำหรับ authz — แต่ต้องแก้ข้อ A/B ก่อน ไม่งั้น role-based guard ไม่มีความหมาย (ทุกคนเป็น ADMIN อยู่แล้ว)
- role คงเดิม `ADMIN` / `STAFF` / `CUSTOMER`
