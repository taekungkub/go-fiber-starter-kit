# EMS Dashboard Monitor — Data Model

ตารางทั้งหมดใหม่ ต่อยอดจาก schema เดิม (`migrations/tb_users.go` เป็น pattern: ฟังก์ชัน `CreateTableX()` คืน SQL string + `MigrateTableX(db)` เรียกตอน startup ใน `main.go`) ให้สร้างไฟล์ migration ในสไตล์เดียวกัน เช่น `migrations/tb_devices.go`, `migrations/tb_telemetry.go`

> **หลีกเลี่ยง Postgres enum:** ตารางใหม่เหล่านี้ตั้งใจไม่ใช้ enum type (เช่น `metric`, `device_template` เป็น `VARCHAR` ไม่ใช่ enum) เพราะ pattern migration แบบ `CREATE TABLE IF NOT EXISTS` ในโปรเจกต์นี้ **ไม่ได้สร้าง enum type ให้** — ดูได้จาก `tb_users.go` ที่อ้าง type `user_role` แต่ enum ตัวจริงถูกประกาศแค่ใน `001_init.sql` ที่ไม่ถูกเรียกใช้ตอน runtime (และ `main.go` ก็ ignore error จาก migrate) การใช้ VARCHAR + registry table (`device_templates`) ยังเข้ากับแนวทาง "เพิ่ม template/metric ใหม่โดยไม่แตะ schema" ด้วย

## โมเดลโดยรวม (long / generic format)

เลือกเก็บ telemetry แบบ **long (metric-per-row)** ไม่ใช่ wide (metric เป็นคอลัมน์) เพราะแต่ละ device มี metric ต่างชุดกัน (power_meter → voltage/current/power/energy, water → flow/volume/pressure, ...) การเก็บแบบ long ทำให้:

- เพิ่ม device template / metric ใหม่ = **insert row ไม่ต้องแก้ schema**
- aggregator + API เขียน code path เดียว generic ใช้ได้ทุก metric

แลกกับ 2 cost ที่ต้องรู้: (1) 1 reading ที่มี N metric = N rows (กระทบ `telemetry_raw` เป็นหลัก คุมด้วย retention), (2) ถ้า dashboard อยากได้หลาย metric ในแถวเดียวต้อง pivot ตอน query (กราฟ 1 metric = 1 series ไม่ต้อง pivot)

```
device_templates ──นิยาม metric──┐
                                  ▼
devices ──(device_template)──▶ ทุก row ของ telemetry อ้าง device_id
   │
   ▼
telemetry_raw (long) ─rollup→ telemetry_agg_hourly ─rollup→ _daily ─rollup→ _monthly
```

## 1. `device_templates` (registry)

ทะเบียนกลางของชนิดอุปกรณ์ + นิยาม metric ที่ template นั้นมี — เป็น source of truth ให้ ingest ใช้ validate payload และให้ dashboard อ่านหน่วย/label

| column       | type         | note                                                        |
| ------------ | ------------ | ----------------------------------------------------------- |
| template_key | VARCHAR(50) PK | เช่น `power_meter`, `water`, `chiller`                     |
| name         | VARCHAR(100) | ชื่อแสดงผล                                                   |
| metrics      | JSONB        | นิยาม metric: `[{"key":"voltage","unit":"V","label":"แรงดัน"}, ...]` |
| is_active    | BOOLEAN      | default TRUE                                                 |
| created_at   | TIMESTAMPTZ  | default NOW()                                               |
| updated_at   | TIMESTAMPTZ  | default NOW()                                               |

เพิ่มชนิดอุปกรณ์ใหม่ = insert 1 row ที่นี่ (พร้อมนิยาม metric ใน JSONB) — ไม่ต้องสร้างตารางหรือแก้ schema

## 2. `devices`

| column          | type          | note                                              |
| --------------- | ------------- | -------------------------------------------------- |
| id              | UUID PK       | `gen_random_uuid()`                                |
| device_id       | VARCHAR(100)  | unique, key ที่ telemetry อ้างถึง (ตรงกับ topic/แหล่งข้อมูล) |
| site_id         | VARCHAR(100)  | ไซต์ที่ติดตั้ง                                     |
| name            | VARCHAR(200)  | ชื่อที่ใช้แสดงบน dashboard                         |
| device_template | VARCHAR(50)   | FK → `device_templates.template_key` — บอกว่ามี metric ชุดไหน |
| is_active       | BOOLEAN       | default TRUE                                       |
| last_seen_at    | TIMESTAMPTZ   | อัปเดตทุกครั้งที่ ingest เข้ามา                    |
| created_at      | TIMESTAMPTZ   | default NOW()                                      |
| updated_at      | TIMESTAMPTZ   | default NOW()                                      |

