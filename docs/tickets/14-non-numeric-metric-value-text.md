# 14: ตัดสิน + รองรับ metric ที่ไม่ใช่ตัวเลข (`value_text`)

**What to build:** `telemetry_raw.value` เป็น `NUMERIC` เท่านั้น แต่บาง device อาจส่ง metric ที่เป็น string/bool (เช่น `status: "running"`, `alarm: true`) — ตัดสินแนวทางแล้ว implement

**Blocked by:** 06 (ต้องมี `telemetry_raw` จริงก่อนถึงจะแก้ schema เพิ่ม)

**Status:** needs-decision (ยังไม่ ready-for-agent จนกว่าจะเลือกแนวทาง)

อ้างอิง: [metric-storage.md](../metric-storage.md#trade-off-ที่ต้องรู้-ไม่ฟรี), [data-model.md §3](../data-model.md)

## ทางเลือก (ยังไม่ตัดสิน)

- **A) map เป็นตัวเลข** ที่ต้นทาง (ingest) เช่น `status: "running"` → `1`, `"stopped"` → `0` — ไม่ต้องแก้ schema แต่ต้องมี mapping table/rule ต่อ metric และเสีย semantic เดิม (query กลับเป็น string ไม่ได้ตรง ๆ)
- **B) เพิ่มคอลัมน์ `value_text VARCHAR NULL`** ใน `telemetry_raw` (และ agg tables ถ้าต้องสรุปด้วย เช่น "ค่า status ล่าสุดของ bucket") — เก็บ semantic เดิม แต่เพิ่ม nullable column ทุกตาราง + aggregator ต้องรู้ว่า metric ไหนอ่านจาก `value` ไหนอ่านจาก `value_text`

## Checklist (เติมหลังตัดสินใจ)

- [ ] เลือกแนวทาง (A หรือ B) — ผูกกับ device จริงที่ต้องใช้ ณ ตอนนั้น (ยังไม่มี use case จริงตอนนี้)
- [ ] ถ้า B: migration เพิ่ม `value_text` ใน `telemetry_raw` (+ agg tables ถ้าจำเป็น)
- [ ] `device_templates.metrics` เพิ่ม flag บอกชนิดค่า (เช่น `"value_type":"text"`) ให้ ingest/worker เลือกคอลัมน์ถูก
- [ ] อัปเดต [metric-storage.md](../metric-storage.md) และ [data-model.md](../data-model.md) ให้ตรงกับแนวทางที่เลือก
