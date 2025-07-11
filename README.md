# Model Router

A lightweight HTTP proxy service that routes model requests to ports on localhost.

### API Endpoints

#### Catch all (/)
- Body: Must include a "model" field specifying which model to route to

#### Metrics Endpoint (/metrics)
- Method: GET
- Query Parameter: model (e.g., /metrics?model=llama3)

## Usage

```bash
docker run -p 8087:8087 ghcr.io/tinfoilsh/model-router -m "llama3-3-70b_8081,deepseek-r1:70b_8082"
```
