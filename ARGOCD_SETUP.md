# ArgoCD Installation & Local Deployment Guide

This guide describes how to install ArgoCD in your local Kubernetes cluster and deploy the Go web server using GitOps.

---

## 📋 Prerequisites
* **Docker Desktop** (with Kubernetes enabled)
* **kubectl** and **Git** command-line tools installed

---

## 🛠️ Step 1: Install ArgoCD

1. **Create the `argocd` Namespace**:
   ```powershell
   kubectl create namespace argocd
   ```

2. **Apply the ArgoCD Installation Manifest**:
   We use the `--server-side` flag to bypass the 262144-byte limit on metadata annotations for the ApplicationSet CustomResourceDefinition (CRD):
   ```powershell
   kubectl apply --server-side -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

3. **Verify All ArgoCD Pods are Running**:
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

## 🤖 Step 3: Automated CI/CD (GitHub Actions)

We have configured a GitHub Actions workflow under `.github/workflows/main.yml`. Every time you push a change to the `argocd-deployment` branch:
1. GitHub Actions automatically builds the Docker image.
2. It pushes the image to GitHub Container Registry (GHCR) as `ghcr.io/vvek1996/oryhydra:sha-<COMMIT_SHA>`.
3. It automatically updates the image tag in `k8s/deployment.yaml` with the new commit SHA and commits the change back to the repository.
4. ArgoCD detects the change in `k8s/deployment.yaml` and **automatically** syncs and deploys the new version to your local cluster.

---

## 🚀 Step 4: Deploy the Application via ArgoCD

1. **Push the Changes**:
   Ensure you have staged, committed, and pushed the new workflow file and updated manifests:
   ```powershell
   git add .
   git commit -m "Add GitHub Actions workflow for automated GitOps deployment"
   git push origin argocd-deployment
   ```

2. **Apply the Application Manifest**:
   Apply the manifest to the cluster to register the app with ArgoCD:
   ```powershell
   kubectl apply -f k8s/argocd-app.yaml
   ```

ArgoCD will now sync the repository, detect your manifests under `k8s/`, and automatically create the deployment.

---

## 🔍 Step 5: Verify the Deployment

1. **Check Status**:
   Go to your ArgoCD dashboard at **[https://localhost:8443](https://localhost:8443)**. You will see the `go-server-app` application appear. Click on it to see the visual tree of running Kubernetes resources.

2. **Access the App**:
   The Go web server service is exposed via a LoadBalancer on port `8085`. Access the health endpoint from your host machine at:
   👉 **[http://localhost:8085/health](http://localhost:8085/health)**

3. **Verify Auto-Deployments**:
   To test the automation:
   * Modify the Go code in `cmd/server/main.go` (e.g. change the response message).
   * Commit and push the changes: `git commit -am "test auto-deploy" && git push`
   * Watch GitHub Actions run the build and update the manifest.
   * Watch ArgoCD automatically detect the change and perform a rolling update of the pods.
   * Refresh the page at `http://localhost:8085/health` to see your new changes live!

> [!NOTE]
> Since the Docker image is published to GitHub Container Registry (`ghcr.io`), ensure that your package visibility settings for `oryhydra` under your GitHub Profile -> **Packages** is set to **Public** so that your local Kubernetes cluster can pull it without needing registry pull secrets.
