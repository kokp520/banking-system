FROM golang:1.21.8-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# server直接讀取文件
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/api ./api

RUN mkdir -p logs

EXPOSE 8080

CMD ["./main -c ./config/config_release.yaml"]