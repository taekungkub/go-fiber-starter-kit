# EMS Dashboard Monitor — Implementation Plan

ต่อยอดจาก pipeline mock ที่มีอยู่แล้ว (`internal/ingest` → `internal/queue` → `internal/worker`) ทำเป็นเฟส เพื่อให้แต่ละเฟส deploy/ทดสอบได้อิสระ

## Phase 0 — Schema เตรียมพื้นฐาน

- เพิ่ม migration ตาม [data-model.md](./data-model.md): `tb_device_templates.go`, `tb_devices.go`, `tb_telemetry.go` (raw + `telemetry_agg_hourly`/`_daily`/`_monthly`)
- เรียก `MigrateTableDeviceTemplates(db)` → `MigrateTableDevices(db)` → `MigrateTableTelemetry(db)` ต่อจาก `migrations.MigrateTableUsers(db)` ใน `main.go` (templates ก่อน เพราะ devices อ้าง `device_template`)
- seed `device_templates` เริ่มต้นอย่างน้อย 1 ตัว (เช่น `power_meter` พร้อมนิยาม metrics) เพื่อใช้ทดสอบ ingest

## Phase 1 — Real batch write (ยังไม่ต้องมี MQTT จริง)

- แก้ `internal/worker/worker.go`: เปลี่ยน `fmt.Println("Processing batch...")` เป็น bulk insert ลง `telemetry_raw` จริง (ใช้ `sqlx.NamedExec` หรือ `pgx` COPY ถ้า throughput สูง)
- เพิ่ม flush ด้วย timer (`time.Ticker`) ควบคู่กับ flush ตาม batch size เต็ม เพื่อไม่ให้ข้อมูลค้างเมื่อ throughput ต่ำ
- ทำ batch size / flush interval เป็นค่า config (`WORKER_BATCH_SIZE`, `WORKER_FLUSH_INTERVAL`)
- แก้ `worker.Job` ให้มี field ที่ parse แล้วแบบ long — 1 reading fan-out เป็นหลาย metric row เช่น `Readings []struct{ DeviceID, Metric string; Value float64; RecordedAt time.Time }` แทน raw `[]byte` เพียงอย่างเดียว (ดู long format ใน [data-model.md](./data-model.md))
- ทดสอบด้วย mock ingest เดิม (`StartMQTTConsumerMock`) ที่ยังทำงานอยู่ — เห็นข้อมูลจริงลง DB ก็ถือว่า phase นี้เสร็จ

## Phase 2 — Ingestion Bridge จริง (MQTT ↔ Node-RED)

- เพิ่ม dependency `github.com/eclipse/paho.mqtt.golang`
- เพิ่ม config: `MQTT_BROKER_URL`, `MQTT_USERNAME`, `MQTT_PASSWORD`, `MQTT_CLIENT_ID`, `MQTT_TOPIC_PATTERN` (default `ems/+/+/telemetry`)
- เขียน `internal/ingest/mqtt_consumer.go` ใหม่ (หรือไฟล์ใหม่ `mqtt_bridge.go` แล้วค่อยลบตัว mock) ที่:
  - เชื่อมต่อ broker, subscribe topic pattern
  - แยก `site_id`/`device_id` จาก topic string
  - parse JSON payload → fan-out แต่ละ key เป็น metric row, validate key กับ `device_templates.metrics` ของ device (key แปลก → log/ข้าม)
  - push เป็น `worker.Job` เข้า `queue.JobQueue`
  - reconnect handler (paho รองรับ `OnConnectionLost`/`OnReconnecting`)
- คง mock ไว้เป็นทางเลือกสำหรับ dev ที่ไม่มี broker (เปิดผ่าน config flag เช่น `MQTT_MOCK=true`)
- อัปเดต device `last_seen_at` เมื่อรับ payload (อาจทำใน worker ตอน batch insert หรือ upsert แยก)

## Phase 3 — Aggregator

