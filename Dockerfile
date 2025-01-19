# Gunakan base image untuk Golang
FROM golang:1.23.4-alpine

# Set working directory di dalam container
WORKDIR /app

COPY . .

RUN mkdir -p /app/bin/plugins

RUN mkdir -p /usr/local/bin

# Build all plugins
RUN for dir in /app/plugins/*; do \
    if [ -d "$dir" ]; then \
        echo "Building $(basename "$dir")..."; \
        cd "$dir" && go mod tidy && go build -ldflags="-s -w" -o "/usr/local/bin/$(basename "$dir" | sed 's/^plugin-//')"; \
        cd - || exit; \
    fi; \
done


RUN go mod tidy && go build -ldflags="-s -w" -o /usr/local/bin/sidra main.go

# Jalankan binary
ENTRYPOINT [ "sidra" ]
