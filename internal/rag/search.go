package rag

import (
	"math"
)

func CosineSimilarity(a, b []float32) float32 {

	var dot float32
	var normA float32
	var normB float32

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	return dot / (float32(math.Sqrt(float64(normA))) *
		float32(math.Sqrt(float64(normB))))
}

// func (s *VectorStore) Search(query []float32, topK int) []Document {

// 	type Result struct {
// 		Doc   Document
// 		Score float32
// 	}

// 	var results []Result

// 	for _, doc := range s.Docs {

// 		score := CosineSimilarity(query, doc.Embedding)

// 		results = append(results, Result{
// 			Doc:   doc,
// 			Score: score,
// 		})
// 	}

// 	sort.Slice(results, func(i, j int) bool {
// 		return results[i].Score > results[j].Score
// 	})

// 	var top []Document

// 	for i := 0; i < topK && i < len(results); i++ {
// 		top = append(top, results[i].Doc)
// 	}

// 	return top
// }

func Search(queryEmb []float32, store []Document) string {

	bestScore := float32(-1)
	bestText := ""

	for _, doc := range store {

		score := CosineSimilarity(queryEmb, doc.Embedding)

		if score > bestScore {
			bestScore = score
			bestText = doc.Text
		}
	}

	return bestText
}
