# EMS Dashboard Monitor — Architecture

## 1. ภาพรวม

ระบบ EMS (Energy Management System) Dashboard Monitor รับข้อมูล telemetry จากอุปกรณ์วัดพลังงาน (มิเตอร์ไฟฟ้า/เซนเซอร์) ผ่าน MQTT (ต้นทางคือ Node-RED), เก็บข้อมูลดิบ, ประมวลผลรวม (aggregate) เป็นช่วงเวลาต่าง ๆ และแสดงผลผ่าน REST API สำหรับ dashboard

ระบบต่อยอดจาก pipeline ที่มีอยู่แล้วในโปรเจกต์นี้ (`internal/ingest` → `internal/queue` → `internal/worker`, ดู `MQTT.md`) โดยเปลี่ยนจาก mock producer เป็นของจริง และเพิ่มชั้น aggregation กับ API

```
Node-RED (field devices)
    │  MQTT publish
    ▼
MQTT Broker
    │  subscribe
    ▼
Ingestion Bridge (internal/ingest)
    │  parse + validate payload → worker.Job
    ▼
JobQueue (internal/queue, buffered chan)
    │
    ▼
Workers (internal/worker, N goroutines)
    │  batch (size N) → bulk insert
    ▼
PostgreSQL: telemetry_raw (long: 1 row = 1 metric)
    │
    ▼ (scheduled job, e.g. cron / ticker)
Aggregator (layered rollup)
    │  raw → hourly → daily → monthly
    ▼
PostgreSQL: telemetry_agg_hourly / _daily / _monthly
    │
    ▼
REST API (internal/api/telemetry, internal/api/device)
    │
    ▼
Dashboard (frontend, out of scope of this repo)
```

## 2. องค์ประกอบหลัก

### 2.1 Ingestion Bridge (`internal/ingest`) — **จุดสำคัญของระบบ**

> **Ingest path = MQTT (ตัดสินแล้ว)** อุปกรณ์/Node-RED publish ขึ้น broker, bridge นี้ subscribe แล้วป้อนเข้า pipeline — ไม่ใช้ HTTP POST ตรงแบบโปรเจกต์อื่น

Bridge เป็นตัวเชื่อม "โลกภายนอก (MQTT/Node-RED)" กับ "pipeline ภายใน (queue→worker→DB)" หน้าที่ของมันคือ **แปลง + validate + fan-out + ส่งต่อ** เท่านั้น **ห้ามแตะ DB เอง** (แยก concern ให้ worker เป็นคนเขียน) เพื่อให้ bridge บางและทน load

หลักการออกแบบ:

**a) การเชื่อมต่อ (paho `eclipse/paho.mqtt.golang`)**
- config: `MQTT_BROKER_URL`, `MQTT_USERNAME`, `MQTT_PASSWORD`, `MQTT_CLIENT_ID`, `MQTT_TOPIC_PATTERN` (default `ems/+/+/telemetry`)
- `client_id` ต้อง**คงที่ต่อ instance** (ถ้าสุ่มทุก reconnect จะเสีย session) — ถ้า scale หลาย instance ให้ต่อท้ายด้วย hostname/index
- **QoS 1** (at-least-once) — รับซ้ำได้แต่ไม่หาย เหมาะกับ telemetry; ต้องเผื่อ **dedup/idempotent** ปลายทาง (ดูข้อ e)
- `SetAutoReconnect(true)` + `OnConnectionLost` / `OnReconnecting` handler + resubscribe เมื่อ reconnect (paho บาง config ไม่ auto-resub)
- `SetCleanSession(false)` ถ้าอยากให้ broker คิว message ที่พลาดช่วงหลุด (แลกกับภาระ broker) — ตัดสินตาม broker จริง

