# 04: device_templates — schema + seed + CRUD API

**What to build:** ทะเบียนกลางของชนิดอุปกรณ์ + นิยาม metric — ตาราง `device_templates` (metrics เป็น JSONB), seed อย่างน้อย `power_meter` พร้อมนิยาม metric, และ REST API `GET/POST/PATCH /device-templates` เพื่อให้ **เพิ่ม metric key ใหม่ผ่าน API ได้โดยไม่ต้องแตะ schema**

**Blocked by:** 03 (endpoint คุมสิทธิ์ด้วย `RequirePermission`)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §1](../data-model.md), [api-spec.md](../api-spec.md)

- [ ] migration สร้าง `device_templates` และถูกเรียกตอน startup (ก่อน devices)
- [ ] seed `power_meter` (พร้อม metrics เช่น voltage/current/power/energy + หน่วย, energy ตั้ง `cumulative:true`)
- [ ] `GET /device-templates` + `GET /device-templates/:key` คืน metrics JSONB
- [ ] `POST /device-templates` สร้าง template ใหม่ได้ (permission `device:write`)
- [ ] `PATCH /device-templates/:key` เพิ่ม/แก้ metric ใน `metrics` ได้ (เพิ่ม key ใหม่แล้วสะท้อนใน GET)
- [ ] validation ผ่าน `pkg/core.ValidateStruct`, response ผ่าน envelope เดิม
