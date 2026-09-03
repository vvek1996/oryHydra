# Zot OIDC Integration - Solution C (Subdomain-Based Routing with HTTPS Ory Stack)

This directory contains the Kubernetes manifests for **Solution C**, which deploys a fully secure, subdomain-scoped container registry integrated with Ory Hydra and Kratos running entirely over HTTPS.

---

## 1. Architecture & Routing Flow

Solution C runs the entire user authentication lifecycle and backend APIs over secure HTTPS on port `8443` using a custom SAN certificate:

*   **Registry Domain**: `https://zot.localhost:8443` (subdomain-based)
*   **Ory Stack / UI / Backend**: `https://localhost:8443`

```mermaid
sequenceDiagram
    autonumber
    actor User as Browser
    participant GW as Traefik Gateway (8443 / 8080)
    participant Zot as Zot Pod
    participant Proxy as Nginx Loopback Proxy (8443)
    participant Hydra as Ory Hydra (4444 / 4445)
    participant Kratos as Ory Kratos (4433 / 4434)

    User->>GW: HTTPS https://zot.localhost:8443/
    GW->>Zot: Forward to zot-service (port 5000)
    User->>GW: Click Login -> Redirects to https://localhost:8443/login
    GW->>Kratos: Load secure login forms
    User->>Kratos: Enter credentials (kpvivek196@gmail.com)
    Kratos->>Hydra: Authorize OIDC Request
    Hydra->>User: Redirect to callback URL (https://zot.localhost:8443/zot/auth/callback/oidc)
    User->>GW: Callback with Code & State
    GW->>Zot: Forward Callback
    Zot->>Proxy: Exchange Auth Code (via https://localhost:8443/.ory/hydra/...)
    Proxy->>GW: Forward raw TCP stream to Traefik Service IP
    GW->>Hydra: Validate Token Exchange (TLS Handshake with Custom Cert)
    Hydra->>Zot: Return ID & Access Token
    Zot->>User: 302 Redirect to https://zot.localhost:8443/home (Authenticated)
```

---

## 2. Key Features & Configurations

### A. Traefik HTTP-to-HTTPS Redirection ([`traefik.yaml`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s3/traefik.yaml))
*   Includes a `redirect-to-https` middleware that intercepts any plain HTTP requests on port `8080` and redirects them permanently (301) to HTTPS port `8443`.
*   This prevents insecure cookie drops and solves the Kratos `security_csrf_violation` issue due to protocol mixing.

### B. Ory Stack HTTPS Configs ([`kratos.yaml`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s3/kratos.yaml) & [`hydra.yaml`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s3/hydra.yaml))
*   **Kratos**: Public base URL (`https://localhost:8443/.ory/kratos/`), CORS allowed origins, and return/UI URLs are set strictly to HTTPS.
*   **Hydra**: Issuer is set to `https://localhost:8443/.ory/hydra/`. All redirects (`URLS_LOGIN`, `URLS_CONSENT`, etc.) target HTTPS. Cookies are forced secure (`SERVE_COOKIES_SECURE=true`).

### C. Zot Trust & Loopback Proxy ([`zot/`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s3/zot/))
*   **TCP Stream proxying**: The Nginx sidecar container (`loopback-proxy`) runs a TCP stream proxy on port `8443` to route Zot's local OIDC calls through Traefik (`traefik.traefik.svc.cluster.local:8443`) dynamically.
*   **Direct Certificate Trust**: The `traefik-tls-cert` secret is generated in the `zot` namespace directly from `certs/tls.crt` by Kustomize and mounted to `/etc/ssl/certs/ca-certificates.crt`, allowing Zot to natively trust Traefik over HTTPS without extra sidecars or initContainers.

### D. Namespace Isolation
*   **`traefik` namespace**: Houses the Traefik Gateway deployment, configuration, service, and TLS secret.
*   **`zot` namespace**: Houses the Zot registry deployment, proxy, and configurations.
*   **`default` namespace**: Houses the Ory Stack (Kratos, Hydra, Postgres, Seeding jobs) and Application services.

For continuous rotation and dynamic sync patterns, see **[`docs/tls_rotation_and_sync_strategies.md`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/docs/tls_rotation_and_sync_strategies.md)**.

---

To deploy the components:

```powershell
# Deploy ALL components at once:
kubectl apply -k .

# Delete ALL components at once:
kubectl delete -k .
```

---

## 4. Authentication Options & Verification

Zot is configured with **Dual Authentication**:

### Option A: Sign In with Ory Hydra (OIDC SSO)
1. Open an **Incognito Window** and navigate to `https://zot.localhost:8443/`.
2. Click **Login** in the top right.
3. Click **"Login with Ory Hydra"**.
4. Authenticate using your Kratos credentials:
   * **Email**: `kpvivek196@gmail.com`
   * **Password**: `secretpassword`
5. Click **Consent**. You will be redirected back to `https://zot.localhost:8443/home` authenticated.

### Option B: Local Admin Credentials (UI & Docker CLI)
Used for break-glass registry administration or CI/CD pipelines:

* **Default Admin Account**:
  * **Username**: `admin`
  * **Password**: `adminpassword`

#### CLI Login Example:
```powershell
docker login zot.localhost:8443 -u admin -p adminpassword
```