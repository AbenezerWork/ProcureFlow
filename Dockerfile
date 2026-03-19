FROM golang:1.25.0-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/procureflow-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/procureflow-migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/procureflow-api /app/procureflow-api
COPY --from=builder /out/procureflow-migrate /app/procureflow-migrate

EXPOSE 8080

ENTRYPOINT ["/app/procureflow-api"]
