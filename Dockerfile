# Sử dụng chính Dockerfile của hãng làm gốc hoặc build từ source
FROM golang:1.21-alpine AS builder

WORKDIR /app
# Sao chép mã nguồn plugnmeet-server (nếu bạn muốn tự build)
COPY . .
RUN go build -o plugnmeet-server main.go

FROM alpine:latest
WORKDIR /app
# Cài đặt các thư viện cần thiết như ffmpeg cho ghi hình
RUN apk add --no-cache ffmpeg

# Copy file thực thi và cấu hình của bạn vào
COPY --from=builder /app/plugnmeet-server .
COPY config.yaml .

EXPOSE 3000
CMD ["./plugnmeet-server", "-config", "config.yaml"]
