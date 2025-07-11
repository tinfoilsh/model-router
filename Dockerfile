FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin .
EXPOSE 8087
CMD ["/bin"]
