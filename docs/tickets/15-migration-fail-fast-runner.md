# 15: Migration fail-fast + ordered runner ใน main.go

**What to build:** `main.go` ปัจจุบันเรียก `migrations.MigrateTableUsers(db)` โดย **ทิ้ง error ที่ฟังก์ชันคืนมา** (`cmd/main.go:34`) — ถ้า migrate ล้มเหลว (DDL ผิด, partition syntax พัง, permission ไม่พอ) แอปจะ **start ขึ้นปกติแต่ query/insert ตารางนั้นพังหมด** ตรวจไม่พบจนกว่าจะมีคน complain data หาย ต้องเปลี่ยนเป็น fail-fast: migrate พัง = แอปไม่ start

**Blocked by:** none (แก้ได้เลย — ยิ่งทำก่อน 04/05/06 ยิ่งดี เพราะทุกตารางใหม่จะพึ่ง runner นี้)

**Status:** ready-for-agent

อ้างอิง: `cmd/main.go:34`, `migrations/tb_users.go` (pattern `MigrateTableX(db) error`)

> **หมายเหตุ enum:** เหตุที่ data-model.md เลี่ยง Postgres enum ก็เพราะ error ถูกกลืนตรงนี้ (enum `user_role` ที่ประกาศใน `001_init.sql` ไม่ถูกสร้าง runtime แต่ไม่มีใครรู้เพราะ error หาย) — ปิด risk นี้แล้วจะรู้ทันทีถ้า migration/enum พลาด

- [ ] เพิ่ม runner รวม (เช่น `migrations.RunAll(db)` หรือ loop ใน main) ที่เรียก `MigrateTableX` **ตามลำดับ dependency** (users → device_types → devices → telemetry_* → agg_*) และ **`log.Fatal` ทันทีที่ตัวใดคืน error** — ไม่ start Fiber ต่อ
- [ ] เช็ก error ของ `MigrateTableUsers` ที่มีอยู่ (ตอนนี้ถูกทิ้ง) ให้เข้า runner เดียวกัน
- [ ] error log บอกชัดว่า migration **ตัวไหน** พังและ error อะไร
- [ ] ทุก `MigrateTableX` ใหม่ต้องคืน `error` ตาม pattern เดิม (ไม่ swallow ข้างใน)
- [ ] verify: จงใจใส่ DDL ผิดตัวหนึ่ง → แอปต้อง exit non-zero ไม่ start ขึ้น
