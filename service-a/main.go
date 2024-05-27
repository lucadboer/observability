package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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
	tr := otel.Tracer("service-a")
	ctx, span := tr.Start(c.Request.Context(), "handleCepRequest")
	defer span.End()

	var request ZipCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	if len(request.Cep) != 8 || !isNumeric(request.Cep) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	jsonPayload, err := json.Marshal(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", "http://service-b:8081/cep", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"message": "can not find zipcode"})
		return
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
	exporter, err := zipkin.New(
		"http://zipkin:9411/api/v2/spans",
	)
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-a"),
		)),
	)
	return tp
}