Index: unique `(device_id)`, index `(site_id)`, index `(device_template)`

## 3. `telemetry_raw` (long)

ข้อมูลดิบ 1 row = 1 metric ของ 1 reading — ปริมาณสูง เขียนแบบ batch insert จาก `internal/worker`

| column          | type         | note                                                      |
| --------------- | ------------ | --------------------------------------------------------- |
| id              | BIGSERIAL PK |                                                           |
| device_id       | VARCHAR(100) | soft ref → `devices.device_id` (**ไม่บังคับ FK constraint** เพื่อ insert เร็ว — integrity ย้ายไปกันที่ ingest layer: validate device_id กับ cache ก่อน enqueue, ดู [ticket 07](./tickets/07-mqtt-ingestion-bridge.md)) |
| device_template | VARCHAR(50)  | denormalize ไว้เพื่อ filter/agg ต่อชนิดโดยไม่ต้อง join (optional แต่แนะนำ) |
| metric          | VARCHAR(50)  | เช่น `voltage`, `power`, `flow` — ต้องตรงกับ key ใน template |
| value           | NUMERIC      | ค่าที่วัดได้                                              |
| recorded_at     | TIMESTAMPTZ  | timestamp จาก payload (ไม่ใช่เวลารับ)                     |
| ingested_at     | TIMESTAMPTZ  | default NOW() — เวลาที่ระบบรับ/insert                     |

Index: `(device_id, metric, recorded_at)` — ครอบคลุม query "metric X ของ device Y ช่วงเวลา" และเป็นลำดับที่ aggregator ใช้ scan/GROUP BY

> **metric ที่ไม่ใช่ตัวเลข** (เช่น status เป็น string/bool): โมเดลนี้ `value` เป็น NUMERIC ตัวเดียว ถ้ามี metric แบบนั้นให้ map เป็นเลข (0/1) หรือเพิ่มคอลัมน์ `value_text VARCHAR NULL` แยก — ตัดสินตอนเจอ device จริงที่ต้องใช้

พิจารณา partition ตามเดือน (`recorded_at`) เมื่อ volume สูง — ดู §5

## 4. Aggregated telemetry — `telemetry_agg_hourly` / `_daily` / `_monthly`

ข้อมูลสรุปตามช่วงเวลา คำนวณโดย Aggregator job — **interval อยู่ในชื่อตาราง** (ตาม convention ของ production เดิม) และเก็บแบบ long เหมือน raw

เหตุที่เริ่มหยาบที่ **hourly** (ไม่มี 1m/5m): ปัญหา bloat แทบหายไป — @100 devices ต่อ 1 metric: hourly ~876K rows/ปี, daily ~36K/ปี, monthly ~1.2K/ปี ตารางพวกนี้ **ไม่ต้อง partition / ไม่ต้อง retention job** เก็บยาวได้สบาย (ตัวที่ต้องคุม retention คือ `telemetry_raw` เท่านั้น) ถ้าภายหลังต้องการ 5m/15m ค่อยเพิ่มตาราง `telemetry_agg_5min` ฯลฯ ตาม pattern เดิม

Schema (เหมือนกันทั้ง 3 ตาราง):

| column        | type         | note                                              |
| ------------- | ------------ | -------------------------------------------------- |
| id            | BIGSERIAL PK |                                                    |
| device_id     | VARCHAR(100) |                                                    |
| metric        | VARCHAR(50)  |                                                    |
| bucket_start  | TIMESTAMPTZ  | จุดเริ่มของช่วงเวลา (`date_trunc('hour'/'day'/'month', ...)`) |
| value_avg     | NUMERIC      | ค่าเฉลี่ยในช่วง                                    |
| value_min     | NUMERIC      |                                                    |
| value_max     | NUMERIC      |                                                    |
| value_sum     | NUMERIC      | ผลรวม (ใช้กับ metric ที่ sum ได้ / ดูหมายเหตุ energy) |
| value_first   | NUMERIC      | ค่าแรกของ bucket (เรียงตาม `recorded_at`) — ใช้กับ energy delta / กราฟ open |
| value_last    | NUMERIC      | ค่าสุดท้ายของ bucket — ใช้ตอบ "last/ค่าปิด" และ energy delta (`last-first`) |
| sample_count  | INT          | จำนวน row ที่ใช้คำนวณ bucket นี้ — **จำเป็นสำหรับ rollup ชั้นถัดไป** |
| created_at    | TIMESTAMPTZ  | default NOW()                                      |

