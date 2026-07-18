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
   * **Password**: *[The decoded password from above]*
psd: zzo85BOcHTwCEpv9
---

## 📦 Step 3: Build the Application Docker Image

Build the local Docker container for the Go web server. Docker Desktop shares its daemon, making this image accessible to your Kubernetes cluster:
```powershell
docker build -t go-server:latest .
```

---

## 🚀 Step 4: Push Manifests to Git

Since ArgoCD uses GitOps, it requires the files to be pushed to your remote repository so it can sync them:

1. **Stage and Commit the Manifests**:
   ```powershell
   git add Dockerfile k8s/
   git commit -m "Add Dockerfile and Kubernetes manifests for ArgoCD deployment"
   ```

2. **Push the Branch**:
   ```powershell
   git push -u origin argocd-deployment
   ```

---

## 📡 Step 5: Deploy the Application via ArgoCD

Apply the ArgoCD Application manifest. This tells ArgoCD to sync the manifests from the `k8s/` folder in the `argocd-deployment` branch of your Git repo:
```powershell
kubectl apply -f k8s/argocd-app.yaml
```

ArgoCD will automatically create the `go-server` Deployment and Service in the `default` namespace.

---

## 🔍 Step 6: Verify the Deployment

1. **Check Status**:
   Go to your ArgoCD dashboard at **[https://localhost:8443](https://localhost:8443)**. You will see the `go-server-app` application appear. Click on it to see the visual tree of running Kubernetes resources:
   * Pod status
   * ReplicaSets
   * Service and Endpoints

2. **Access the App**:
   The Go web server service is exposed via a LoadBalancer on port `8085`. Access the health endpoint from your host machine at:
   👉 **[http://localhost:8085/health](http://localhost:8085/health)**
