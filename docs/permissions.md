# Permissions (Authorization Model)

ระบบตอนนี้มีแค่ **role-based** ล้วน ๆ (`middleware.RequireRole("ADMIN")` อ่าน `c.Locals("role")`) ซึ่งหยาบเกินไปสำหรับ EMS ที่มีหลาย resource (device, telemetry) และหลายระดับการเข้าถึง เอกสารนี้ออกแบบชั้น **permission (scope)** แบบ `resource:action` มาคั่นระหว่าง role กับ endpoint

> ต้องแก้ [Known gaps A/B ใน auth-user.md](./auth-user.md#known-gaps--gotchas) ก่อน — ตอนนี้ทุกคน register เป็น ADMIN และ `PATCH /users/:id` ไม่มี guard ถ้า role ยังกำหนดไม่ถูก permission layer ก็ไม่มีความหมาย

## 1. รูปแบบชื่อ permission

```
<resource>:<action>
```

- `resource` — โดเมนของข้อมูล: `user`, `device`, `telemetry`, `dashboard`
- `action` — `read` | `write` | `delete`
  - `read` = GET / list / query
  - `write` = create + update (POST / PATCH)
  - `delete` = soft delete (DELETE)

รายการ permission ที่ใช้จริงในระบบ:

| Permission          | ครอบคลุม endpoint (ดู api-spec.md)                          |
| ------------------- | ----------------------------------------------------------- |
| `user:read`         | `GET /users`, `GET /users/:id`                              |
| `user:write`        | `POST /users`, `PATCH /users/:id`                          |
| `user:delete`       | `DELETE /users/:id`                                        |
| `device:read`       | `GET /devices`, `GET /devices/:id`                        |
| `device:write`      | `POST /devices`, `PATCH /devices/:id`                    |
| `device:delete`     | `DELETE /devices/:id`                                     |
| `telemetry:read`    | `GET /telemetry/raw`, `/telemetry/aggregate`, `/latest`   |
| `telemetry:write`   | (สำหรับ ingest/admin tooling ที่เขียน telemetry ผ่าน HTTP — ปกติ telemetry เข้าทาง MQTT worker ไม่ผ่าน endpoint) |
| `dashboard:read`    | `GET /dashboard/summary`                                  |

> `telemetry:write` แทบไม่ถูกใช้ผ่าน HTTP เพราะ pipeline หลักคือ MQTT → worker (ดู architecture.md) มีไว้เผื่อ admin backfill/manual insert เท่านั้น — จะยังไม่สร้าง endpoint จนกว่าจะจำเป็น

## 2. Role → Permission matrix

Role ยังเป็นตัวหลักที่ผูกกับ user (คอลัมน์ `users.role`) permission ได้มาจากการ map role → set ของ permission (static table ในโค้ด ไม่เก็บใน DB)

| Permission        | ADMIN | STAFF | CUSTOMER |
| ----------------- | :---: | :---: | :------: |
| `user:read`       |  ✅   |  ✅   |    —     |
| `user:write`      |  ✅   |  —    |    —     |
| `user:delete`     |  ✅   |  —    |    —     |
| `device:read`     |  ✅   |  ✅   |   ✅     |
| `device:write`    |  ✅   |  ✅   |    —     |
| `device:delete`   |  ✅   |  —    |    —     |
| `telemetry:read`  |  ✅   |  ✅   |   ✅     |
| `telemetry:write` |  ✅   |  —    |    —     |
| `dashboard:read`  |  ✅   |  ✅   |   ✅     |

หลักการ:
- **ADMIN** — ทำได้ทุกอย่าง
- **STAFF** — ดูได้หมด, จัดการ device ได้, แต่แก้/ลบ user ไม่ได้
- **CUSTOMER** — ดู device/telemetry/dashboard ได้อย่างเดียว (read-only)

> **CUSTOMER กับ multi-tenant:** matrix นี้คุมแค่ "action ไหนทำได้" ยังไม่คุม "ดูของ site ไหนได้" ถ้าต้องแยกข้อมูลตามลูกค้า (row-level) ต้องเพิ่มการเช็ค `site_id` ต่อ record แยกต่างหาก — เป็น open question เดียวกับใน [architecture.md](./architecture.md#4-สิ่งที่ยังต้องตัดสินใจ-open-questions) ยังไม่ทำในเฟสแรก

## 3. การ implement

### 3.1 นิยาม permission กลาง (`pkg/core` หรือ `internal/authz`)

```go
// ตัวอย่างโครง — วางใน internal/authz/permission.go
package authz

const (
    UserRead        = "user:read"
    UserWrite       = "user:write"
    UserDelete      = "user:delete"
    DeviceRead      = "device:read"
    DeviceWrite     = "device:write"
    DeviceDelete    = "device:delete"
    TelemetryRead   = "telemetry:read"
    TelemetryWrite  = "telemetry:write"
    DashboardRead   = "dashboard:read"
)

// rolePermissions คือ source of truth ของ matrix ด้านบน
var rolePermissions = map[string][]string{
    "ADMIN": {
        UserRead, UserWrite, UserDelete,
        DeviceRead, DeviceWrite, DeviceDelete,
        TelemetryRead, TelemetryWrite, DashboardRead,
    },
    "STAFF": {
        UserRead, DeviceRead, DeviceWrite,
        TelemetryRead, DashboardRead,
    },
    "CUSTOMER": {
        DeviceRead, TelemetryRead, DashboardRead,
    },
}

func HasPermission(role, perm string) bool {
    for _, p := range rolePermissions[role] {
        if p == perm {
            return true
        }
    }
    return false
}
```

### 3.2 Middleware `RequirePermission`

เพิ่มใน `middleware/` คู่กับ `RequireRole` เดิม (pattern เดียวกัน — อ่าน `c.Locals("role")` ที่ `JWTMiddleware` ใส่ไว้แล้ว):

```go
// middleware/permission_middleware.go
func RequirePermission(perms ...string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        role, ok := c.Locals("role").(string)
        if !ok || role == "" {
            return core.SendError(c, fiber.StatusForbidden, "Access denied: no role found")
        }
        for _, p := range perms {
            if authz.HasPermission(role, p) {
                return c.Next()
            }
        }
        return core.SendError(c, fiber.StatusForbidden, "Access denied: missing permission")
    }
}
```

### 3.3 ใช้งานใน router

```go
// internal/api/telemetry/router.go
func TelemetryRouter(app fiber.Router, h Handler) {
    t := app.Group("/telemetry", middleware.JWTMiddleware())
    t.Get("/raw",       middleware.RequirePermission(authz.TelemetryRead),  h.GetRaw)
    t.Get("/aggregate", middleware.RequirePermission(authz.TelemetryRead),  h.GetAggregate)
    t.Get("/devices/:id/latest", middleware.RequirePermission(authz.TelemetryRead), h.GetLatest)
}

// internal/api/device/router.go
func DeviceRouter(app fiber.Router, h Handler) {
    d := app.Group("/devices", middleware.JWTMiddleware())
    d.Get("/",     middleware.RequirePermission(authz.DeviceRead),   h.GetAll)
    d.Get("/:id",  middleware.RequirePermission(authz.DeviceRead),   h.GetByID)
    d.Post("/",    middleware.RequirePermission(authz.DeviceWrite),  h.Create)
    d.Patch("/:id",middleware.RequirePermission(authz.DeviceWrite),  h.Update)
    d.Delete("/:id",middleware.RequirePermission(authz.DeviceDelete), h.Delete)
}
```

## 4. permission อยู่ใน JWT หรือ derive จาก role?

**เลือก: derive จาก role ตอน request (ไม่ฝังใน JWT)**

| | ฝัง permission ใน JWT claims | derive จาก role (แนะนำ) |
| --- | --- | --- |
| แก้ matrix แล้วมีผลทันที | ต้องรอ token หมดอายุ/reissue | มีผลทันทีทุก request |
| ขนาด token | ใหญ่ขึ้น | เท่าเดิม (มีแค่ `role`) |
| โค้ดที่ต้องแตะ | ต้องแก้ `core.Claims` + `GenerateTokenPair` | แค่เพิ่ม map + middleware |

`core.Claims` ปัจจุบันมีแค่ `userId/email/role` (`pkg/core/jwt.go`) — การ derive จาก role ไม่ต้องแตะ token layer เลย ถ้าอนาคตต้องการ per-user override (นอกเหนือ role) ค่อยเพิ่มตาราง `user_permissions` แล้ว merge กับ permission ของ role

## 5. ความสัมพันธ์กับ `RequireRole` เดิม

- `RequireRole` ยังใช้ได้ต่อ ไม่ต้องรื้อ แต่ endpoint ใหม่ควรใช้ `RequirePermission` เพราะสื่อความหมายชัดกว่าและ map ตรงกับ matrix
- ค่อย ๆ migrate ของเดิม: `RequireRole("ADMIN")` ที่ `POST /users` → `RequirePermission(authz.UserWrite)`, ที่ `DELETE /users/:id` → `RequirePermission(authz.UserDelete)`
- ข้อดี: จุดเดียว (`rolePermissions` map) เป็น source of truth ว่าใครทำอะไรได้ แทนที่จะกระจาย role string ตาม router หลายไฟล์
