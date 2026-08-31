# 13: Logout token revocation + Redis wiring

**What to build:** logout ทำให้ access token เดิมใช้ไม่ได้ทันที (ไม่ใช่รอหมดอายุ 15 นาที) — wire Redis client ที่ตอนนี้ถูก comment ค้างใน composition root, blacklist access token ตอน logout แล้วให้ JWT middleware เช็ค blacklist ทุก request รวมถึงย้าย rate-limiter storage ไป Redis (ทางเลือก)

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

อ้างอิง: [auth-user.md](../auth-user.md) gap D (+ D-related comments ใน `main.go` / `JWTMiddleware` / auth usecase)

- [ ] Redis client ถูกสร้างจริงใน composition root ตาม config (ไม่ comment ค้าง), degrade ได้ถ้า Redis ล่ม (ไม่ทำ app ตาย)
- [ ] logout ใส่ access token ลง blacklist ด้วย TTL = เวลาที่เหลือของ token
- [ ] `JWTMiddleware` ปฏิเสธ token ที่อยู่ใน blacklist → 401
- [ ] token ที่ blacklist หมดอายุแล้วถูกลบเอง (TTL) ไม่ค้างใน Redis
- [ ] (ทางเลือก) rate-limiter ใช้ Redis storage แทน in-memory เพื่อให้ทำงานข้ามหลาย instance
- [ ] มี test: logout แล้วเรียก endpoint ที่ต้อง auth ด้วย token เดิม → 401
