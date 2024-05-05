package services

import (
	"context"

	cohere "github.com/cohere-ai/cohere-go/v2"
	client "github.com/cohere-ai/cohere-go/v2/client"
)

type EmbeddingService interface {
	CreateEmbedding(texts []string) ([][]float64, error)
}

type CohereEmbeddingService struct {
	client *client.Client
}

func NewCohereEmbeddingService(token string) *CohereEmbeddingService {
	return &CohereEmbeddingService{
		client: client.NewClient(client.WithToken(token)),
	}
}

func (s *CohereEmbeddingService) CreateEmbedding(texts []string) ([][]float64, error) {
	resp, err := s.client.Embed(
		context.TODO(),
		&cohere.EmbedRequest{
			Texts:     texts,
			Model:     cohere.String("embed-english-v3.0"),
			InputType: cohere.EmbedInputTypeSearchDocument.Ptr(),
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.EmbeddingsFloats.Embeddings, nil
}
