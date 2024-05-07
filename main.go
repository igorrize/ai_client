package main

import (
	"github.com/igorrize/ai_client/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"io"
	"net/http"
	"strings"
  "os"
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
		es := services.NewCohereEmbeddingService(cohereApiKey)
		embeddings, err := es.CreateEmbedding(strings.Fields(data))
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "data.html", map[string]interface{}{"embeddings": embeddings})
	})
	e.Logger.Fatal(e.Start(":8080"))
}