- สร้าง `internal/aggregator/aggregator.go` (package ใหม่ ระดับเดียวกับ `internal/worker`)
- `StartAggregator(db *sqlx.DB, interval time.Duration)` รันเป็น ticker (เช่นทุก 1 ชม.) แบบ **layered rollup**:
  - **hourly** (base): query `telemetry_raw` ช่วงที่ยังไม่ agg (เก็บ watermark ในตัวแปรหรือ table เล็ก `agg_watermark`) → `GROUP BY device_id, metric, date_trunc('hour', recorded_at)` → `avg/min/max/sum(value)`, `count`
  - **daily / monthly**: rollup ต่อจากชั้นบน (daily จาก hourly, monthly จาก daily) — ไม่ scan raw ซ้ำ
  - upsert: `INSERT INTO telemetry_agg_hourly ... ON CONFLICT (device_id, metric, bucket_start) DO UPDATE` (และ `_daily` / `_monthly`)
  - ⚠️ avg ชั้นบนถ่วงน้ำหนักด้วย `sample_count`; metric ที่มี `cumulative:true` (เช่น energy) ใช้ `last-first` ไม่ใช่ `sum` — ดู [data-model.md §4.1](./data-model.md)
- เริ่มใน `main.go` คู่กับ `worker.StartWorkers(...)`

## Phase 4 — API layer

- สร้าง `internal/api/device_template/{...}.go` — CRUD registry (จุดเพิ่ม metric key ใหม่ผ่าน `PATCH /device-templates/:key`)
- สร้าง `internal/api/device/{device,handler,usecase,repository,router}.go` ตาม pattern เดิม (ดู `internal/api/user` เป็นตัวอย่าง)
- สร้าง `internal/api/telemetry/{telemetry,handler,usecase,repository,router}.go` — query แบบ long (param `metric`, `interval=hourly|daily|monthly`)
- ต่อ router เข้า `main.go` แบบเดียวกับ auth/user module, คุมสิทธิ์ด้วย `RequirePermission` (ดู [permissions.md](./permissions.md))
- ใช้ `pkg/core.ValidateStruct` สำหรับ query param validation (`from`/`to`/`interval`/`metric`)
- endpoint ตาม [api-spec.md](./api-spec.md)

## Phase 5 — Hardening / polish

- retention `telemetry_raw` = **90 วัน**: range-partition รายเดือน + job สร้าง partition เดือนถัดไปล่วงหน้า + job `DROP`/`DETACH` partition ที่เก่ากว่า 90 วัน (เก็บ ~3 partition ล่าสุด)
- metrics พื้นฐาน: queue length, batch flush latency, MQTT reconnect count (log ก่อน ยังไม่ต้องมี metrics exporter จนกว่าจะจำเป็น)
- พิจารณา WebSocket/live push ถ้า requirement ยืนยันว่าต้องการ real-time (ดู open question ใน architecture.md)

## ลำดับความสำคัญที่แนะนำ

**ถ้าเน้น ingestion (ตามที่ตัดสิน):** Phase 0 → 1 → **2 (MQTT bridge)** → 4 → 3 → 5

ดัน Phase 2 ขึ้นมาเร็ว เพราะ MQTT bridge เป็นหัวใจของระบบ (ดู [architecture.md §2.1](./architecture.md)) อยากพิสูจน์ว่า Node-RED → broker → bridge → queue → worker → DB ไหลจริงตั้งแต่ต้น ๆ; Phase 1 (batch write จริง) ต้องมาก่อนเพื่อให้ bridge มีปลายทางเขียนลง DB

**ถ้าเน้นเห็นผลบน dashboard เร็ว:** Phase 0 → 1 → 4 → 3 → 2 → 5 — ใช้ mock ingest ไปก่อน ทำ API query ให้ frontend เห็นข้อมูลเร็วสุด แล้วค่อยต่อ MQTT จริงทีหลัง
