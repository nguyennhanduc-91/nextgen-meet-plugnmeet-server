# Stage 1: Build the custom Go binary
FROM golang:latest AS builder

WORKDIR /build

RUN export DEBIAN_FRONTEND=noninteractive; \
    apt update && \
    apt install -y git wget build-essential ca-certificates \
    libasound2-dev libopus-dev libopusfile-dev libssl-dev libsoxr-dev

# Clone the official repo
RUN git clone https://github.com/mynaparrot/plugnmeet-server.git .

# Copy our custom OpenAI provider files into the source tree
COPY pkg/insights/providers/openai ./pkg/insights/providers/openai
COPY pkg/services/insights/insights_service.go ./pkg/services/insights/insights_service.go

ENV SPEECHSDK_ROOT=/opt/speechsdk

# Download and extract Speech SDK
RUN mkdir -p "$SPEECHSDK_ROOT" && \
     wget -O SpeechSDK-Linux.tar.gz https://aka.ms/csspeech/linuxbinary && \
     tar --strip 1 -xzf SpeechSDK-Linux.tar.gz -C "$SPEECHSDK_ROOT" && \
     rm SpeechSDK-Linux.tar.gz

# Build the custom binary with CGO enabled
RUN TARGETARCH=$(dpkg --print-architecture) && \
    case "$TARGETARCH" in \
        "amd64") SPEECHSDK_ARCH_DIR="x64" ;; \
        "arm64") SPEECHSDK_ARCH_DIR="arm64" ;; \
        *) echo "Unsupported architecture for Speech SDK: $TARGETARCH"; exit 1 ;; \
    esac && \
    CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH GO111MODULE=on \
    CGO_CFLAGS="-I$SPEECHSDK_ROOT/include/c_api" \
    CGO_LDFLAGS="-L$SPEECHSDK_ROOT/lib/${SPEECHSDK_ARCH_DIR} -lMicrosoft.CognitiveServices.Speech.core" \
    go build -trimpath -ldflags '-w -s -buildid=' -a -o plugnmeet-server main.go

# Stage 2: Create the final image
FROM debian:stable-slim

RUN export DEBIAN_FRONTEND=noninteractive; \
    apt update && \
    apt install --no-install-recommends -y wget ca-certificates libreoffice mupdf-tools \
    libasound2 libssl3 libopus0 libsoxr0 libopusfile0 && \
    apt clean && \
    rm -rf /var/lib/apt/lists/*

# Copy the compiled application
COPY --from=builder /build/plugnmeet-server /usr/bin/plugnmeet-server

# Copy the Speech SDK libraries from the builder stage
COPY --from=builder /opt/speechsdk /opt/speechsdk

# Copy the entrypoint script
COPY --from=builder /build/docker-build/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Copy configuration
WORKDIR /app
COPY config.yaml .

EXPOSE 3000

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["plugnmeet-server", "-config", "config.yaml"]
