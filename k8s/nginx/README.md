# NGINX as Reverse Proxy for Mainflux services

### 1. Create TLS Secrets
- Create TLS server side certificate and keys
```bash 
cd certs
kubectl create secret tls mainflux-secret --key mainflux-server.key --cert mainflux-server.crt
```

### 2. Create Config Map
- Create Config Map based on the default.conf file. 
```bash 
cd ..
kubectl create configmap mainflux-nginx-config --from-file=default.conf
```

### 3. Create Nginx Deployment and Service
- Create Deployment and Service from mainflux-dashflux.yaml file.
```bash 
kubectl create -f mainflux-nginx.yaml
```