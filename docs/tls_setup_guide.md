# Traefik Custom TLS Certificate Guide

This document explains how to set up, mount, and trust a custom TLS certificate for your local domains (`localhost`, `zot.localhost`, `headlamp.localhost`) on port `8443`.

By trusting this certificate on your Windows host, you will achieve:
1. **Secure Padlock (HTTPS)** in your web browsers without self-signed certificate warnings.
2. **Native Docker CLI Trust** without needing to add Zot to Docker Desktop's `insecure-registries` list.

---

## 1. How the Certificate is Managed & Generated

To keep local certificates clean and automated, the TLS configuration is organized under two directories:
* **[`certs/`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/certs/)**: Contains the active certificates (`tls.crt` and `tls.key`) mounted into Traefik.
* **[`certs-gen/`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/certs-gen/)**: Contains the automation configuration and scripts:
  * **[`tls-conf.cnf`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/certs-gen/tls-conf.cnf)**: The OpenSSL configuration file defining the SANs (Subject Alternative Names).
  * **[`generate.ps1`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/certs-gen/generate.ps1)**: A PowerShell script (for Windows) to generate/renew the certificates using Docker.
  * **[`generate.sh`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/certs-gen/generate.sh)**: A Bash script (for Linux/macOS) to generate/renew the certificates using Docker.

### Generating / Renewing Certificates
If you need to generate new certificates or update the domain list:
1. Modify the `[ alt_names ]` section inside `certs-gen/tls-conf.cnf`.
2. Run the generation script:
   ```powershell
   # On Windows (PowerShell):
   powershell.exe -ExecutionPolicy Bypass -File certs-gen/generate.ps1

   # On Linux/macOS (Bash):
   ./certs-gen/generate.sh
   ```

The script will automatically compile your domains and output the updated `tls.crt` and `tls.key` files into the `certs/` directory.

---

## 2. Dynamic Traefik Configuration

The certificate files are mounted inside the Traefik deployment container at `/etc/traefik/certs/` and configured under Traefik's dynamic file provider:

### Traefik Deployment Mount ([`k8s2/traefik.yaml`](file:///c:/Users/ADMIN/Desktop/notes/oryHydra/k8s2/traefik.yaml))
```yaml
spec:
  containers:
    - name: traefik
      volumeMounts:
        - name: cert-volume
          mountPath: /etc/traefik/certs
  volumes:
    - name: cert-volume
      secret:
        secretName: traefik-tls-cert
```

### Dynamic File Config (`routes.yml`)
```yaml
tls:
  certificates:
    - certFile: /etc/traefik/certs/tls.crt
      keyFile: /etc/traefik/certs/tls.key
```

This configuration ensures that Traefik automatically inspects the TLS client request (SNI) and serves the custom certificate when accessing any of the `*.localhost` domains.

---

## 3. How to Trust the Certificate on Windows

To make your Windows machine trust the self-signed certificate natively:

1. Open **PowerShell** as **Administrator**.
2. Run the following command to import the certificate into your Windows **Trusted Root Certification Authorities** store:
   ```powershell
   Import-Certificate -FilePath "c:\Users\ADMIN\Desktop\notes\oryHydra\certs\tls.crt" -CertStoreLocation "Cert:\LocalMachine\Root"
   ```
3. **Restart your browser** (Chrome/Edge) to load the new certificate authority.

---

## 4. (Optional) Cleanup Docker Desktop Settings

Once the certificate is trusted by Windows, Docker Desktop will trust Zot automatically. You can remove Zot from your insecure registries:

1. Open **Docker Desktop Settings** -> **Docker Engine**.
2. Remove `"zot.localhost:8443"` from the `insecure-registries` array.
3. Click **Apply & restart**.
