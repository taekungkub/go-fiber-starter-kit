# 11: `GET /telemetry/aggregate` API

**What to build:** endpoint ให้ dashboard ดึง series สรุปสำหรับกราฟ — เลือกตารางตาม `interval` (`hourly`/`daily`/`monthly`) แล้วคืน bucket ที่มี stat ครบ (avg/min/max/sum/first/last/count) ให้ client เลือก plot เอง รวมถึง energy delta (`last-first`)

**Blocked by:** 10 (ต้องมีครบทั้ง hourly/daily/monthly เพื่อรองรับทุก interval)

**Status:** ready-for-agent

อ้างอิง: [api-spec.md](../api-spec.md) (`GET /telemetry/aggregate`)

- [ ] รับ `device_id`, `metric` (รับหลายค่าได้), `interval`, `from`, `to`
- [ ] `interval` ไม่อยู่ใน `hourly|daily|monthly` → 400
- [ ] คืน rows เรียงตาม `bucket_start` พร้อม field `value_avg/min/max/sum/first/last` + `sample_count`
- [ ] metric ที่ `cumulative:true` (energy) — ผลช่วง = `last-first` (คำนวณจาก field ที่คืน หรือให้ `value_delta`)
- [ ] endpoint คุมสิทธิ์ด้วย `telemetry:read`
