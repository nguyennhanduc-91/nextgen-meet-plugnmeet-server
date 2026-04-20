# Lấy bản build sẵn mới nhất của hãng
FROM mynaparrot/plugnmeet-server:latest

WORKDIR /app

# Copy file cấu hình của bạn vào thư mục /app
COPY config.yaml .

EXPOSE 3000

# Chạy lệnh trực tiếp (không dùng ./)
CMD ["plugnmeet-server", "-config", "config.yaml"]
