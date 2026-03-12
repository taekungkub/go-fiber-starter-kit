package rag

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

func EmbedText(client *openai.Client, text string) ([]float32, error) {

	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.AdaEmbeddingV2,
			Input: text,
		},
	)

	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}