> เก็บ stat หลายตัวต่อ bucket (`avg/min/max/sum/first/last/count`) ในแถวเดียว — **client เลือกเองว่าจะ plot ตัวไหน** ไม่ต้องแยก endpoint ต่อ stat (ดู [api-spec.md](./api-spec.md))

Index: unique `(device_id, metric, bucket_start)` — ใช้เป็น upsert key (`ON CONFLICT DO UPDATE`) เวลา aggregator รันซ้ำ/ชนกัน และทำหน้าที่เป็น index หลักของ query

### 4.1 Rollup แบบชั้น (layered) + ความถูกต้องของ avg

```
telemetry_raw ─▶ telemetry_agg_hourly ─▶ telemetry_agg_daily ─▶ telemetry_agg_monthly
```

- **hourly** (base): `GROUP BY device_id, metric, date_trunc('hour', recorded_at)` จาก raw → `avg/min/max/sum(value)`, `count(*)`
- **daily / monthly**: rollup **ต่อจากชั้นบน** (daily จาก hourly, monthly จาก daily) ไม่ต้อง scan raw ใหม่

> ⚠️ **avg-of-avg ต้องถ่วงน้ำหนักด้วย `sample_count`** ไม่งั้นค่าเพี้ยน:
> - `value_avg` (daily) = `sum(value_avg * sample_count) / sum(sample_count)`
> - `value_min` = `min(value_min)`, `value_max` = `max(value_max)`, `value_sum` = `sum(value_sum)`, `sample_count` = `sum(sample_count)`
> - `value_first` = `value_first` ของ bucket ที่ `bucket_start` น้อยสุด, `value_last` = `value_last` ของ bucket ที่ `bucket_start` มากสุด (เรียงตามเวลา ไม่ใช่ min/max ของค่า)

> ⚠️ **metric สะสม (cumulative) เช่น `energy`**: ค่าที่ meter ส่งมาเป็นเลขสะสม การหา "พลังงานที่ใช้ในช่วง" = `value_last - value_first` ของ bucket นั้น ไม่ใช่ `value_sum` — เก็บ `value_first`/`value_last` ไว้ในตารางแล้ว (คำนวณด้วย `FIRST_VALUE`/`LAST_VALUE` ตอน rollup จาก raw) API จึงคืน delta ได้โดยไม่ต้อง scan raw ซ้ำ aggregator รู้ว่า metric ไหน cumulative จาก flag `cumulative:true` ใน `device_templates.metrics`

### 4.2 เพิ่ม metric key ใหม่ = แก้ที่เดียว (`device_templates.metrics`)

ข้อได้เปรียบหลักของ long format: **เพิ่ม metric ใหม่ไม่ต้องแตะ schema/migration เลย** เพราะ metric เป็น *ค่าใน row* ไม่ใช่ *คอลัมน์* ทำแค่:

```sql
-- power_meter เพิ่ม metric 'frequency'
UPDATE device_templates
SET metrics = metrics || '[{"key":"frequency","unit":"Hz"}]'::jsonb
WHERE template_key = 'power_meter';
```

(ผ่าน API คือ `PATCH /device-templates/:key` — ดู [api-spec.md](./api-spec.md)) หลังจากนั้น:

| ขั้น | ต้องแก้อะไร |
| --- | --- |
| ingest validate | ไม่ต้อง — อ่าน key ที่รับได้จาก `device_templates.metrics` (ผ่าน validate ทันที) |
| worker insert    | ไม่ต้อง — insert เป็น row `(device_id, metric='frequency', value, ...)` ปกติ |
| aggregator       | ไม่ต้อง — `GROUP BY metric` ครอบ metric ใหม่ให้เอง |
| API              | ไม่ต้อง — query ด้วย `?metric=frequency` ได้เลย |

**policy สำหรับ key ที่ยังไม่อยู่ใน template** (แนะนำ strict): ingest เจอ key ที่ไม่มีในนิยาม → **log แล้วข้าม** ไม่ auto-insert (กันข้อมูลขยะจากพิมพ์ key ผิด) อยากรับจริงให้เพิ่มใน registry ก่อน

> ⚠️ **strict = ข้อมูลหายเงียบ:** การ "log แล้วข้าม" ทำให้ data loss มองไม่เห็นถ้าไม่มี metric/alert — ต้องนับ drop **แยกตามสาเหตุ** (unknown device / unknown metric / parse error) ให้เห็น rate ไม่งั้นจะไม่รู้ว่าเสีย metric ไปกี่ device กี่วัน ดู [ticket 07](./tickets/07-mqtt-ingestion-bridge.md)

