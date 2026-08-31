# EMS Dashboard Monitor — Docs

design docs สำหรับต่อยอด go-fiber-starter-kit เป็นระบบ EMS (Energy Management) dashboard monitor — จุดเริ่มก่อนลงมือ code

## อ่านตามลำดับนี้

1. **[architecture.md](./architecture.md)** — ภาพรวมระบบ, pipeline (MQTT → ingest bridge → queue → worker → DB → aggregator → API), Ingestion Bridge design (§2.1 = หัวใจ), MQTT topic, decisions/open questions
2. **[data-model.md](./data-model.md)** — schema ทั้งหมด (long format): `device_templates`, `devices`, `telemetry_raw`, `telemetry_agg_hourly/daily/monthly`, rollup + retention
3. **[api-spec.md](./api-spec.md)** — REST endpoints: device-templates, devices, telemetry (raw/aggregate/latest)
4. **[permissions.md](./permissions.md)** — authz แบบ `resource:action` (`telemetry:read`, ...) + role matrix
5. **[auth-user.md](./auth-user.md)** — โมดูล auth/user เดิม + ⚠️ **known gaps A–E ต้องแก้ก่อน** (register เป็น ADMIN ทุกคน, PATCH user ไม่มี guard, ฯลฯ)
6. **[implementation-plan.md](./implementation-plan.md)** — แผน 6 phase + ลำดับที่แนะนำ

## Decisions ที่ล็อกแล้ว

- **Ingest = MQTT** (Node-RED → broker), long format (metric-per-row)
- **Storage = long / generic** + `device_templates` registry → เพิ่ม metric/template ไม่ต้องแตะ schema
- **Aggregate = hourly / daily / monthly** (layered rollup), stat ครบต่อ bucket (avg/min/max/sum/first/last/count)
- **Retention `telemetry_raw` = 90 วัน** (partition รายเดือน + drop)

## ยังต้องเคาะ (ก่อน/ระหว่าง code)

- real-time push (WebSocket/SSE) หรือ polling
- multi-tenant / row-level access (CUSTOMER เห็นเฉพาะ site ตัวเอง?)
- metric ที่ไม่ใช่ตัวเลข → ต้องมี `value_text` ไหม

## เริ่ม code ที่ไหน

Phase 0 (migrations: `device_templates` → `devices` → `telemetry_*` + seed 1 template) — ไม่ขึ้นกับ open question ใด ๆ เริ่มได้เลย
