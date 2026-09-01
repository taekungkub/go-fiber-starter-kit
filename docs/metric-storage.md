# Metric Storage Design

คำถามตั้งต้น: อุปกรณ์แต่ละชนิดมี metric ไม่เหมือนกัน (`power_meter` → voltage/current/power/energy, `water` → flow/volume/pressure, `chiller` → ...) จะเก็บยังไงไม่ให้ schema ระเบิดทุกครั้งที่มี device ชนิดใหม่หรือ metric ใหม่โผล่มา

เอกสารนี้รวมการตัดสินใจที่กระจายอยู่ใน [data-model.md](./data-model.md) (§1, §3, §4) ให้อ่านรวดเดียวจบ

## การตัดสินใจหลัก: long format + JSONB registry (ไม่ใช่ wide/enum)

มีให้เลือก 2 แบบ:

| แบบ | metric ใหม่ = | device ชนิดใหม่ = |
| --- | --- | --- |
| **A) wide** (metric เป็นคอลัมน์, เช่น `voltage NUMERIC, current NUMERIC, ...`) | `ALTER TABLE ADD COLUMN` ทุกตาราง (raw + agg 3 ชั้น) + แก้ struct/query ทุกจุด | คอลัมน์ส่วนใหญ่เป็น NULL (sparse table) |
| **B) long** (metric เป็นค่าใน row, เลือกใช้) | insert row ใหม่ ไม่ต้องแก้ schema | insert 1 row ใน registry table |

**เลือก B** เพราะ device ชนิดต่างกันมี metric set ไม่เท่ากันมาก การทำ wide จะทำให้ตารางกว้างขึ้นเรื่อย ๆ และ schema migration ทุกครั้งที่เพิ่ม metric เป็นต้นทุนที่ไม่คุ้ม

### โครงสร้าง

```
device_templates (registry) ──นิยาม metric ของแต่ละชนิดอุปกรณ์──┐
                                                                  ▼
devices ──(device_template FK)──▶ บอกว่า device นี้มี metric ชุดไหน
   │
   ▼
telemetry_raw (long: 1 row = 1 metric ของ 1 reading)
   │  รายละเอียด: docs/metric-storage.md#telemetry_raw-1-row-1-metric
   ▼
telemetry_agg_hourly ─rollup→ telemetry_agg_daily ─rollup→ telemetry_agg_monthly
```

## 1. `device_templates` — ที่เดียวที่นิยาม metric ของแต่ละชนิดอุปกรณ์

```sql
template_key VARCHAR(50) PK   -- 'power_meter', 'water', 'chiller'
name         VARCHAR(100)
metrics      JSONB            -- [{"key":"voltage","unit":"V","label":"แรงดัน"}, {"key":"energy","unit":"kWh","cumulative":true}, ...]
is_active    BOOLEAN
```

`metrics` เป็น JSONB ไม่ใช่ตารางแยกหรือ enum — เพราะเป็น**ข้อมูล** (metadata ที่ต่างกันตามชนิดอุปกรณ์) ไม่ใช่โครงสร้าง เพิ่ม/แก้ metric key = `UPDATE`/`PATCH` แถวเดียว ไม่แตะ DDL:

```sql
UPDATE device_templates
SET metrics = metrics || '[{"key":"frequency","unit":"Hz"}]'::jsonb
WHERE template_key = 'power_meter';
```

Registry นี้เป็น **source of truth** ให้ 3 ที่อ่านค่าเดียวกัน:
- **ingest**: validate ว่า key ที่ส่งเข้ามาตรงกับที่นิยามไว้ไหม (policy แนะนำ: strict — key ที่ไม่รู้จัก log แล้วข้าม ไม่ auto-insert กันข้อมูลขยะจากพิมพ์ผิด)
- **aggregator**: รู้ว่า metric ไหนเป็น cumulative (เช่น `energy`) ต้องคำนวณ delta แทนการ sum
- **dashboard/API**: อ่านหน่วย (`unit`) และ label ไปแสดงผล โดยไม่ต้อง hardcode ฝั่ง frontend

