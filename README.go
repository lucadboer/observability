# Serviço de Previsão do Tempo por CEP

Este é um serviço escrito em Go que recebe um CEP como entrada, consulta a localização associada a esse CEP e, em seguida, busca a previsão do tempo para essa localização. Ele retorna a temperatura atual em Celsius, Fahrenheit e Kelvin juntamente com o nome da cidade.

## Funcionalidades

- Recebe um CEP válido via POST.
- Consulta a localização associada ao CEP usando a API ViaCEP.
- Busca a previsão do tempo para a cidade usando a API WeatherAPI.
- Retorna a temperatura atual em Celsius, Fahrenheit e Kelvin juntamente com o nome da cidade.
- Implementa OpenTelemetry para rastreamento distribuído entre os serviços.

## Tecnologias Utilizadas

- **Go**: Linguagem de programação utilizada para desenvolver o serviço.
- **Gin**: Framework web utilizado para criar o servidor HTTP.
- **Resty**: Biblioteca utilizada para fazer requisições HTTP de forma simples.
- **OpenTelemetry**: Ferramenta utilizada para rastreamento distribuído entre os serviços.
- **Docker**: Utilizado para facilitar a implantação e execução dos serviços em contêineres.

## Como Rodar Localmente

Certifique-se de ter o Docker instalado na sua máquina.

1. Clone este repositório:

2. Crie um arquivo `.env` na raiz do projeto service-b e adicione as seguintes variáveis de ambiente:

```bash
WEATHER_API_KEY=sua-chave-de-api-do-weatherapi
```

3. Execute o seguinte comando na raiz do projeto para iniciar os serviços:

```bash
docker-compose up --build
```

Isso iniciará os serviços `service-a` e `service-b` em contêineres Docker.

4. Agora você pode enviar uma solicitação POST para `localhost:8081/cep` com o seguinte corpo JSON:

```json
{
  "cep": "29902555"
}
```