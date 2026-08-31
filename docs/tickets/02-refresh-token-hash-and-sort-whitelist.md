# 02: Refresh-token hash consistency + sort whitelist

**What to build:** refresh token ทำงานถูกต้องทุกเส้นทาง — hash ด้วยวิธีเดียวกันทั้งตอน register / login / refresh (ปัจจุบันปนกันระหว่าง bcrypt กับ sha256 ทำให้ refresh หลัง register พังและหลัง login ใช้ได้แค่ครั้งเดียว) และ query param `sort` ผ่าน whitelist ก่อนใส่ลง `ORDER BY` กัน SQL injection

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

อ้างอิง: [auth-user.md](../auth-user.md) gaps C + E

- [ ] register → refresh token ครั้งแรกสำเร็จ
- [ ] login → refresh หลายครั้งติดต่อกันสำเร็จ (ไม่พังหลังครั้งแรก)
- [ ] วิธี hash refresh token เหมือนกันทั้ง register/login/refresh (deterministic, verify ได้)
- [ ] `sort` ที่ไม่อยู่ใน whitelist ของแต่ละ repo ถูก reject หรือ fallback เป็น default column (ไม่ interpolate ค่าดิบ)
- [ ] มี test ยิง `sort=` ด้วยค่าที่เป็น SQL แล้วไม่หลุดเข้า query
