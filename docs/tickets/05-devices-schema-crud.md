# 05: devices — schema + CRUD API

**What to build:** ทะเบียนอุปกรณ์ที่ส่งข้อมูลเข้ามา — ตาราง `devices` (มี `device_template` อ้าง registry, `last_seen_at`, soft delete) และ `CRUD /devices` พร้อม filter ตาม site/template/active

**Blocked by:** 04 (device อ้าง `device_template` ที่ต้องมีใน registry ก่อน)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §2](../data-model.md), [api-spec.md](../api-spec.md)

- [ ] migration สร้าง `devices` (index: unique device_id, site_id, device_template)
- [ ] `POST /devices` ต้องระบุ `device_template` ที่มีอยู่จริง (ไม่มี → 400)
- [ ] `GET /devices` list + pagination + filter `site_id` / `device_template` / `is_active`
- [ ] `GET /devices/:id`, `PATCH /devices/:id`, `DELETE /devices/:id` (soft delete)
- [ ] endpoint คุมสิทธิ์ด้วย permission `device:read/write/delete`
- [ ] sort ใช้ whitelist (ต่อจาก ticket 02)
