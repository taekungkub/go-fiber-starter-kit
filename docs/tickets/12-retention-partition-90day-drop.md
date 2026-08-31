# 12: Retention — partition management + 90-day drop

**What to build:** ควบคุมขนาด `telemetry_raw` ตาม retention 90 วัน — job สร้าง partition รายเดือนล่วงหน้า (กัน insert ตกร่อง) และ job ลบ partition ที่เก่ากว่า 90 วันด้วย `DROP`/`DETACH PARTITION` (instant ไม่มี dead tuple) เก็บราว 3 partition ล่าสุด

**Blocked by:** 06 (`telemetry_raw` ต้องเป็น partitioned table แล้ว)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §5](../data-model.md), [implementation-plan.md Phase 5](../implementation-plan.md)

- [ ] job สร้าง partition ของเดือนถัดไปไว้ล่วงหน้า (idempotent)
- [ ] job ลบ/detach partition ที่เก่ากว่า 90 วัน
- [ ] insert ที่ตกในเดือนที่ยังไม่มี partition ไม่ error (มี default/สร้างทัน)
- [ ] verify: หลังรัน drop เหลือเฉพาะ partition ในกรอบ 90 วัน
