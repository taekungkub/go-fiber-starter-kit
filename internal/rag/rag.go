package rag

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func AskLLM(client *openai.Client, contextText, question string) (string, error) {

	prompt := "Answer using this context:\n" + contextText + "\n\nQuestion:" + question

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func BuildVectorStore(client *openai.Client, docs []string) ([]Document, error) {

	var store []Document

	for _, doc := range docs {

		emb, err := EmbedText(client, doc)
		if err != nil {
			return nil, err
		}

		store = append(store, Document{
			Text:      doc,
			Embedding: emb,
		})
	}

	return store, nil
}
