package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ericolvr/gin-kafka/internal/infra/akafka"
	"github.com/ericolvr/gin-kafka/internal/infra/repository"
	"github.com/ericolvr/gin-kafka/internal/infra/web"
	"github.com/ericolvr/gin-kafka/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(host.docker.internal:3306/products)")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	msgChan := make(chan *kafka.Message)
	go akafka.Consume([]string{"prodcut"}, "host.docker.internal:9094", msgChan)

	repository := repository.NewProductRepositoryMysql(db)
	createProductUseCase := usecase.NewCreateProductUseCase(repository)
	listProductsUseCase := usecase.NewListProductUseCase(repository)

	productHandlers := web.NewProductHandlers(createProductUseCase, listProductsUseCase)

	r := chi.NewRouter()
	r.Post("/products", productHandlers.CreateProductHandler)
	r.Get("/products", productHandlers.ListProductsHandler)

	http.ListenAndServe(":5500", r)

	for msg := range msgChan {
		dto := usecase.CreateProductInputDto{}
		err := json.Unmarshal(msg.Value, &dto)

		if err != nil {
			// log this
		}

		_, err = createProductUseCase.Execute(dto)
	}
}
