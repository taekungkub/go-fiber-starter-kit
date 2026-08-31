# EMS Dashboard Monitor — API Spec

ทุก endpoint อยู่ใต้ `/api` (ตาม `api := app.Group("/api")` ใน `main.go`) และใช้ response envelope เดิมจาก `pkg/core` (`SendSuccess` / `SendError` / `SendValidationError`)

การคุมสิทธิ์ต่อ endpoint ใช้ permission แบบ `resource:action` (เช่น `telemetry:read`, `device:write`) — นิยาม, role→permission matrix และวิธี implement `RequirePermission` อยู่ใน [permissions.md](./permissions.md)

## Device Templates — `internal/api/device_template`

ทะเบียนชนิดอุปกรณ์ + นิยาม metric (ดู [data-model.md §1](./data-model.md)) — **จุดเดียวที่เพิ่ม/แก้ metric ของแต่ละ template โดยไม่ต้องแตะ schema DB**

| Method | Endpoint                       | Permission        | Description                                  |
| ------ | ------------------------------- | ----------------- | --------------------------------------------- |
| GET    | `/device-templates`             | `device:read`     | list templates ทั้งหมด (dashboard ใช้อ่านหน่วย/label metric) |
| GET    | `/device-templates/:key`        | `device:read`     | รายละเอียด template เดียว (รวม `metrics` JSONB) |
| POST   | `/device-templates`             | `device:write`    | สร้าง template ใหม่ (`template_key`, `name`, `metrics`) |
| PATCH  | `/device-templates/:key`        | `device:write`    | แก้ไข template — **ใช้เพิ่ม metric key ใหม่** (อัปเดต `metrics`) |

> **เพิ่ม metric key ใหม่ทำที่นี่ที่เดียว:** `PATCH /device-templates/:key` เพื่อเพิ่ม element ใน `metrics` (เช่น `{"key":"frequency","unit":"Hz"}`, ถ้าเป็นค่าสะสมใส่ `"cumulative":true`) — หลังจากนั้น ingest/worker/aggregator/API รองรับ metric ใหม่เองโดยไม่ต้องแก้โค้ดหรือ migrate (ดู [data-model.md §4.2](./data-model.md))

## Device — `internal/api/device`

| Method | Endpoint             | Permission        | Description                          |
| ------ | --------------------- | ----------------- | ------------------------------------- |
| GET    | `/devices`             | `device:read`     | list devices (pagination + sort, ตาม `pkg/core.Pagination`) |
| GET    | `/devices/:id`         | `device:read`     | รายละเอียด device เดียว                |
| POST   | `/devices`             | `device:write`    | สร้าง device ใหม่ (ต้องระบุ `device_template` ที่มีใน registry) |
| PATCH  | `/devices/:id`         | `device:write`    | แก้ไข device (name, site_id, is_active) |
| DELETE | `/devices/:id`         | `device:delete`   | soft delete (ถ้าต้องการ, ไม่บังคับ)     |

Query params สำหรับ `GET /devices`: `page`, `limit`, `sort`, `site_id` (filter), `device_template` (filter), `is_active` (filter)

> **ระวัง SQL injection ที่ `sort`:** repository เดิม (`internal/api/user/repository.go`) ต่อค่า `sort`/`order` เข้า query ด้วย `fmt.Sprintf("ORDER BY %s %s")` ตรง ๆ และ `pkg/core/sorting.go` ยัง **ไม่ได้** implement การ validate ตาม whitelist (มีแค่ comment ค้างไว้) — repo ใหม่ทุกตัวต้อง whitelist ชื่อคอลัมน์ที่ sort ได้จริง (map จาก query param → db column ที่รู้จัก) ก่อนใส่ลง SQL ห้ามส่งค่าดิบจาก client เข้า `ORDER BY`

## Telemetry — `internal/api/telemetry`

| Method | Endpoint                    | Permission       | Description                                     |
| ------ | ---------------------------- | ---------------- | ------------------------------------------------ |
| GET    | `/telemetry/raw`              | `telemetry:read` | ดึงจุดข้อมูลดิบทุกจุดในช่วง (zoom ละเอียด)         |
| GET    | `/telemetry/aggregate`        | `telemetry:read` | ดึง series สรุปตาม bucket (กราฟช่วงยาว)          |
| GET    | `/telemetry/devices/:id/latest`| `telemetry:read` | ค่าล่าสุดต่อ metric (live tile/KPI)             |