**b) Subscribe + แยก topic**
- subscribe `ems/+/+/telemetry` (wildcard ครอบทุก site/device)
- แยก `site_id` / `device_id` ออกจาก topic string ในทุก message (ดู [§3](#3-mqtt-topic-convention))

**c) Validate + resolve template (มี cache)**
- lookup `device_id` → หา `device_template` + นิยาม metric — **cache in-memory** (map + TTL หรือ refresh เป็นรอบ) อย่า query DB ต่อ message เพราะ throughput สูง
- `device_id` ที่ไม่รู้จัก → นโยบายเลือกได้: **log/ข้าม** (strict) หรือ auto-register device ใหม่แบบ inactive ให้ admin มายืนยัน — เริ่มด้วย strict
- payload JSON parse ไม่ผ่าน → log + `raw_dropped` counter แล้วข้าม (ห้าม panic/หยุด loop)

**d) Fan-out เป็น metric rows (long format)**
- payload 1 ก้อน → แตกแต่ละ key เป็น 1 reading `(device_id, metric, value, recorded_at)`
- validate key กับ `device_templates.metrics` — key แปลก (ไม่มีในนิยาม) → log/ข้ามเฉพาะ key นั้น (ไม่ทิ้งทั้ง payload)
- `recorded_at` เอาจาก field `timestamp` ใน payload ถ้ามี, ไม่มีค่อย fallback เป็นเวลารับ

**e) ส่งเข้า queue + backpressure (สำคัญ)**
- push `worker.Job` เข้า `queue.JobQueue`
- **queue เต็มจะทำยังไง** ต้องตัดสินใจ (ปัจจุบัน channel buffered 1000 — push จะ block เมื่อเต็ม):
  - *block* → กัน MQTT ack ค้าง อาจทำให้ broker resend — ปลอดภัยกับข้อมูลแต่ดัน backpressure กลับไป broker
  - *drop + นับ metric* → ไม่หน่วง แต่ข้อมูลหาย
  - เริ่มด้วย block + เพิ่มจำนวน worker/ขนาด queue เป็น config, ใส่ metric `queue_depth` เฝ้าดู
- dedup: ถ้า QoS1 ส่งซ้ำ ให้ worker ใช้ upsert หรือ unique `(device_id, metric, recorded_at)` กันซ้ำที่ raw (ตัดสินตอนเจอปัญหาจริง)

**f) อัปเดต `devices.last_seen_at`** เมื่อรับ payload (throttle เช่นไม่เกิน 1 ครั้ง/นาที/device เพื่อไม่ให้ write ถี่เกิน) — ทำใน worker ตอน batch ก็ได้

โครงไฟล์ที่แนะนำ: แยกเป็น `internal/ingest/mqtt_bridge.go` (จริง) กับคง `mqtt_consumer.go` (mock) ไว้ เปิดใช้ตัวไหนผ่าน config `MQTT_MOCK` — dev ที่ไม่มี broker ยังรันได้

หมายเหตุสถานะ: ยังไม่ได้ implement — ปัจจุบันมีแค่ mock ticker (`internal/ingest/mqtt_consumer.go`) ที่ push job ปลอมทุก 1 วินาที

### 2.2 Queue (`internal/queue`)

คงพฤติกรรมเดิม (`chan worker.Job` buffered) แต่ `Job` ต้อง extend ให้เก็บ metric rows ที่ parse แล้วสำหรับ raw insert (long format — ดู [data-model.md §3](./data-model.md))

### 2.3 Worker / Batch writer (`internal/worker`)

ปัจจุบัน batch แค่ log อย่างเดียว (`fmt.Println`) — ต้องเปลี่ยนเป็น bulk insert ลง `telemetry_raw` จริง โดย:

- batch size และ flush interval ควรปรับเป็น config ได้ (ไม่ hardcode `10`)
- ต้อง flush ด้วย timer ด้วย ไม่ใช่รอ batch เต็มอย่างเดียว (ป้องกันข้อมูลค้างใน memory ถ้า throughput ต่ำ)
- insert ผิดพลาดต้อง log พร้อม payload ที่ fail (พิจารณา dead-letter ในอนาคต ถ้าจำเป็น)

