# EMS Dashboard Monitor — Tickets

tracer-bullet tickets แตกจาก design docs — เรียงตาม dependency (blocker ก่อน) ใช้เป็น checklist track งาน/เทส

## Progress

- [ ] **01** · [Fix register role + user-update authz](./01-fix-register-role-and-user-update-authz.md) — _blocked by: none_
- [ ] **02** · [Refresh-token hash + sort whitelist](./02-refresh-token-hash-and-sort-whitelist.md) — _blocked by: none_
- [ ] **13** · [Logout token revocation + Redis wiring](./13-logout-token-revocation-redis.md) — _blocked by: none_
- [ ] **15** · [Migration fail-fast + ordered runner](./15-migration-fail-fast-runner.md) — _blocked by: none (ทำก่อน 04–06)_
- [ ] **03** · [Permission layer + RequirePermission](./03-permission-layer-require-permission.md) — _blocked by: 01_
- [ ] **04** · [device_templates: schema + CRUD + seed](./04-device-templates-schema-crud-seed.md) — _blocked by: 03_
- [ ] **05** · [devices: schema + CRUD](./05-devices-schema-crud.md) — _blocked by: 04_
- [ ] **06** · [telemetry_raw + worker batch write](./06-telemetry-raw-and-worker-batch-write.md) — _blocked by: 05_
- [ ] **07** · [MQTT Ingestion Bridge](./07-mqtt-ingestion-bridge.md) — _blocked by: 06, 04_
- [ ] **08** · [Telemetry read API (raw/latest)](./08-telemetry-read-api-raw-latest.md) — _blocked by: 06_
- [ ] **09** · [Aggregator hourly](./09-aggregator-hourly.md) — _blocked by: 06_
- [ ] **10** · [Daily + monthly rollup](./10-rollup-daily-monthly.md) — _blocked by: 09_
- [ ] **11** · [/telemetry/aggregate API](./11-telemetry-aggregate-api.md) — _blocked by: 10_
- [ ] **12** · [Retention: partition + 90-day drop](./12-retention-partition-90day-drop.md) — _blocked by: 06_
- [ ] **14** · [ตัดสิน + รองรับ metric ที่ไม่ใช่ตัวเลข (value_text)](./14-non-numeric-metric-value-text.md) — _blocked by: 06, needs-decision_

## เส้นทาง

```
01 ─▶ 03 ─▶ 04 ─▶ 05 ─▶ 06 ─┬─▶ 07  (bridge, +04)
02 (อิสระ)                   ├─▶ 08  (read API)
13 (อิสระ)                   ├─▶ 09 ─▶ 10 ─▶ 11  (aggregate)
15 (อิสระ, ทำก่อน 06)         ├─▶ 12  (retention)
                             └─▶ 14  (value_text, needs-decision)
```

เริ่มได้ทันที (ไม่มี blocker): **01 / 02 / 13 / 15** (15 ควรทำก่อน 04–06 เพราะทุกตารางใหม่พึ่ง migration runner) — หลัง **06** เดินขนานได้: 07 / 08 / 09 / 12

## ยังไม่อยู่ในชุดนี้ (open question)

real-time push (WebSocket/SSE), multi-tenant row-level access, `/dashboard/summary` — ดู [architecture.md §4](../architecture.md). `value_text` สำหรับ metric ที่ไม่ใช่ตัวเลขมี ticket แล้ว (14) แต่ยัง needs-decision
