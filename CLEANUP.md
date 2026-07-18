# Cleanup Guide

This guide describes how to delete the deployed web servers and clean up the ArgoCD installation from your local Kubernetes cluster.

---

## 🗑️ Option 1: Remove Only the Applications (Keep ArgoCD)

If you want to remove the two Go servers (`go-server` and `second-server`) but keep ArgoCD and Argo CD Image Updater installed in your cluster, use one of the methods below:

### Method A: Delete via ArgoCD Application Manifest (Recommended)
Deleting the ArgoCD Application resource will automatically cascade-delete all the Deployments, Services, and Pods created for both web servers:
```powershell
kubectl delete -f k8s/argocd-app.yaml
```

### Method B: Delete via GitOps (Manifest Pruning)
If you want to keep the ArgoCD application registration but remove the deployments:
1. Delete the manifest files from the `k8s/` directory in your Git branch:
   * `k8s/deployment.yaml`
   * `k8s/service.yaml`
   * `k8s/second-deployment.yaml`
   * `k8s/second-service.yaml`
2. Commit and push the deletions:
   ```powershell
   git rm k8s/deployment.yaml k8s/service.yaml k8s/second-deployment.yaml k8s/second-service.yaml
   git commit -m "Remove web server manifests"
   git push origin argocd-deployment
   ```
3. ArgoCD will detect the file deletions in Git and automatically **prune** (delete) the running Deployments and Services in your cluster.

---

## 🧹 Option 2: Complete Uninstall (ArgoCD & Services)

If you want to completely uninstall the web servers, Argo CD Image Updater, and ArgoCD from your cluster, run the following commands:

1. **Delete the Web Server App**:
   ```powershell
   kubectl delete app go-server-app -n argocd
   ```

2. **Uninstall Argo CD Image Updater** (Helm release):
   ```powershell
   helm uninstall argocd-image-updater --namespace argocd
   ```

3. **Uninstall ArgoCD Core Components**:
   ```powershell
   kubectl delete -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

4. **Delete the Namespace**:
   ```powershell
   kubectl delete namespace argocd
   ```

These steps will completely clean your Kubernetes cluster back to its original state.
