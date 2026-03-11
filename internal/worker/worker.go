package worker

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func StartWorkers(db *sqlx.DB, jobQueue <-chan Job, num int) {

	for i := 0; i < num; i++ {
		go worker(db, jobQueue)
	}

}

func worker(db *sqlx.DB, jobQueue <-chan Job) {

	// batch ใช้เก็บ job ชั่วคราวก่อนจะประมวลผลเป็นชุด (batch processing)
	batch := []Job{}

	// วนรับ job จาก queue ไปเรื่อย ๆ
	// range channel จะทำงานจนกว่า channel จะถูกปิด
	for job := range jobQueue {

		// จำลอง delay ของการประมวลผล เช่น network, DB หรือ heavy computation
		// เพื่อทดสอบ behavior ของ queue และ worker
		time.Sleep(2000 * time.Millisecond) // delay 2 วินาที

		// เพิ่ม job เข้าไปใน batch
		batch = append(batch, job)

		// fmt.Println("current batch size:", len(batch))

		// ถ้า batch มีครบ 10 array
		if len(batch) >= 10 {

			// ตรงนี้ปกติจะเป็น logic เช่น
			// - insert ลง database แบบ batch
			// - process message หลายตัวพร้อมกัน
			fmt.Println("Processing batch of", len(batch), "array")

			// หลังจาก process เสร็จให้ reset batch
			// เพื่อเริ่มเก็บ job รอบใหม่
			batch = []Job{}
		}
	}
}
