package ask

import (
	"fmt"
	"strings"
)

// --- DTOs ---

type CreateChatDTO struct {
	Msg string `json:"msg" validate:"required"`
}

type Weather struct {
	Temp         float64
	Humidity     float64
	Pressure     float64
	WindSpeed    float64
	RainToday    bool
	RainTomorrow bool
}

func DatasetToText(data []Weather) string {

	var result string

	result += "Temp,Humidity,Pressure,WindSpeed,RainToday,RainTomorrow\n"

	for _, d := range data {

		result += fmt.Sprintf(
			"%.1f,%.1f,%.1f,%.1f,%t,%t\n",
			d.Temp,
			d.Humidity,
			d.Pressure,
			d.WindSpeed,
			d.RainToday,
			d.RainTomorrow,
		)
	}

	return result
}

func WeatherToDocs(data []Weather) []string {

	var docs []string

	for _, d := range data {

		doc := fmt.Sprintf(
			"Temp %.1fC, Humidity %.1f%%, Pressure %.1f hPa, Wind %.1f m/s, RainToday %t, RainTomorrow %t",
			d.Temp,
			d.Humidity,
			d.Pressure,
			d.WindSpeed,
			d.RainToday,
			d.RainTomorrow,
		)

		docs = append(docs, doc)
	}

	return docs
}

type WeatherResponse struct {
	Temp         float64 `json:"temp"`
	Humidity     float64 `json:"humidity"`
	Pressure     float64 `json:"pressure"`
	WindSpeed    float64 `json:"wind_speed"`
	RainToday    bool    `json:"rain_today"`
	RainTomorrow bool    `json:"rain_tomorrow"`
}

func DetectIntent(msg string) string {

	msg = strings.ToLower(msg)

	switch {

	case strings.Contains(msg, "ค่าแรก"):
		return "dataset"

	case strings.Contains(msg, "ค่าสุดท้าย"):
		return "dataset"

	case strings.Contains(msg, "ค่าเฉลี่ย"):
		return "average"

	case strings.Contains(msg, "ฝน"):
		return "prediction"

	default:
		return "chat"
	}
}

func ChatPrompt(msg string) string {

	text := `
คุณเป็นผู้ช่วย AI smart energy ที่มีความรู้เกี่ยวกับพลังงาน
User Question:
%s

Instructions:
- ใช้ข้อมูลจาก Context เท่านั้น
- ถ้า Context ไม่มีคำตอบ ให้ตอบว่า "ไม่พบข้อมูลในเอกสาร"
- ห้ามใช้ความรู้ทั่วไป
- ตอบเป็นภาษาไทย
- ตอบคำถามของผู้ใช้ตามปกติ
- สามารถอธิบายเพิ่มเติมได้
`

	return fmt.Sprintf(text, msg)
}

func DatasetPrompt(dataset, msg string) string {

	text := `
Weather Dataset:
%s

User Question: %s

ถ้าผู้ใช้ขอข้อมูล dataset เช่น
"ขอค่าแรก"

ให้ตอบเป็น JSON format เท่านั้น

Example JSON:

{
 "temp": 0,
 "humidity": 0,
 "pressure": 0,
 "wind_speed": 0,
 "rain_today": false,
 "rain_tomorrow": false
}

ห้ามตอบข้อความอื่น
`

	return fmt.Sprintf(text, dataset, msg)
}

func PredictionPrompt(dataset, msg string) string {

	text := `
Weather Dataset:
%s

User Question:
%s

ให้วิเคราะห์ข้อมูล Weather Dataset และทำนายว่า "พรุ่งนี้ฝนตกหรือไม่"

Rules:
- ต้องใช้ข้อมูลจาก dataset ในการคำนวณ
- ห้าม copy record จาก dataset ตรงๆ
- ให้คำนวณความน่าจะเป็นจากข้อมูลที่ใกล้เคียง
- ตอบเป็น JSON format เท่านั้น
- ห้ามตอบข้อความอื่น
- คำอธิบาย (reason) ต้องเป็นภาษาไทย

Example JSON:

{
 "rain_probability": 0.0,
 "prediction": "rain | no_rain",
 "reason": "short explanation"
}
`

	return fmt.Sprintf(text, dataset, msg)
}

func AveragePrompt(dataset, msg string) string {

	text := `
Weather Dataset:
		%
	User Question: %s
	Prediction ต้องคำนวณจาก dataset

	Example JSON:

	{
	 "avg_temp": 0,
	 "avg_humidity": 0,
	 "avg_pressure": 0,
	 "avg_wind_speed": 0,
	 "avg_rain_today": false,
	 "avg_rain_tomorrow": false
	}
	ห้ามตอบข้อความอื่น
}
`

	return fmt.Sprintf(text, dataset, msg)
}
