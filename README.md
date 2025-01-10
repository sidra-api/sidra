# Sidra using portal

Signup to https://portal.sidra.id

## Install Standalone / as Loadbalancer
```
./sidra --config=./config.yaml
or
dataplanid=UUID ./sidra 
or
dataplanid=UUID ./sidra --config=./config.yaml
```

### Environtment Variable

```yaml
dataplaneid: UUID
SSL_ON: false (default false)
SSL_CERT_FILE: /etc/ssl/certs/server.crt
SSL_KEY_FILE: /etc/ssl/certs/server.key
PORT: 8080
SSL_PORT: 8433
```

## Example Config

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
---

## Install on Kubernetes (Helm)

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

## Install on Docker

``` docker pull  ghcr.io/sidra-api/sidra:latest ```

``` docker run   ghcr.io/sidra-api/sidra --rm -p 8080:8080 -e dataplaneid=UUID ```

---

# Sidra using Config File

## Standalone
```
./sidra --config=./config.yaml
```

## Kubernetes

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

## Docker

``` docker pull  ghcr.io/sidra-api/sidra:latest ```

``` docker run   ghcr.io/sidra-api/sidra --rm -p 8080:8080 -v ./config.yaml:/tmp/config.yaml```

---

## Install Custom Plugin

### Step 1. Create Docker file
```Dockerfile
FROM ghcr.io/sidra-api/sidra:latest

COPY YOUR_PLUGIN_BINARY /usr/local/bin/plugin_YOUR_PLUGIN_BINARY

ENTRYPOINT ["sidra"]
```

### Step 2. Build Docker Image and Push to Registry
```bash
docker build -t sidra:latest .
docker tag sidra:latest your_docker_repo:latest
docker push your_docker_repo:latest

```