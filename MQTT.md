MQTT
↓
Consumer
↓
JobQueue
↓
Worker
↓
Batch (100)
↓
processBatch
↓
PostgreSQL

Producer (MQTT / Mock)
│
▼
JobQueue (buffer 1000)
│
├── Worker 1
├── Worker 2
├── Worker 3
├── Worker 4
└── Worker 5
