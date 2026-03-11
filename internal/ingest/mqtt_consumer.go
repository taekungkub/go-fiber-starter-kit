package ingest

import (
	"go-fiber-stater-kit/internal/queue"
	"go-fiber-stater-kit/internal/worker"
	"log"
	"time"
)

func StartMQTTConsumerMock(interval time.Duration) {

	ticker := time.NewTicker(interval)

	go func() {

		for t := range ticker.C {

			job := worker.Job{
				Payload: []byte(`{"device_id":"meter-001","voltage":210,"current":5,"power":1050,"created_at":` + t.Format("2006-01-02T15:04:05Z07:00") + `}`),
			}

			queue.JobQueue <- job

			log.Println("mock mqtt data:", string(job.Payload))

			// log.Printf(
			// 	"after push queue=%d/%d",
			// 	len(queue.JobQueue),
			// 	cap(queue.JobQueue),
			// )

		}

	}()

}
