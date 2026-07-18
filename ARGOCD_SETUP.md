# ArgoCD Installation & Local Deployment Guide

This guide describes how to install ArgoCD and Argo CD Image Updater in your local Kubernetes cluster and deploy the Go web servers using GitOps.

---

## 📋 Prerequisites
* **Docker Desktop** (with Kubernetes enabled)
* **kubectl** and **Git** command-line tools installed

---

## 🛠️ Step 1: Install ArgoCD & Image Updater

1. **Create the `argocd` Namespace**:
   ```powershell
   kubectl create namespace argocd
   ```

2. **Apply the ArgoCD Installation Manifest**:
   We use the `--server-side` flag to bypass the 262144-byte limit on metadata annotations for the ApplicationSet CustomResourceDefinition (CRD):
   ```powershell
   kubectl apply --server-side -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

3. **Install Argo CD Image Updater via Helm**:
   ```powershell
   helm repo add argo https://argoproj.github.io/argo-helm
   helm repo update
   helm upgrade --install argocd-image-updater argo/argocd-image-updater --namespace argocd
   ```

4. **Verify All Pods are Running**:
   ```powershell
   kubectl get pods -n argocd
   ```

---

## 🔑 Step 2: Access the ArgoCD Web UI

1. **Port-Forward the ArgoCD Server**:
   Expose the service on local port `8443`:
   ```powershell
   kubectl port-forward -n argocd svc/argocd-server 8443:443
   ```
   *Keep this terminal window open.*

2. **Retrieve the Initial Admin Password**:
   Run the following PowerShell snippet to pull and decode the default admin password:
   ```powershell
   $pwd = kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}"
   [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($pwd))
   ```

3. **Log In**:
   * Open **[https://localhost:8443](https://localhost:8443)** in your browser.
   * *Note: Bypass the SSL warning by clicking **Advanced** -> **Proceed to localhost (unsafe)**.*
   * **Username**: `admin`
   * **Password**: `zzo85BOcHTwCEpv9`

---

## 🤖 Step 3: Automated CI/CD & Image Updater

We use **Argo CD Image Updater** to perform rolling updates automatically without making commits back to Git:
1. When you push to the `argocd-deployment` branch, GitHub Actions builds the Docker image containing both server binaries and pushes it to GitHub Container Registry (`ghcr.io/vvek1996/oryhydra:latest`).
2. The **Argo CD Image Updater** running in your cluster checks your container registry every 2 minutes.
3. When it detects a new build of the image, it tells ArgoCD to update the running deployments **in-memory**, triggering a rolling restart for both `go-server` and `second-server`.
4. No automated commits are written back to your Git history!

---

## 🚀 Step 4: Deploy the Applications

1. **Push the Changes**:
   Ensure you have staged, committed, and pushed the updated workflow file, Dockerfile, and manifests:
   ```powershell
   git add .
   git commit -m "Configure multiple services and update Dockerfile"
   git push origin argocd-deployment
   ```

2. **Apply the Application Manifest**:
   Apply the manifest to the cluster to register the app with ArgoCD:
   ```powershell
   kubectl apply -f k8s/argocd-app.yaml
   ```

---

## 🔍 Step 5: Verify the Deployment

1. **Check Status**:
   Go to your ArgoCD dashboard at **[https://localhost:8443](https://localhost:8443)**. The application `go-server-app` will appear and sync. You should see both `go-server` and `second-server` deployments running.

2. **Access the Services**:
   Both services are exposed outside the cluster via LoadBalancers:
   * **First Server** (port 8080 target): Exposes the `/health` endpoint on host port `8085`:
     👉 **[http://localhost:8085/health](http://localhost:8085/health)**
   * **Second Server** (port 8081 target): Exposes the `/ok` endpoint on host port `8086`:
     👉 **[http://localhost:8086/ok](http://localhost:8086/ok)**

3. **Verify Auto-Deployments**:
   To test the automation:
   * Modify the Go code in either `cmd/server/main.go` or `cmd/second/main.go`.
   * Commit and push changes: `git commit -am "test multi-deploy" && git push`
   * Once the GitHub Actions workflow completes, wait up to 2 minutes.
   * Verify the Image Updater logs to see the detection and rollout:
     ```powershell
     kubectl logs -n argocd -l app.kubernetes.io/name=argocd-image-updater
     ```
   * Refresh the corresponding endpoint in your browser to see your new changes live!

> [!NOTE]
> Since the Docker image is published to GitHub Container Registry (`ghcr.io`), ensure that your package visibility settings for `oryhydra` under your GitHub Profile -> **Packages** is set to **Public** so that your local Kubernetes cluster can pull it without needing registry pull secrets.
