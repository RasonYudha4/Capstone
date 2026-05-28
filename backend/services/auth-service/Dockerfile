# stage 1: builder
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

# stage 2: runner
FROM alpine:3.21

WORKDIR /app

RUN adduser -D appuser

COPY --from=builder /app/main .
EXPOSE 8089
USER appuser
CMD ["./main"]
