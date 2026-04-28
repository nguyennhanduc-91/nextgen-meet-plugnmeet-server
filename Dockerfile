# Stage 1: Build the custom Go binary
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git build-base

WORKDIR /build

# Clone the official repo
RUN git clone https://github.com/mynaparrot/plugnmeet-server.git .

# Copy our custom OpenAI provider files into the source tree
COPY pkg/insights/providers/openai ./pkg/insights/providers/openai
COPY pkg/services/insights/insights_service.go ./pkg/services/insights/insights_service.go

# Build the custom binary
RUN go build -o plugnmeet-server main.go

# Stage 2: Create the final image
FROM alpine:3.19

RUN apk update && apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the custom compiled binary from builder
COPY --from=builder /build/plugnmeet-server /usr/local/bin/plugnmeet-server

# Copy configuration
COPY config.yaml .

EXPOSE 8080

CMD ["plugnmeet-server", "-config", "config.yaml"]
