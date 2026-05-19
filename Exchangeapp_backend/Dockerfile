# 编译阶段
FROM golang:1.26.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN go build -o main ./cmd/server/.

# 运行阶段
FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/main .
COPY ./config/config.yml ./config/

EXPOSE 3000

CMD ["./main"]