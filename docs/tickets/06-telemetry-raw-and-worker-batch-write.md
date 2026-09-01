# 06: telemetry_raw (long, partitioned) + real worker batch write

**What to build:** เก็บข้อมูลดิบจริงลง DB — ตาราง `telemetry_raw` แบบ long (1 row = 1 metric, partition รายเดือนตาม `recorded_at`) และเปลี่ยน worker จากที่ตอนนี้แค่ `fmt.Println` เป็น bulk insert จริง โดยใช้ mock ingest เดิมเป็นตัวป้อนข้อมูลเพื่อพิสูจน์ว่าข้อมูลไหลลงตารางจริง

**Blocked by:** 05 (raw อ้าง device_id + อัปเดต `last_seen_at`; ต้องมี devices ก่อนถึงจะ demo ได้ครบ)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §3](../data-model.md), [implementation-plan.md Phase 1](../implementation-plan.md)

- [ ] migration สร้าง `telemetry_raw` (long: device_id/metric/value/recorded_at/ingested_at) เป็น partitioned table (PK รวม `recorded_at`), index `(device_id, metric, recorded_at)`
- [ ] สร้าง partition **เดือนปัจจุบัน + DEFAULT partition** พร้อมตอน migrate — ต้องมี partition รองรับ**ก่อน**เริ่ม ingest ไม่งั้น insert แรกจะ hard-fail (job pre-create/drop เต็มรูปแบบอยู่ [ticket 12](./12-retention-partition-90day-drop.md))
- [ ] `worker.Job` เก็บ metric rows ที่ parse แล้ว (fan-out) ไม่ใช่ `[]byte` ดิบ
- [ ] worker ทำ bulk insert จริง + flush ตาม batch size **และ** timer (ไม่ค้างเมื่อ throughput ต่ำ)
- [ ] batch size / flush interval เป็น config
- [ ] mock ingest ทำงานแล้วเห็น row จริงใน `telemetry_raw`
- [ ] insert error ถูก log พร้อม payload ที่ fail (ไม่ทำ worker ตาย)
- [ ] อัปเดต `devices.last_seen_at` (throttle ได้)
- [ ] เพิ่ม env ใหม่ (`WORKER_BATCH_SIZE`, `WORKER_FLUSH_INTERVAL`) ใน `config.Config` + `.env.example`
