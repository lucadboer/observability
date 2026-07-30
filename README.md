# observability

Two Go services behind a Zipkin backend: `service-a` validates a Brazilian zip code (CEP) and
forwards it to `service-b`, which resolves the city and returns the current temperature in Celsius,
Fahrenheit and Kelvin. Both emit OpenTelemetry spans.

Written as the tracing challenge of the Full Cycle Go post-graduate program.

## Read this before cloning

This repository is published as a code sample, and it has known gaps. They are listed here rather
than glossed over:

1. **`docker compose up --build` fails as-is.** `service-b/Dockerfile` ends with `COPY .env .`, but
   `.env` is in `.gitignore` and was never committed. Create `service-b/.env` with a
   [WeatherAPI](https://www.weatherapi.com/) key before building:
   ```bash
   echo 'WEATHER_API_KEY=your-key-here' > service-b/.env
   ```
2. **Spans are not correlated across the two services.** No `TextMapPropagator` is registered and the
   outbound request is not instrumented, so the trace context never reaches `service-b`. Zipkin shows
   two independent traces — one per service — not a single distributed trace. Making this a real
   distributed trace means registering the W3C TraceContext propagator on both sides and wrapping the
   client and Gin handlers with `otelhttp`/`otelgin`.
3. **The OpenTelemetry Collector is deployed but bypassed.** `otel-collector-config.yaml` defines a
   working OTLP-in / Zipkin-out pipeline and `OTEL_EXPORTER_OTLP_ENDPOINT` is set in
   `docker-compose.yaml`, but both services construct a Zipkin exporter pointing straight at
   `zipkin:9411`. The collector receives nothing.

## Stack

Go · Gin · Resty · OpenTelemetry SDK + Zipkin exporter · OpenTelemetry Collector (contrib) · Zipkin ·
Docker Compose

## Running it

```bash
echo 'WEATHER_API_KEY=your-key-here' > service-b/.env    # required, see above
docker compose up --build
```

| Service | Port |
|---|---|
| `service-a` | 8080 |
| `service-b` | 8081 |
| Zipkin UI | 9411 |
| Collector (OTLP gRPC) | 4317 |

Send a CEP to `service-a`:

```bash
curl -X POST http://localhost:8080/cep \
  -H 'Content-Type: application/json' \
  -d '{"cep":"15910000"}'
```

```json
{ "city": "Dirce Reis", "temp_C": 24.3, "temp_F": 75.7, "temp_K": 297.45 }
```

Then open the Zipkin UI at `http://localhost:9411` and search for `service-a` and `service-b`.

`test.http` includes the valid case plus the two rejection paths: a syntactically valid but unknown
CEP, and a 7-digit CEP.

<!-- SCREENSHOT: replace this line with a Zipkin screenshot showing the handleCepRequest and get-cep
     spans. Save it as .github/zipkin-trace.png. Note that today the two services appear as separate
     traces — see gap 2 above. -->

## How the request flows

```
POST /cep  →  service-a :8080
              validates: 8 digits, numeric  →  422 otherwise
              POST http://service-b:8081/cep
                            ↓
              service-b :8081
                ViaCEP        → city name for the CEP  (404 if unknown)
                WeatherAPI    → current temperature in Celsius
                converts      → Fahrenheit, Kelvin
```

Validation is duplicated in both services on purpose: `service-b` does not trust its caller and
re-checks the CEP shape itself.

## Technical decisions worth noting

**Each service declares its own `service.name` resource attribute**, which is what makes them
separately searchable in Zipkin.

**The exporter is wrapped in `WithBatcher`** rather than a simple span processor, so spans are queued
and flushed in batches instead of one HTTP call to Zipkin per span.

**Manual spans, not auto-instrumentation.** `handleCepRequest` and the nested `get-cep` span are
started by hand, so the trace reflects the two logical stages of the handler rather than raw HTTP
timing. The cost of that choice is gap 2 above — hand-rolled spans without a propagator do not cross
the process boundary.

## Notes

Course challenge. The log messages are in Portuguese, as written.

## License

MIT
