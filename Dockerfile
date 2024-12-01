# Gunakan base image untuk Golang
FROM golang:1.23 AS builder

# Set working directory di dalam container
WORKDIR /app

# Copy semua file source sidra-plugins-hub ke container
COPY . .

# Build binary untuk sidra-plugins-hub
RUN go mod tidy && go build -o sidra-plugins-hub main.go

# Gunakan image minimal untuk hasil akhir
FROM alpine:latest

# Copy binary dari stage builder ke stage ini
COPY --from=builder /app/sidra-plugins-hub /usr/local/bin/sidra-plugins-hub

# Jalankan binary
ENTRYPOINT ["/usr/local/bin/sidra-plugins-hub"]
