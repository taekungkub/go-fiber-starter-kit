# 01: Fix register role default + user-update authorization

**What to build:** ผู้ใช้ที่สมัครใหม่ได้ role `CUSTOMER` เป็นค่าเริ่มต้น (ไม่ใช่ `ADMIN` ทุกคนอย่างที่เป็นอยู่) และ `PATCH /users/:id` อนุญาตเฉพาะเจ้าของ record หรือ `ADMIN` เท่านั้น พร้อมห้ามผู้ใช้ที่ไม่ใช่ ADMIN แก้ field `role` ของตัวเอง/ผู้อื่น — ปิดช่องยกระดับสิทธิ์

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

อ้างอิง: [auth-user.md](../auth-user.md) gaps A + B

- [ ] register แล้วผู้ใช้ใหม่มี role `CUSTOMER` (ADMIN สร้าง role สูงผ่าน `POST /users` เท่านั้น)
- [ ] non-owner ที่ไม่ใช่ ADMIN เรียก `PATCH /users/:id` ของคนอื่น → 403
- [ ] non-ADMIN แก้ field `role` (ของตัวเองหรือใคร) → ถูกปฏิเสธ
- [ ] owner แก้ข้อมูลตัวเอง (ที่ไม่ใช่ role) ได้ปกติ, ADMIN แก้ได้ทุก record
- [ ] มี test ครอบเคส escalation (customer พยายามตั้งตัวเองเป็น ADMIN → fail)
