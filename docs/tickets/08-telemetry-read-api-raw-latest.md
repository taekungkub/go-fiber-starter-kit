# 08: Telemetry read API — `/raw` + `/latest`

**What to build:** endpoint ให้ dashboard ดึงข้อมูลดิบ — `GET /telemetry/raw` (จุดดิบตาม device/metric/ช่วงเวลา, จำกัดความกว้างช่วง) และ `GET /telemetry/devices/:id/latest` (ค่าล่าสุดต่อ metric สำหรับ live tile)

**Blocked by:** 06 (ต้องมีข้อมูลใน `telemetry_raw`)

**Status:** ready-for-agent

อ้างอิง: [api-spec.md](../api-spec.md) (Telemetry)

- [ ] `GET /telemetry/raw` รับ `device_id`, `metric`, `from`, `to` (required) + pagination
- [ ] ช่วง `from/to` กว้างเกินที่กำหนด (เช่น > 24h) → 400 validation error
- [ ] `metric` ไม่อยู่ในนิยาม template ของ device → 400
- [ ] `GET /telemetry/devices/:id/latest` คืนค่าล่าสุดต่อ metric (`DISTINCT ON (metric)`)
- [ ] endpoint คุมสิทธิ์ด้วย `telemetry:read`
