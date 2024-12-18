# Gunakan base image untuk Golang
FROM golang:1.23 AS builder

# Set working directory di dalam container
WORKDIR /app

COPY . .

RUN mkdir -p /app/bin/plugins

RUN for dir in ./plugins/*; do \
    if [ -d "$dir" ]; then \
        echo "Building $(basename $dir)..."; \
        cd $dir && go mod tidy && go build -o /app/bin/plugins/plugin_$(basename $dir); \
        cd -; \
    fi; \
done


RUN go mod tidy && go build -o sidra main.go

# Gunakan image minimal untuk hasil akhir
FROM alpine:latest

# Copy binary dari stage builder ke stage ini
COPY --from=builder /app/sidra /usr/local/bin/sidra
COPY --from=builder /app/bin/plugins/* /usr/local/bin

# Jalankan binary
ENTRYPOINT ["/usr/local/bin/sidra"]
