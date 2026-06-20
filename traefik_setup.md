# Traefik Standalone Setup & Routing Guide

This document describes how to set up **Traefik** as a reverse proxy in an Ubuntu environment without using Docker or Docker Compose. This configuration allows you to route frontend, backend, Ory Kratos, and Ory Hydra traffic through a single port (`80` or `8080`), avoiding CORS issues and ensuring Ory cookies are treated as first-party.

---

## 1. Installation

Run the following commands to download the Traefik binary and move it to your system path:

```bash
# 1. Download the Traefik Linux binary
curl -sL https://github.com/traefik/traefik/releases/download/v2.10.4/traefik_v2.10.4_linux_amd64.tar.gz -o traefik.tar.gz

# 2. Extract the binary
tar -xzf traefik.tar.gz traefik

# 3. Make it executable and move to bin
chmod +x traefik
mv traefik /usr/local/bin/

# 4. Clean up the downloaded tarball
rm traefik.tar.gz
```

---

## 2. Configuration Files

Create a configuration directory:
```bash
mkdir -p /etc/traefik
```

We will define two configurations:
* **Static configuration** (`traefik.yml`) to initialize entrypoints and tell Traefik where to find dynamic routing rules.
* **Dynamic configuration** (`routes.yml`) to define the rules, paths, and local ports for each application.

### A. Static Configuration: `/etc/traefik/traefik.yml`

Create `/etc/traefik/traefik.yml` with the following content:

```yaml
entryPoints:
  web:
    address: ":80" # Change to ":8080" if port 80 is occupied or restricted on your container

providers:
  file:
    filename: "/etc/traefik/routes.yml"
    watch: true # Reloads configuration dynamically on file changes without restarting Traefik

log:
  level: INFO
```

### B. Dynamic Configuration: `/etc/traefik/routes.yml`

Create `/etc/traefik/routes.yml` with the following content to map prefix paths to your running service ports:

```yaml
http:
  routers:
    # 1. Route /.ory/kratos to Kratos Public API (4433)
    kratos-router:
      rule: "PathPrefix(`/.ory/kratos`)"
      entryPoints:
        - web
      service: kratos-service
      middlewares:
        - kratos-strip

    # 2. Route /.ory/hydra to Hydra Public API (4444)
    hydra-router:
      rule: "PathPrefix(`/.ory/hydra`)"
      entryPoints:
        - web
      service: hydra-service
      middlewares:
        - hydra-strip

    # 3. Route /api to Express Backend (4000)
    backend-router:
      rule: "PathPrefix(`/api`)"
      entryPoints:
        - web
      service: backend-service
      middlewares:
        - api-strip

    # 4. Route / to React Frontend (3000)
    frontend-router:
      rule: "PathPrefix(`/`)"
      entryPoints:
        - web
      service: frontend-service

  services:
    kratos-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:4433"
    hydra-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:4444"
    backend-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:4000"
    frontend-service:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:3000"

  middlewares:
    kratos-strip:
      stripPrefix:
        prefixes:
          - "/.ory/kratos"
    hydra-strip:
      stripPrefix:
        prefixes:
          - "/.ory/hydra"
    api-strip:
      stripPrefix:
        prefixes:
          - "/api"
```

---

## 3. Starting Traefik

Run Traefik as a background process:

```bash
traefik --configfile=/etc/traefik/traefik.yml &
```

To stop Traefik, you can use:
```bash
pkill traefik
```

---

## 4. Updates to Service Configurations

After starting Traefik on port `80`, update the base configurations of each component:

### Ory Kratos Configuration (`/etc/kratos/kratos.yml`)
Update Kratos' base URL to point to the Traefik entrypoint:
```yaml
serve:
  public:
    base_url: http://localhost/.ory/kratos/
```

### Ory Hydra Setup (`hydra.sh`)
Update issuer and login/consent/logout URLs:
```bash
export URLS_SELF_ISSUER=http://localhost/.ory/hydra/
export URLS_LOGIN=http://localhost/login
export URLS_CONSENT=http://localhost/consent
export URLS_LOGOUT=http://localhost/api/logout
```

### React Frontend Client Configuration (e.g. `Login.tsx`)
Initialize the SDK pointing to the Traefik path:
```typescript
const ory = new FrontendApi(
  new Configuration({
    basePath: "http://localhost/.ory/kratos",
    baseOptions: { withCredentials: true },
  })
);
```
