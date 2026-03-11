package queue

import "go-fiber-stater-kit/internal/worker"

// JobQueue คือ channel ที่ใช้เก็บงาน (Job) เพื่อส่งให้ worker ไปประมวลผล
// make(chan worker.Job, 1000)
//
// chan worker.Job  = channel ที่ส่งข้อมูลประเภท Job
// 1000            = buffer size (เก็บ job ได้สูงสุด 1,000 งาน)
//
// ถ้า queue ยังไม่เต็ม producer (เช่น MQTT consumer)
// สามารถ push job เข้า queue ได้ทันทีโดยไม่ block
//
// ถ้า queue เต็ม goroutine ที่ push จะ block
// จนกว่า worker จะดึง job ออกไป
//
// flow:
// MQTT Consumer → push Job → JobQueue → Worker ดึงไป process
//
var JobQueue = make(chan worker.Job, 1000)
