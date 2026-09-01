# 12: Retention — partition management + 90-day drop

**What to build:** ควบคุมขนาด `telemetry_raw` ตาม retention 90 วัน — job สร้าง partition รายเดือนล่วงหน้า (กัน insert ตกร่อง) และ job ลบ partition ที่เก่ากว่า 90 วันด้วย `DROP`/`DETACH PARTITION` (instant ไม่มี dead tuple) เก็บราว 3 partition ล่าสุด

**Blocked by:** 06 (`telemetry_raw` ต้องเป็น partitioned table + มี partition เดือนปัจจุบันแล้ว — ดู 06)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §5](../data-model.md), [implementation-plan.md Phase 5](../implementation-plan.md)

> **ทำไม risk สูง:** partitioned table ของ Postgres ถ้า insert ตกในช่วงที่**ยังไม่มี partition รองรับ** จะ **error ทันที (hard fail)** ไม่ใช่ตกลง table เปล่า — แปลว่าพลาดสร้าง partition เดือนถัดไป = ingest ทั้งสายหยุดเขียน DB ตอนเที่ยงคืนวันที่ 1 ของเดือน ต้องกันด้วย 2 ชั้น (pre-create ล่วงหน้า + DEFAULT partition เป็น safety net)

## Pre-create (กัน insert ตกร่อง)
- [ ] job สร้าง partition **เดือนปัจจุบัน + เดือนถัดไป** แบบ idempotent (`CREATE TABLE IF NOT EXISTS ... PARTITION OF ... FOR VALUES FROM ... TO ...`)
- [ ] job รัน **ตอน startup** (กันกรณี deploy ข้ามเดือน) **และ** เป็นรอบ (เช่น daily) — ไม่รอ cron รายเดือนอย่างเดียว
- [ ] มี **DEFAULT partition** เป็น catch-all กันกรณี pre-create พลาด → insert ไม่ hard-fail แต่ตกลง default แทน
- [ ] monitor/log ถ้า DEFAULT partition **มี row เข้ามาจริง** (แปลว่า pre-create พลาด — ต้อง alert เพราะ retention drop จัดการ default ไม่ได้ตรง ๆ)

## Drop (90 วัน)
- [ ] job ลบ/detach partition ที่เก่ากว่า 90 วัน
- [ ] **ไม่ลบ** partition เดือนปัจจุบัน/ถัดไป (กันลบพลาด) และไม่แตะ DEFAULT
- [ ] verify: หลังรัน drop เหลือเฉพาะ partition ในกรอบ 90 วัน (~3 partition)

## Test
- [ ] จำลอง insert `recorded_at` เป็นเดือนที่ยังไม่มี partition → ยืนยันตกลง DEFAULT ไม่ error
- [ ] รัน pre-create แล้ว insert เดือนถัดไปลง partition ที่ถูกต้อง (ไม่ใช่ default)
