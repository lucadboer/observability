package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type ZipCodeRequest struct {
	Cep string `json:"cep"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}

	r := gin.Default()
	r.POST("/cep", handleCepRequest)

	tp := initTracer()
	otel.SetTracerProvider(tp)

	r.Run(":8081")
}

func handleCepRequest(c *gin.Context) {
	tr := otel.Tracer("service-b")
	ctx, span := tr.Start(c.Request.Context(), "handleCepRequest")
	defer span.End()

	var request ZipCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	cep := request.Cep
	if len(cep) != 8 || !isNumeric(cep) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	log.Printf("Received CEP: %s", cep)

	citySpanCtx, citySpan := tr.Start(ctx, "get-city-name")
	city, err := getCityByZipCode(citySpanCtx, cep)
	citySpan.End()

	if err != nil {
		log.Printf("Error getting city by zip code: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"message": "can not find zipcode"})
		return
	}

	log.Printf("City found: %s", city)

	tempSpanCtx, tempSpan := tr.Start(ctx, "get-city-temp")
	tempC := getTemperature(tempSpanCtx, city)
	tempSpan.End()

	response := WeatherResponse{
		City:  city,
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273,
	}

	c.JSON(http.StatusOK, response)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func getCityByZipCode(ctx context.Context, cep string) (string, error) {
	client := resty.New()
	resp, err := client.R().SetContext(ctx).Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body(), &result)
	if _, ok := result["erro"]; ok {
		return "", fmt.Errorf("CEP not found")
	}

	return result["localidade"].(string), nil
}

func getTemperature(ctx context.Context, cityName string) float64 {
	apiKey := os.Getenv("WEATHER_API_KEY")
	client := resty.New()

	cityName = url.QueryEscape(cityName)

	resp, err := client.R().SetContext(ctx).Get(fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=yes", apiKey, cityName))
	if err != nil {
		log.Println("Erro ao fazer requisição HTTP:", err)
		return 0
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		log.Println("Erro ao decodificar resposta JSON:", err)
		return 0
	}

	current, ok := result["current"].(map[string]interface{})
	if !ok {
		log.Println("Erro: campo 'current' não encontrado na resposta")
		return 0
	}

	tempC, ok := current["temp_c"].(float64)
	if !ok {
		log.Println("Erro: campo 'temp_c' não encontrado ou não é um número")
		return 0
	}

	log.Println("Temperatura obtida com sucesso para", cityName)
	return tempC
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
			semconv.ServiceNameKey.String("service-b"),
		)),
	)
	return tp
}
