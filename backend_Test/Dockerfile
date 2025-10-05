FROM golang:1.22 as builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/weather ./cmd/server

# TOM: Distroless is a good choice for a small project but it would have been good to know why you chose it specifically in the readme
FROM gcr.io/distroless/base-debian12
ENV PORT=8080 CACHE_TTL_MS=3000 HTTP_TIMEOUT_MS=1500
EXPOSE 8080
COPY --from=builder /out/weather /weather
USER nonroot:nonroot
ENTRYPOINT ["/weather"]
