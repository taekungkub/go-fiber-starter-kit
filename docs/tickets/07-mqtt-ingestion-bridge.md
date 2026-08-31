# 07: MQTT Ingestion Bridge (paho, จริง)

**What to build:** แทน mock ticker ด้วย MQTT client จริง (paho) ที่ subscribe จาก broker (Node-RED เป็นต้นทาง), parse payload JSON, validate key กับนิยาม metric ของ template, fan-out เป็น metric rows แล้ว enqueue เข้า `queue.JobQueue` — ครบเส้น publish → broker → bridge → queue → worker → DB

**Blocked by:** 06 (worker/ปลายทางเขียน DB ต้องพร้อม), 04 (ต้องมี template ไว้ validate key + resolve metric)

**Status:** ready-for-agent

อ้างอิง: [architecture.md §2.1](../architecture.md) (Ingestion Bridge design), [implementation-plan.md Phase 2](../implementation-plan.md)

- [ ] เชื่อม broker ตาม config (broker url / user / pass / client_id / topic pattern) — client_id คงที่ต่อ instance
- [ ] subscribe `ems/+/+/telemetry`, แยก site_id/device_id จาก topic
- [ ] cache device→template ใน memory (ไม่ query DB ต่อ message)
- [ ] fan-out payload → metric rows, validate key กับ `device_templates.metrics` (key แปลก → log/ข้ามเฉพาะ key)
- [ ] payload/parse พังหรือ device ไม่รู้จัก → log + ข้าม ไม่ทำ loop ตาย
- [ ] auto-reconnect + resubscribe เมื่อหลุด broker
- [ ] backpressure ชัดเจนเมื่อ queue เต็ม (block หรือ drop+นับ) + toggle `MQTT_MOCK` ให้ dev รันได้ไม่มี broker
- [ ] demo: publish จริงขึ้น broker แล้วเห็น row ใน `telemetry_raw`
- [ ] เพิ่ม env ใหม่ (`MQTT_BROKER_URL`, `MQTT_USERNAME`, `MQTT_PASSWORD`, `MQTT_CLIENT_ID`, `MQTT_TOPIC_PATTERN`, `MQTT_MOCK`) ใน `config.Config` + `.env.example`
- [ ] expose ตัวนับ observability พื้นฐาน (queue_depth, reconnect count, raw_dropped) — log ก่อนได้