> เทียบกับ wide (แบบ A ที่ไม่ได้เลือก): metric ใหม่ = `ALTER TABLE ADD COLUMN` ทุกตาราง (raw + hourly + daily + monthly) + แก้ struct/query ทุกจุด

## 5. Retention & partition (เฉพาะ `telemetry_raw`)

| ตาราง                   | เก็บย้อนหลัง (แนะนำเริ่มต้น) | จัดการ                                 |
| ----------------------- | --------------------------- | --------------------------------------- |
| `telemetry_raw`         | **90 วัน** (ตัดสินแล้ว)      | **range-partition รายเดือน** ตาม `recorded_at` แล้ว purge ด้วย `DROP`/`DETACH PARTITION` (instant, ไม่มี dead tuple) — เก็บ ~3 partition ล่าสุด |
| `telemetry_agg_hourly`  | 1–2 ปี ขึ้นไป               | ตารางเล็ก ไม่ต้อง partition             |
| `telemetry_agg_daily`   | หลายปี / ไม่จำกัด           | จิ๋ว                                    |
| `telemetry_agg_monthly` | ไม่จำกัด                    | จิ๋วมาก                                 |

หมายเหตุ partition ของ `telemetry_raw`:
- partitioned table ต้องมี partition key อยู่ใน PK → PK = `(id, recorded_at)`
- ⚠️ **insert ที่ตกในช่วงที่ยังไม่มี partition = hard fail ทันที** (ไม่ตกลง table เปล่า) — พลาดสร้าง partition เดือนถัดไป = ingest หยุดเขียนตอนเที่ยงคืนวันที่ 1 กันด้วย 2 ชั้น: **(1) pre-create เดือนปัจจุบัน+ถัดไป** (รันตอน startup + เป็นรอบ) **(2) DEFAULT partition** เป็น catch-all + alert ถ้ามี row ตกลง default จริง
- partition เดือนปัจจุบัน + DEFAULT ต้องมีตั้งแต่ตอน migrate ([ticket 06](./tickets/06-telemetry-raw-and-worker-batch-write.md)) — job pre-create/drop เต็มรูปแบบอยู่ [ticket 12](./tickets/12-retention-partition-90day-drop.md) (Phase 5 hardening ของ [implementation-plan.md](./implementation-plan.md))
- ค่าจริงของ retention ยืนยันกับ product (ผูกกับ open question ใน [architecture.md](./architecture.md#4-สิ่งที่ยังต้องตัดสินใจ-open-questions))
- ทางเลือกที่ตรงงานนี้สุดถ้าไม่อยากดูแลเอง: **TimescaleDB** (hypertable + continuous aggregates จัดการ rollup ทุก interval + retention อัตโนมัติ ไม่ต้องมีตาราง agg เอง) — upgrade path ไม่ใช่เฟสแรก

## 6. `agg_watermark` (internal, ตัวเล็ก)

ให้ aggregator รู้ว่า rollup ถึง bucket ไหนแล้ว จะได้ไม่ scan raw ซ้ำทั้งตาราง

| column        | type        | note                                         |
| ------------- | ----------- | --------------------------------------------- |
| tier          | VARCHAR(10) PK | `hourly` / `daily` / `monthly`             |
| last_bucket   | TIMESTAMPTZ | bucket ล่าสุดที่ agg เสร็จแล้ว                |
| updated_at    | TIMESTAMPTZ | default NOW()                                |

รอบถัดไป aggregator อ่านเฉพาะ raw/ชั้นล่างที่ `> last_bucket` เท่านั้น (ทางเลือกที่ง่ายกว่าถ้าไม่อยากมี table นี้: คำนวณ watermark จาก `MAX(bucket_start)` ของตาราง agg เอง)

## 7. ความสัมพันธ์กับ schema เดิม

- ไม่แก้ `users` table — telemetry ไม่ผูกกับ user โดยตรง สิทธิ์เข้าถึง (ใครดู site ไหนได้) ถ้าต้องทำ multi-tenant ค่อยเพิ่มตาราง `user_site_access` ภายหลัง (ดู open question ใน architecture.md)
- role ที่มีอยู่ (`ADMIN`, `STAFF`, `CUSTOMER`) ใช้คุม endpoint ใหม่ผ่าน permission layer (`resource:action`) — ดู [permissions.md](./permissions.md)
