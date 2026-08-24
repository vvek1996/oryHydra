# Kubernetes Deployment Plan for Ory Hydra and Ory Kratos Stack

This plan outlines the creation of Dockerfiles, Kubernetes manifests, and configurations to deploy the Ory Hydra & Kratos stack in a Kubernetes cluster using a split-service topology while keeping the Go backend codebase completely unmodified.

It also configures the Zot Container Registry to authenticate against our in-cluster Ory OIDC stack, securing OIDC credentials via a Kubernetes Secret and exposing the registry externally over HTTPS (port `8443`) using Traefik.

## User Review Required

> [!IMPORTANT]
> - **Go Backend & Zot Loopback Proxies**: To keep the Go backend code completely unmodified and bypass strict OIDC issuer URL verification, both the Go Backend and the Zot Registry pods run an Nginx loopback proxy sidecar container. These sidecars resolve `localhost` endpoints internally to appropriate cluster service names.
> - **Kubernetes OIDC Secret**: The Zot OIDC credentials (`client-id` and `client-secret`) are stored in a Kubernetes `Secret` and mounted into the container to prevent exposing plain-text credentials.
> - **Traefik HTTPS Gateway**: Traefik is exposed on port `8443` (LoadBalancer) and routes requests directly to the Zot service with SSL/TLS enabled (`tls: {}`).

## Docker Build Commands

Run these commands from the root of the workspace to build the container images:

```bash
# 1. Build the Go Backend container image
docker build -t hydra-backend-go:latest ./hydra-backend-go

# 2. Build the React Client container image
docker build -t hydra-client:latest ./hydra-client
```

---

## Proposed Changes

We will use the existing Dockerfiles and create the manifests inside the `k8s/` directory.

### 1. Dockerization of Custom Services

- **Dockerfile (Go Backend)**: Multi-stage Go build containerizing the application.
- **Dockerfile (React Client)**: Serves static Vite-built frontend files via Nginx configured on port `3000`.

---

### 2. Kubernetes Manifests

#### [MODIFY] [hydra.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/hydra.yaml)
- **Job**: Modified client registration to idempotently register/import BOTH `test-client` and `zot-client` with appropriate callback URIs.

#### [MODIFY] [traefik.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/traefik.yaml)
- **ConfigMap (`traefik-config`)**: Configured HTTPS Secure entrypoint on port `8443`, routing to `zot-service` with TLS enabled.
- **Service**: Exposed LoadBalancer port `8443`.

#### [NEW] [zot.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/zot.yaml)
- **Secret (`zot-oidc-secret`)**: Stores the OIDC credentials securely.
- **ConfigMap (`zot-config` & `zot-proxy-config`)**: Holds Zot registry configuration and Nginx proxy sidecar routing settings.
- **Deployment**: Runs the full UI-enabled Zot registry container and the Nginx loopback proxy sidecar.
- **Service**: Exposes Zot Registry internally on port `5000`.

---

## Verification Plan

### Manual Verification
1. Build the Docker containers using the commands listed in **Docker Build Commands**.
2. Apply the K8s configurations:
   ```bash
   kubectl apply -f ./k8s
   ```
3. Verify that all components are running:
   ```bash
   kubectl get pods
   ```
4. Verify OIDC registration was successful:
   ```bash
   kubectl logs job/hydra-client-registration
   ```
5. Test HTTPS gateway routing to Zot:
   ```bash
   curl.exe -k https://localhost:8443/v2/_catalog
   ```
   *(Expected response: `{"code":"UNAUTHORIZED","message":"authentication required"...}`)*

### Deleting the Deployment
To remove all deployed Kubernetes resources, run:
```bash
kubectl delete -f ./k8s
```
