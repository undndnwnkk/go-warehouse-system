package main

import (
	"github.com/go-openapi/runtime/middleware"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/openapi/openapi.yaml")
	})

	opts := middleware.SwaggerUIOpts{
		SpecURL: "/openapi.yaml",
		Path:    "swagger",
	}
	sh := middleware.SwaggerUI(opts, nil)

	mux.Handle("/swagger", sh)

	log.Println("Swagger UI available in http://localhost:8085/swagger")
	log.Fatal(http.ListenAndServe(":8085", mux))
}
