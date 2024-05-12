package main

import (
	"fmt"
	"github.com/igorrize/ai_client/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.POST("/fill", func(c echo.Context) error {
		data := c.FormValue("data")
		cohereApiKey := os.Getenv("COHEREAPI")
		words := strings.Fields(data)
		embeddingChunk := 92 //according to cohere docs
		chunks := make([][]string, 0, (len(words)+embeddingChunk-1)/embeddingChunk)
		for i := 0; i < len(words); i += embeddingChunk {
			end := i + embeddingChunk
			if end > len(words) {
				end = len(words)
			}
			chunks = append(chunks, words[i:end])
		}
		es := services.NewCohereEmbeddingService(cohereApiKey)
		var wg sync.WaitGroup
		type vectorWithPayload struct {
			vector  [][]float32
			payload []string
		}
		results := make(chan vectorWithPayload)
		for _, chunk := range chunks {
			wg.Add(1)
			go func(n []string) {
				defer wg.Done()

				embeddings, err := es.CreateEmbedding(chunk)
				if err != nil {
					return
				}
				// transform float 64 to 32
        embeddingsFloat32 := make([][]float32, len(embeddings))
				for i, emb := range embeddings {
					embeddingsFloat32[i] = make([]float32, len(emb))
					for j, val := range emb {
						embeddingsFloat32[i][j] = float32(val)
					}
				}

				results <- vectorWithPayload{
					vector:  embeddingsFloat32,
          payload: chunk,
				}

			}(chunk)
		}
		go func() {
			wg.Wait()
			close(results)
		}()
		vs, err := services.NewQdrantService("localhost:6334")
		if err != nil {
			log.Fatalf("Failed to create Qdrant service: %v", err)
		}
		collectionsCount, err := vs.ListCollections()
		collectionName := fmt.Sprintf("Collecton_%s", len(collectionsCount)+1)
		vectorSize := uint64(1024)
		defaultSegmentNumber := uint64(1)
		vs.CreateCollection(collectionName, vectorSize, 1, defaultSegmentNumber)
		for result := range results {
			vs.UpsertPoints(collectionName, true, result.vector, result.payload)
		}
		return c.Render(http.StatusOK, "data.html", map[string]interface{}{"embeddings": results})
	})
	e.Logger.Fatal(e.Start(":8080"))
}
