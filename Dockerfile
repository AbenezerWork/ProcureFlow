FROM golang:1.25.0-bookworm AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/procureflow-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/procureflow-api /app/procureflow-api

EXPOSE 8080

ENTRYPOINT ["/app/procureflow-api"]
