# Sidra using portal

Signup to https://portal.sidra.id

## Standalone
```
dataplanid=UUID ./sidra
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
    helm upgrade --install sidra sid/sidra --set dataplaneid=UUID
    ```

## Docker

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

## Example plugin

```yaml
GatewayService:
  host: "test.sh:8080"
Routes:
  - methods: "GET,POST"
    upstream_host: "localhost"
    upstream_port: "8081"
    path: "/api"
    path_type: "prefix"
    plugins: ""
Plugins:
  - name: "example-jwt"
    type_plugin: "jwt"
    enabled: 1
    config: '{"key":"value"}'

```

---

