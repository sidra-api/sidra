# Gunakan base image untuk Golang
FROM golang:1.23 AS builder

# Set working directory di dalam container
WORKDIR /app

COPY . .

RUN mkdir -p /app/bin/plugins

RUN mkdir -p /usr/local/bin

RUN for dir in ./plugins/*; do \
    if [ -d "$dir" ]; then \
        echo "Building $(basename $dir)..."; \
        cd $dir && go mod tidy && go build -o /usr/local/bin/plugin_$(basename $dir); \
        cd -; \
    fi; \
done


RUN go mod tidy && go build -o /usr/local/bin/sidra main.go

# Jalankan binary
ENTRYPOINT [ "sidra" ]
