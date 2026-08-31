# 09: Aggregator hourly + `telemetry_agg_hourly`

**What to build:** งาน background ที่ roll up `telemetry_raw` เป็นสรุปรายชั่วโมง — ตาราง `telemetry_agg_hourly` (long) + aggregator ที่คำนวณ avg/min/max/sum/first/last/count ต่อ device ต่อ metric ต่อ hour อย่างถูกต้อง และมี watermark กันคำนวณซ้ำทั้งตาราง

**Blocked by:** 06 (ต้องมี raw ให้ roll up)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §4 + §4.1 + §6](../data-model.md), [implementation-plan.md Phase 3](../implementation-plan.md)

- [ ] migration สร้าง `telemetry_agg_hourly` + unique `(device_id, metric, bucket_start)`
- [ ] aggregator รันเป็นรอบ (ticker) เริ่มใน `main.go` คู่กับ worker
- [ ] rollup raw→hourly: `GROUP BY device_id, metric, date_trunc('hour')` → avg/min/max/sum + `value_first`/`value_last` (ตามเวลา) + count
- [ ] upsert `ON CONFLICT DO UPDATE` (รันซ้ำ/ชนกันแล้วค่าไม่เพี้ยน)
- [ ] มี watermark (`agg_watermark` หรือ `MAX(bucket_start)`) เพื่อไม่ scan raw ทั้งตารางทุกรอบ
- [ ] verify ค่าใน hourly ตรงกับที่คำนวณมือจาก raw ตัวอย่าง
