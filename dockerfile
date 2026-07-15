FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o api main.go models.go database.go auth.go handlers.go errors.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api .

EXPOSE 8080
CMD ["./api"]