### 2.4 Aggregator (ใหม่ — ยังไม่มีในโค้ด)

Background job แยกต่างหาก (เริ่มใน `main.go` เหมือน worker) ที่รันเป็นรอบ (เช่นทุก 1 ชม.) แบบ **layered rollup**:

- **hourly** (base): อ่าน `telemetry_raw` ช่วงที่ยังไม่ agg → `GROUP BY device_id, metric, date_trunc('hour', ...)` → `avg/min/max/sum(value)`, `count`
- **daily / monthly**: rollup ต่อจากชั้นบน (daily จาก hourly, monthly จาก daily) ไม่ scan raw ซ้ำ
- upsert ลง `telemetry_agg_hourly` / `_daily` / `_monthly` (`ON CONFLICT (device_id, metric, bucket_start) DO UPDATE`)
- ⚠️ avg ชั้นบนต้องถ่วงน้ำหนักด้วย `sample_count`, และ metric สะสม (เช่น `energy`) ต้องใช้ `last-first` ไม่ใช่ `sum` — ดู [data-model.md §4.1](./data-model.md)

รายละเอียด schema (long format, `device_templates`, rollup) ดู [data-model.md](./data-model.md)

### 2.5 API layer (ใหม่)

โมดูลใหม่ตาม pattern เดิมของโปรเจกต์ (`internal/api/<module>/{router,handler,usecase,repository,<module>}.go`):

- `internal/api/device` — CRUD/list อุปกรณ์
- `internal/api/telemetry` — query raw และ aggregate telemetry สำหรับ dashboard

รายละเอียด endpoint ดู [api-spec.md](./api-spec.md)

## 3. MQTT topic convention

```
ems/{site_id}/{device_id}/telemetry
```

- `site_id` — รหัสไซต์/สถานที่ติดตั้ง
- `device_id` — รหัสอุปกรณ์ (ตรงกับ `devices.device_id`)
- payload เป็น JSON โดย **key แต่ละตัวคือ metric ตามนิยามใน `device_templates` ของ device นั้น** (power_meter ส่ง voltage/current/power/energy, water ส่ง flow/volume/pressure ฯลฯ) — Ingestion Bridge จะ **fan-out แต่ละ key เป็น 1 row** ใน `telemetry_raw` (long format)

```json
// device template = power_meter
{
  "voltage": 220.5,
  "current": 4.8,
  "power": 1058.4,
  "energy": 123456.7,
  "timestamp": "2026-08-31T10:00:00Z"
}
// → 4 rows: (device_id, metric=voltage, value=220.5, ...), (…current…), (…power…), (…energy…)
```

ingest ควร validate key ที่รับมากับ `device_templates.metrics` (key ที่ไม่รู้จัก → log/ข้าม)

Ingestion Bridge subscribe ด้วย wildcard `ems/+/+/telemetry` เพื่อรับทุก site/device ในครั้งเดียว

## 4. Decisions & open questions

ตัดสินแล้ว:
- **Ingest = MQTT** (ไม่ใช่ HTTP POST ตรง)
- **retention `telemetry_raw` = 90 วัน** — purge ด้วย drop partition รายเดือน (ดู [data-model.md §5](./data-model.md))

ยังต้องตัดสิน:
- ต้องการ real-time push ไปหน้า dashboard ด้วยไหม (WebSocket/SSE) หรือ polling ผ่าน REST พอ
- ต้องรองรับ multi-tenant (หลายลูกค้า/หลายไซต์แยกสิทธิ์) หรือไม่ — ถ้าใช่ต้องผูกกับ role/user เดิมใน `internal/api/user`
- metric ที่ไม่ใช่ตัวเลข (status string/bool) มีไหม — ถ้ามีต้องเพิ่ม `value_text` ใน `telemetry_raw`
