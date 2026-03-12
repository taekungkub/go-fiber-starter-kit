package rag

type Document struct {
	ID        int
	Text      string
	Embedding []float32
}

type VectorStore struct {
	Docs []Document
}

func NewStore() *VectorStore {
	return &VectorStore{
		Docs: []Document{},
	}
}

func (s *VectorStore) Add(doc Document) {
	s.Docs = append(s.Docs, doc)
}