> **ไม่แยกเส้นตาม stat** — คำถามว่าควรมี `telemetry/raw`, `telemetry/raw-point`, `telemetry/raw-agg` แยกกันไหม: **ไม่ต้อง** แบ่งตาม stat (sum/avg/last) เพราะ agg row เดียวเก็บครบทุก stat อยู่แล้ว (`value_avg/min/max/sum/first/last`) ให้ client เลือก field ที่จะ plot เอง สิ่งที่ควรแยกคือแยกตาม **เจตนาการอ่าน 3 แบบ** ที่ query pattern + ปริมาณข้อมูลต่างกันจริง:
>
> | endpoint | ตอบอะไร | อ่านจาก | ปริมาณ |
> | --- | --- | --- | --- |
> | `/telemetry/raw` | ทุกจุดดิบในช่วงสั้น (zoom) | `telemetry_raw` | หนัก — จำกัดช่วง |
> | `/telemetry/aggregate` | series สรุปเป็น bucket (ช่วงยาว) | `telemetry_agg_*` | เบา |
> | `.../latest` | ค่าปัจจุบันตอนนี้ | `telemetry_raw` (จุดล่าสุด) | จิ๋ว |
>
> ข้อมูลเก็บแบบ **long (metric-per-row)** ทุก endpoint จึงรับ param `metric` เพื่อเลือก series ค่า metric ที่ใช้ได้ของแต่ละ device ดูจาก `device_templates.metrics`

### `GET /telemetry/raw`

Query params:

- `device_id` (required)
- `metric` (required) — เช่น `voltage`, `power` (ต้องอยู่ในนิยาม template ของ device นั้น)
- `from`, `to` (RFC3339, required) — จำกัดช่วงเวลาสูงสุด (เช่น ไม่เกิน 24 ชม.) เพราะ raw ปริมาณสูง
- `page`, `limit`

Response `data`: `Paging` ของ raw record (`device_id, metric, value, recorded_at`)

### `GET /telemetry/aggregate`

Query params:

- `device_id` (required)
- `metric` (required) — หรือรับหลายค่า (`metric=voltage&metric=power`) ถ้าอยากได้หลาย series ครั้งเดียว
- `interval` (required) — `hourly` | `daily` | `monthly`
- `from`, `to` (RFC3339, required)

Response `data`: array ของ agg rows เรียงตาม `bucket_start` — แต่ละ row มีครบทุก stat ให้ client เลือก plot:

```
{ device_id, metric, bucket_start,
  value_avg, value_min, value_max, value_sum,
  value_first, value_last,           // last = ค่าปิด, energy delta = last - first
  sample_count }
```

handler เลือกตารางตาม `interval` (`telemetry_agg_hourly` / `_daily` / `_monthly`) แล้วคืนตรง ๆ ไม่ต้อง pagination (ช่วงเวลาถูกจำกัดโดย client) ถ้าขอหลาย metric ให้ client group ตาม `metric` เอง

> **metric สะสม (energy):** ค่าใช้ในช่วง = `value_last - value_first` (ไม่ใช่ `value_sum`) — client คำนวณจาก field ที่ให้ หรือจะเพิ่ม field `value_delta` ใน response เฉพาะ metric ที่ `cumulative:true` ก็ได้
>
> **ถ้าต้องการ interval ที่ไม่มีตาราง** (เช่น 10 นาที): ทำ on-the-fly ด้วย `date_bin` (Postgres 14+) บน `telemetry_raw` ในช่วงที่จำกัด — ยังไม่ทำเฟสแรก เริ่มจาก hourly/daily/monthly ที่ precompute ไว้ก่อน

### `GET /telemetry/devices/:id/latest`

Response `data`: ค่าล่าสุดต่อ metric ของ device นั้น — array ของ `(metric, value, recorded_at)` (query แบบ `DISTINCT ON (metric) ... ORDER BY metric, recorded_at DESC`) ใช้แสดงค่าปัจจุบันบน dashboard tile

## Dashboard summary (ทางเลือก — รวม endpoint ให้เรียกครั้งเดียว)

| Method | Endpoint             | Auth | Description                                              |
| ------ | --------------------- | ---- | ----------------------------------------------------------- |
| GET    | `/dashboard/summary`   | JWT  | รวมค่าล่าสุด + สรุป 24 ชม. ของทุก device (หรือของ site ที่ระบุ) |

พิจารณาทำ endpoint นี้ทีหลัง ถ้า dashboard ต้อง fetch หลาย endpoint พร้อมกันแล้วช้า — ไม่ทำตั้งแต่แรกเพื่อไม่ over-engineer

## Error cases ที่ต้อง handle เป็นพิเศษ

- `device_id` ไม่พบ → 404 ผ่าน `core.SendError`
- ช่วงเวลา `from`/`to` ไม่ถูกต้อง หรือกว้างเกินที่กำหนด (เช่น raw query > 24h) → 400 validation error ผ่าน `core.SendValidationError`
- `interval` ไม่อยู่ใน `hourly|daily|monthly` → 400
- `metric` ไม่อยู่ในนิยาม template ของ device นั้น → 400

## Real-time (ยังไม่ตัดสินใจ)

ถ้าต้อง push ข้อมูลสด ๆ ขึ้น dashboard โดยไม่ polling ให้เพิ่ม WebSocket endpoint แยก (เช่น `/ws/telemetry?device_id=`) ที่ subscribe ข้อมูลจาก worker แบบ fan-out — ยังไม่ design รายละเอียดจนกว่าจะยืนยัน requirement (ดู open question ใน `architecture.md`)
