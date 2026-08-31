# 03: Permission layer (`resource:action`) + `RequirePermission`

**What to build:** ชั้น authorization แบบ `resource:action` (เช่น `telemetry:read`, `device:write`) — มี `internal/authz` เก็บ role→permission matrix เป็น source of truth และ middleware `RequirePermission(...)` ที่อ่าน role จาก context แล้วอนุญาต/ปฏิเสธ, migrate route ของ auth/user เดิมมาใช้ permission แทน `RequireRole` ตรง ๆ

**Blocked by:** 01 (role ต้องถูกกำหนดถูกก่อน permission ถึงจะมีความหมาย)

**Status:** ready-for-agent

อ้างอิง: [permissions.md](../permissions.md)

- [ ] มี `internal/authz` ที่ประกาศ permission constants + `map[role][]permission` + `HasPermission(role, perm)`
- [ ] `RequirePermission(perms...)` คืน 403 เมื่อ role ไม่มี permission, ผ่านเมื่อมี
- [ ] route user เดิม (`POST/DELETE /users`) ใช้ `RequirePermission(user:write / user:delete)`
- [ ] ยิงด้วย ADMIN/STAFF/CUSTOMER แล้วได้ผลตรงตาม matrix ใน permissions.md
- [ ] มี test ครอบ matrix อย่างน้อย 1 เคสต่อ role
