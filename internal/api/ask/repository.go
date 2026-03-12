package ask

import (
	"go-fiber-stater-kit/internal/rag"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Repository interface {
	Create(dto *CreateChatDTO) (string, error)
}

type repository struct {
	Client *openai.Client
}

func NewRepository(client *openai.Client) Repository {
	return &repository{Client: client}
}

func (r *repository) Create(dto *CreateChatDTO) (string, error) {

	// docs := []string{
	// 	"พลังงานแสงอาทิตย์คือพลังงานที่ได้จากดวงอาทิตย์",
	// 	"พลังงานลมใช้กังหันลมในการผลิตกระแสไฟฟ้า",
	// 	"สมาร์ตกริดคือระบบโครงข่ายไฟฟ้าดิจิทัลที่ช่วยจัดการพลังงานไฟฟ้าอย่างมีประสิทธิภาพ",
	// }

	dataset := []Weather{
		{30, 85, 1008, 10, true, true},
		{33, 60, 1012, 8, false, false},
		{29, 90, 1007, 12, true, true},
		{34, 55, 1013, 6, false, false},
		{31, 80, 1009, 9, true, true},
		{35, 50, 1015, 7, false, false},
	}

	energyDocs := []string{
		"พลังงานแสงอาทิตย์คือพลังงานที่ได้จากดวงอาทิตย์",
		"พลังงานลมใช้กังหันลมในการผลิตกระแสไฟฟ้า",
		"สมาร์ตกริดคือระบบโครงข่ายไฟฟ้าดิจิทัลที่ช่วยจัดการพลังงานไฟฟ้าอย่างมีประสิทธิภาพ",
	}

	weatherDocs := WeatherToDocs(dataset)

	docs := append(energyDocs, weatherDocs...)

	store, _ := rag.BuildVectorStore(r.Client, docs)

	datasetText := strings.Join(weatherDocs, "\n")

	intent := DetectIntent(dto.Msg)

	var prompt string

	switch intent {

	case "dataset":
		prompt = DatasetPrompt(datasetText, dto.Msg)

	case "average":
		prompt = AveragePrompt(datasetText, dto.Msg)

	case "prediction":
		prompt = PredictionPrompt(datasetText, dto.Msg)

	default:
		prompt = ChatPrompt(dto.Msg)

	}

	qEmb, _ := rag.EmbedText(r.Client, prompt)

	context := rag.Search(qEmb, store)

	answer, _ := rag.AskLLM(r.Client, context, prompt)

	return answer, nil
}
