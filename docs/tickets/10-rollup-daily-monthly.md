# 10: Daily + monthly layered rollup

**What to build:** ต่อยอด rollup เป็นรายวัน/รายเดือน — ตาราง `telemetry_agg_daily` และ `telemetry_agg_monthly` โดย daily roll up **จาก hourly** และ monthly **จาก daily** (ไม่ scan raw ซ้ำ) พร้อม avg ถ่วงน้ำหนักด้วย sample_count และ first/last ตามเวลา

**Blocked by:** 09 (ต้องมี hourly เป็นชั้นฐานก่อน)

**Status:** ready-for-agent

อ้างอิง: [data-model.md §4.1](../data-model.md)

- [ ] migration สร้าง `telemetry_agg_daily` + `telemetry_agg_monthly` (schema เดียวกับ hourly)
- [ ] daily rollup จาก hourly, monthly rollup จาก daily
- [ ] `value_avg` ชั้นบน = ถ่วงน้ำหนักด้วย `sample_count` (ไม่ใช่ avg-of-avg ตรง ๆ)
- [ ] `value_first`/`value_last` = ค่าของ bucket ที่เวลาน้อยสุด/มากสุด, min/max/sum/count รวมถูกต้อง
- [ ] upsert idempotent ทั้งสองชั้น
- [ ] verify daily/monthly ตรงกับผลรวมจาก hourly ตัวอย่าง
