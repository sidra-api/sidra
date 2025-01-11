# Sidra Installation

## Via Portal
Signup to https://portal.sidra.id, create Dataplane and get Dataplane UUID

### Standalone 

` dataplanid=UUID ./sidra `

### Install on Kubernetes (Helm)

Here are the steps to install Sidra via Helm chart:

1. **Add the Helm chart repository:**

    ```bash
    helm repo add sid https://sidra-api.github.io/sidra-helm/charts
    ```

2. **Update the Helm chart repository:**

    ```bash
    helm repo update
    ```

3. **Install the Helm chart:**

    ```bash
    helm upgrade --install sidra sid/sidra --set dataplaneid=UUID
    ```

### Install on Docker with dataplaneid

``` docker pull  ghcr.io/sidra-api/sidra:latest ```

``` docker run   ghcr.io/sidra-api/sidra --rm -p 8080:8080 -e dataplaneid=UUID ```

---

# Example Config

```yaml
GatewayService:
  host: "test.sh"
Routes:
  - methods: "GET,POST"
    upstream_host: "localhost"
    upstream_port: "8081"
    path: "/api"
    path_type: "prefix"
    plugins: "example-jwt"
Plugins:
  - name: "example-jwt"
    type_plugin: "jwt"
    enabled: 1
    config: '{"key":"value"}'
```
## Standalone with config file
```
./sidra --config=./config.yaml
or
dataplanid=UUID ./sidra --config=./config.yaml
```

## Kubernetes with config

Here are the steps to install Sidra via Helm chart:

   1. **Add the Helm chart repository:**

   ```bash
   helm repo add sid https://sidra-api.github.io/sidra/charts
   ```

   2. **Update the Helm chart repository:**

   ```bash
   helm repo update
   ```

   3. **Install the Helm chart:**

   ```bash
   helm upgrade --install sidra sid/sidra  --set-file sidra.config=config.yaml
   ```

## Docker with config

   ``` docker pull  ghcr.io/sidra-api/sidra:latest ```

   ``` docker run   ghcr.io/sidra-api/sidra --rm -p 8080:8080 -v ./config.yaml:/tmp/config.yaml```

---

# Environtment Variable

## Default
```yaml
dataplaneid: UUID
PORT: 8080
```

## SSL ON (Install as Load Balancer)
```yaml
SSL_ON: false (default false)
SSL_CERT_FILE: /etc/ssl/certs/server.crt
SSL_KEY_FILE: /etc/ssl/certs/server.key
SSL_PORT: 8433
```
---

# Plugin

## Install Custom Plugin on Docker

### Step 1. Create Docker file
```Dockerfile
FROM ghcr.io/sidra-api/sidra:latest

COPY YOUR_PLUGIN_BINARY /usr/local/bin/YOUR_PLUGIN_BINARY

ENTRYPOINT ["sidra"]
```

### Step 2. Build Docker Image and Push to Registry
```bash
docker build -t sidra:latest .
docker tag sidra:latest your_docker_repo:latest
docker push your_docker_repo:latest

```

## Install Custom Plugin on Standalone

### Step 1. Build Plugin
```bash
cd YOUR_PLUGIN_GO_PROJECT
go build -o YOUR_PLUGIN_BINARY .
cp YOUR_PLUGIN_BINARY /usr/local/bin/YOUR_PLUGIN_BINARY
```

### Step 2. Register to config.yaml
```yaml
Plugins:
  - name: "plugin_implementation_name"
    type_plugin: "YOUR_PLUGIN_BINARY"
    enabled: 1
    config: '{"key":"value"}'
```

or 

### Register Plugin via https://portal.sidra.id 

as admin access menu plugin_types, then create new plugin_type