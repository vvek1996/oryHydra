# Kubernetes Deployment Plan for Ory Hydra and Ory Kratos Stack

This plan outlines the creation of Dockerfiles, Kubernetes manifests, and configurations to deploy the Ory Hydra & Kratos stack in a Kubernetes cluster using a split-service topology while keeping the Go backend codebase completely unmodified.

## User Review Required

> [!NOTE]
> All services run in dedicated Pods/Deployments. To keep the Go backend code completely unmodified (since it expects Kratos and Hydra to be on `localhost`), we run a lightweight Nginx proxy sidecar inside the Go Backend Pod. This proxy routes local loopback calls to cluster service endpoints.

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

#### [NEW] [postgres.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/postgres.yaml)
- **Deployment & Service**: Runs PostgreSQL `16-alpine`.
- **ConfigMap**: Exposes database schemas setup from `postgres.sql` to initialize `hydra` and `kratos` databases on launch.
- **PersistentVolumeClaim**: Storage volume for Postgres.

#### [NEW] [kratos.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/kratos.yaml)
- **Deployment & Services**: Deploys Ory Kratos and exposes Public (`4433`) and Admin (`4434`) cluster services.
- **ConfigMap**: Contains `kratos.yml` (configured to connect to postgres service) and `identity.schema.json`.
- **InitContainer**: Runs migrations automatically before Kratos starts.

#### [NEW] [hydra.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/hydra.yaml)
- **Deployment & Services**: Deploys Ory Hydra and exposes Public (`4444`) and Admin (`4445`) cluster services.
- **InitContainer**: Runs migrations automatically before Hydra starts.
- **Job**: Automatically registers/imports the OAuth2 client (`36d0db37-f52e-46b6-bf1d-3923fc9cf46d`) once Hydra is ready.

#### [NEW] [backend.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/backend.yaml)
- **Deployment & Service**: Runs the Go backend container alongside an Nginx loopback proxy sidecar container.
- **ConfigMap**: Configures Nginx to forward `localhost` ports (`4433`, `4444`, `4445`) to Kratos/Hydra cluster services.

#### [NEW] [frontend.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/frontend.yaml)
- **Deployment & Service**: Runs the React frontend client on port `3000`.

#### [NEW] [traefik.yaml](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s/traefik.yaml)
- **ConfigMap**: Contains `traefik.yml` and `routes.yml` mapping external paths to the cluster DNS services (e.g. `http://hydra-backend-go:4000`).
- **Deployment & Service**: Exposes the Traefik entrypoint on LoadBalancer port `8080`.

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
4. Test the OAuth2 flow by navigating to `http://localhost:8080/` in your browser.

### Deleting the Deployment
To remove all deployed Kubernetes resources, run:
```bash
kubectl delete -f ./k8s
```