## 2. `telemetry_raw` — 1 row = 1 metric

```sql
device_id       VARCHAR(100)  -- ref devices.device_id (ไม่บังคับ FK เพื่อ insert เร็ว)
device_template VARCHAR(50)   -- denormalize จาก devices เพื่อ filter/group ต่อชนิดโดยไม่ join
metric          VARCHAR(50)   -- ต้องตรงกับ key ใน device_templates.metrics ของ template นั้น
value           NUMERIC
recorded_at     TIMESTAMPTZ
ingested_at     TIMESTAMPTZ
```

1 reading ที่มี 4 metric (เช่น power_meter ส่ง voltage/current/power/energy พร้อมกัน) = **4 rows** ไม่ใช่ 1 row 4 คอลัมน์ — นี่คือต้นทุนที่ต้องรู้ (ดูหัวข้อ trade-off ด้านล่าง)

Index `(device_id, metric, recorded_at)` ครอบทั้ง query แบบ "metric X ของ device Y ช่วงเวลา" และเป็น scan order ที่ aggregator ใช้

## 3. Aggregate tables — เก็บ stat หลายตัวต่อ bucket ในแถวเดียว (ยังเป็น long ต่อ metric)

`telemetry_agg_hourly/_daily/_monthly` ไม่ pivot metric เป็นคอลัมน์เช่นกัน แต่ 1 row ต่อ `(device_id, metric, bucket_start)` เก็บ `avg/min/max/sum/first/last/sample_count` ให้ client เลือก stat เอง

**metric สะสม (cumulative)** เช่น `energy` — ค่าที่ device ส่งมาเป็นเลขสะสม ไม่ใช่ค่าต่อช่วงเวลา ต้องหา "ใช้ไปเท่าไหร่ในช่วงนี้" ด้วย `value_last - value_first` ไม่ใช่ `value_sum` — โค้ด rollup รู้ว่า metric ไหน cumulative จาก flag `cumulative:true` ที่นิยามไว้ใน `device_templates.metrics` (ข้อ 1) ไม่ hardcode ชื่อ metric ไว้ในโค้ด

## Trade-off ที่ต้องรู้ (ไม่ฟรี)

1. **row amplification**: N metric ต่อ reading = N rows ใน `telemetry_raw` — กระทบ volume/throughput ของ ingest+worker โดยตรง คุมด้วย retention 90 วัน + partition รายเดือน (ดู [ticket 12](./tickets/12-retention-partition-90day-drop.md)) ตารางสรุป (`_hourly/_daily/_monthly`) เล็กมากไม่กระทบ
2. **ต้อง pivot ตอน query** ถ้า dashboard อยากได้หลาย metric ในแถวเดียว/กราฟเดียว (เช่น ตารางเปรียบเทียบ voltage+current+power) — กราฟที่เป็น 1 metric = 1 series ไม่ต้อง pivot
3. **metric ที่ไม่ใช่ตัวเลข** (เช่น status string/bool) — ยังไม่ตัดสิน วันนี้ `value` เป็น `NUMERIC` เท่านั้น ทางเลือกคือ map เป็น 0/1 หรือเพิ่มคอลัมน์ `value_text` แยก (ดู [ticket 14](./tickets/14-non-numeric-metric-value-text.md))

## สรุป: เพิ่ม device ชนิดใหม่ / metric ใหม่ ต้องแก้อะไรบ้าง

| เคส | ต้องแก้ |
| --- | --- |
| device ชนิดใหม่ (เช่นเพิ่ม `chiller`) | insert 1 row ใน `device_templates` พร้อม `metrics` JSONB — จบ |
| metric ใหม่ในชนิดที่มีอยู่แล้ว (เช่นเพิ่ม `frequency` ให้ `power_meter`) | `PATCH /device-templates/:key` (หรือ `UPDATE` ตรง ๆ) — จบ, ไม่ต้องแตะ ingest/worker/aggregator/API |
| metric ไม่ใช่ตัวเลข | **ยังไม่รองรับ** — ต้องตัดสินก่อน (ticket 14) |
