package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ZipCodeRequest struct {
	Cep string `json:"cep"`
}

func main() {
	r := gin.Default()
	r.POST("/cep", handleCepRequest)

	tp := initTracer()
	otel.SetTracerProvider(tp)

	r.Run(":8080")
}

func handleCepRequest(c *gin.Context) {
	var request ZipCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	if len(request.Cep) != 8 || !isNumeric(request.Cep) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	// Prepare the JSON payload
	jsonPayload, err := json.Marshal(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	resp, err := http.Post("http://service-b:8081/cep", "application/json", bytes.NewBuffer(jsonPayload))

	if err != nil && resp.StatusCode == http.StatusNotFound {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"message": "can not find zipcode"})
		return
	}

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	c.JSON(http.StatusOK, result)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func initTracer() *sdktrace.TracerProvider {
	exp, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint("http://otel-collector:4317"))
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	return tp
}
