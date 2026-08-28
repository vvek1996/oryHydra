# Zot OCI Registry Push & Pull Guide

This guide explains how to authenticate your local Docker CLI and push/pull OCI container images to and from the Zot Registry (`https://zot.localhost:8443`) when OIDC authentication is enabled.

---

## 1. Prerequisites (TLS Configuration)

Since the registry runs locally over a self-signed TLS certificate, your local Docker engine will block pushes and pulls with certificate validation errors by default.

### Configuring Insecure Registries
1. Open **Docker Desktop**.
2. Click the **Settings** (Gear) icon in the top right.
3. Select **Docker Engine** on the left menu.
4. Add `"zot.localhost:8443"` to the `insecure-registries` array:
   ```json
   {
     "insecure-registries": [
       "zot.localhost:8443"
     ]
   }
   ```
5. Click **Apply & restart**.

---

## 2. API Key Authentication (OIDC CLI Bypass)

When Zot is integrated with an OIDC provider (Ory Hydra), you **cannot** perform a `docker login` using your Kratos password directly. Instead, you must authenticate using a Zot-generated **API Key**.

### Generating the API Key
1. Navigate to the Zot Web UI at **`https://zot.localhost:8443/`** in your browser.
2. Click **Login** and authenticate with your credentials:
   * **Email/Username**: `kpvivek196@gmail.com`
   * **Password**: `secretpassword`
3. Click your profile icon in the top right, go to **API Keys** (or User Settings), and click **Generate API Key**.
4. Copy the generated token string (e.g. `zak_...`).

---

## 3. Docker CLI Integration

### A. Login to the Registry
Run the login command using your registered email as the username. When prompted for a password, paste the **API Key** you generated in the UI:

```bash
docker login zot.localhost:8443 -u kpvivek196@gmail.com
# Password: <Paste Zot API Key here>
# Login Succeeded
```

### B. Tag and Push an Image
Tag a local image with the registry host domain and push it:

```bash
# Tag the local image for Zot
docker tag sha256:<image-hash> zot.localhost:8443/test-image:v1

# Push the image to the Zot registry
docker push zot.localhost:8443/test-image:v1
```

### C. Verify Pull
Remove your local image cache and pull the image directly from your Zot registry to verify it functions correctly:

```bash
# Remove local image tags
docker rmi zot.localhost:8443/test-image:v1

# Pull the image from your Zot registry
docker pull zot.localhost:8443/test-image:v1
```

Once pushed, you can view, inspect, and analyze the image directly from the Zot dashboard at `https://zot.localhost:8443/`.
