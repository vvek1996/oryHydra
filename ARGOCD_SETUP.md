# ArgoCD Installation & Local Deployment Guide

This guide describes how to install ArgoCD and Argo CD Image Updater in your local Kubernetes cluster and deploy the Go web server using GitOps (without committing changes back to your Git repository).

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
1. When you push to the `argocd-deployment` branch, GitHub Actions builds and pushes the image `ghcr.io/vvek1996/oryhydra:latest` to GitHub Container Registry.
2. The **Argo CD Image Updater** running in your cluster checks your container registry every 2 minutes.
3. When it detects a new build of the image, it tells ArgoCD to update the running deployment **in-memory** (in the cluster's active state), triggering a rolling restart.
4. No automated commits are written back to your Git history!

---

## 🚀 Step 4: Deploy the Application

1. **Push the Changes**:
   Ensure you have staged, committed, and pushed the new workflow file and updated manifests:
   ```powershell
   git add .
   git commit -m "Configure Argo CD Image Updater and simplify workflow"
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
   Go to your ArgoCD dashboard at **[https://localhost:8443](https://localhost:8443)**. The application `go-server-app` will appear and sync.

2. **Access the App**:
   The Go web server service is exposed via a LoadBalancer on port `8085`. Access the health endpoint from your host machine at:
   👉 **[http://localhost:8085/health](http://localhost:8085/health)**

3. **Verify Auto-Deployments**:
   * Modify the Go code in `cmd/server/main.go` (e.g. change the response message).
   * Commit and push changes: `git commit -am "test auto-deploy" && git push`
   * Once the GitHub Actions workflow completes, wait up to 2 minutes.
   * Verify the Image Updater logs to see the detection and rollout:
     ```powershell
     kubectl logs -n argocd -l app.kubernetes.io/name=argocd-image-updater
     ```
   * Refresh `http://localhost:8085/health` to see your changes updated automatically!

> [!NOTE]
> Since the Docker image is published to GitHub Container Registry (`ghcr.io`), ensure that your package visibility settings for `oryhydra` under your GitHub Profile -> **Packages** is set to **Public** so that your local Kubernetes cluster can pull it without needing registry pull secrets.
