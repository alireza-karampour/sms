# Deployment Guide

## Overview

The SMS Gateway is designed for deployment in Kubernetes environments using Helm charts. The system consists of multiple components that can be deployed independently or together.

## Prerequisites

### System Requirements

- **Kubernetes**: Version 1.20+
- **Helm**: Version 3.0+
- **PostgreSQL**: Version 13+
- **NATS**: Version 2.8+

### Resource Requirements

**Minimum Requirements**:
- CPU: 2 cores
- Memory: 4GB RAM
- Storage: 20GB

**Recommended Requirements**:
- CPU: 4 cores
- Memory: 8GB RAM
- Storage: 100GB

## Component Overview

The deployment consists of several components:

1. **NATS JetStream**: Message queue system
2. **PostgreSQL**: Database system
3. **SMS API**: REST API server
4. **SMS Worker**: Background message processor

## Helm Charts

### NATS Chart

Located in `charts/nats/`, this chart deploys a NATS JetStream cluster.

#### Chart Structure

```
charts/nats/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── stateful-set.yaml
│   ├── service.yaml
│   ├── config-map.yaml
│   └── ...
└── files/
    ├── config/
    │   ├── cluster.yaml
    │   ├── jetstream.yaml
    │   └── ...
    └── ...
```

#### Key Features

- **StatefulSet**: Persistent NATS cluster
- **JetStream**: Enabled for message persistence
- **Clustering**: Multi-node NATS cluster
- **Monitoring**: Prometheus metrics support
- **TLS**: Transport layer security (configurable)

#### Configuration

```yaml
# charts/nats/values.yaml
jetstream:
  enabled: true
  storage: "file"
  maxMemory: "1Gi"
  maxFile: "10Gi"

cluster:
  enabled: true
  replicas: 3

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
```

### PostgreSQL Chart

Located in `charts/postgres/`, this chart deploys a PostgreSQL database.

#### Chart Structure

```
charts/postgres/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ...
└── ...
```

#### Key Features

- **Deployment**: Single PostgreSQL instance
- **Persistent Storage**: PVC for data persistence
- **Configuration**: Customizable PostgreSQL settings
- **Backup**: Automated backup support (future)

#### Configuration

```yaml
# charts/postgres/values.yaml
postgresql:
  database: "sms_db"
  username: "sms_user"
  password: "secure_password"

persistence:
  enabled: true
  size: "20Gi"

resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "1000m"
```

## Deployment Steps

### 1. Prepare Kubernetes Cluster

```bash
# Verify cluster access
kubectl cluster-info

# Check available nodes
kubectl get nodes

# Verify Helm installation
helm version
```

### 2. Deploy NATS

```bash
# Add NATS Helm repository (if needed)
helm repo add nats https://nats-io.github.io/k8s/helm/charts/

# Deploy NATS with JetStream
helm install nats ./charts/nats \
  --namespace sms-system \
  --create-namespace \
  --values charts/nats/values.yaml

# Verify deployment
kubectl get pods -n sms-system
kubectl get services -n sms-system
```

### 3. Deploy PostgreSQL

```bash
# Deploy PostgreSQL
helm install postgres ./charts/postgres \
  --namespace sms-system \
  --values charts/postgres/values.yaml

# Verify deployment
kubectl get pods -n sms-system
kubectl get services -n sms-system
```

### 4. Initialize Database

```bash
# Get PostgreSQL pod name
POSTGRES_POD=$(kubectl get pods -n sms-system -l app=postgres -o jsonpath='{.items[0].metadata.name}')

# Execute schema creation
kubectl exec -it $POSTGRES_POD -n sms-system -- psql -U sms_user -d sms_db -f /path/to/schema.sql
```

### 5. Deploy SMS Gateway

```bash
# Build Docker image
docker build -t sms-gateway:latest .

# Deploy API server
kubectl create deployment sms-api \
  --image=sms-gateway:latest \
  --namespace sms-system \
  --command -- /app/sms api

# Deploy worker
kubectl create deployment sms-worker \
  --image=sms-gateway:latest \
  --namespace sms-system \
  --command -- /app/sms worker

# Create services
kubectl expose deployment sms-api \
  --port=8080 \
  --target-port=8080 \
  --namespace sms-system
```

## Configuration Management

### ConfigMaps

Create configuration ConfigMaps:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sms-config
  namespace: sms-system
data:
  SmsGW.yaml: |
    api:
      nats:
        address: "nats:4222"
      listen: "0.0.0.0:8080"
      postgres:
        address: "postgres"
        port: 5432
        username: "sms_user"
        password: "secure_password"
    
    worker:
      nats:
        address: "nats:4222"
      postgres:
        address: "postgres"
        port: 5432
        username: "sms_user"
        password: "secure_password"
    
    sms:
      cost: "5.0"
      normal:
        ratelimit: 200
      express:
        ratelimit: 100
```

### Secrets

Create secrets for sensitive data:

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: sms-secrets
  namespace: sms-system
type: Opaque
data:
  postgres-password: <base64-encoded-password>
  nats-password: <base64-encoded-password>
```

## Environment-Specific Deployments

### Development Environment

```bash
# Deploy with development values
helm install sms-dev ./charts/sms \
  --namespace sms-dev \
  --create-namespace \
  --values charts/sms/values-dev.yaml
```

### Staging Environment

```bash
# Deploy with staging values
helm install sms-staging ./charts/sms \
  --namespace sms-staging \
  --create-namespace \
  --values charts/sms/values-staging.yaml
```

### Production Environment

```bash
# Deploy with production values
helm install sms-prod ./charts/sms \
  --namespace sms-prod \
  --create-namespace \
  --values charts/sms/values-prod.yaml
```

## Scaling

### Horizontal Scaling

#### API Server Scaling

```bash
# Scale API servers
kubectl scale deployment sms-api --replicas=3 -n sms-system

# Verify scaling
kubectl get pods -n sms-system -l app=sms-api
```

#### Worker Scaling

```bash
# Scale workers
kubectl scale deployment sms-worker --replicas=5 -n sms-system

# Verify scaling
kubectl get pods -n sms-system -l app=sms-worker
```

#### Database Scaling

```bash
# Scale PostgreSQL (if using read replicas)
kubectl scale deployment postgres-read --replicas=2 -n sms-system
```

### Vertical Scaling

#### Resource Limits

```yaml
# resources.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sms-api
spec:
  template:
    spec:
      containers:
      - name: sms-api
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
```

## Monitoring and Observability

### Prometheus Integration

```yaml
# service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: sms-api
  namespace: sms-system
spec:
  selector:
    matchLabels:
      app: sms-api
  endpoints:
  - port: metrics
    path: /metrics
```

### Grafana Dashboards

Create dashboards for:
- SMS throughput
- Queue depth
- Database performance
- API response times

### Logging

```yaml
# logging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sms-logging
data:
  log-level: "info"
  log-format: "json"
```

## Backup and Recovery

### Database Backup

```bash
# Create backup job
kubectl create job postgres-backup \
  --image=postgres:13 \
  --namespace sms-system \
  --command -- pg_dump -h postgres -U sms_user sms_db > /backup/sms_backup.sql
```

### Configuration Backup

```bash
# Backup ConfigMaps
kubectl get configmap sms-config -n sms-system -o yaml > sms-config-backup.yaml

# Backup Secrets
kubectl get secret sms-secrets -n sms-system -o yaml > sms-secrets-backup.yaml
```

### Disaster Recovery

```bash
# Restore from backup
kubectl apply -f sms-config-backup.yaml
kubectl apply -f sms-secrets-backup.yaml

# Restore database
kubectl exec -it postgres-pod -- psql -U sms_user -d sms_db < sms_backup.sql
```

## Security

### Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: sms-network-policy
  namespace: sms-system
spec:
  podSelector:
    matchLabels:
      app: sms-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
```

### Pod Security Policies

```yaml
# pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: sms-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
```

## Troubleshooting

### Common Issues

1. **Pod Startup Failures**
   ```bash
   # Check pod logs
   kubectl logs -f deployment/sms-api -n sms-system
   
   # Check pod status
   kubectl describe pod <pod-name> -n sms-system
   ```

2. **Database Connection Issues**
   ```bash
   # Test database connectivity
   kubectl exec -it <pod-name> -- nc -zv postgres 5432
   
   # Check database logs
   kubectl logs -f deployment/postgres -n sms-system
   ```

3. **NATS Connection Issues**
   ```bash
   # Test NATS connectivity
   kubectl exec -it <pod-name> -- nc -zv nats 4222
   
   # Check NATS logs
   kubectl logs -f deployment/nats -n sms-system
   ```

### Debug Commands

```bash
# Check all resources
kubectl get all -n sms-system

# Check events
kubectl get events -n sms-system --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n sms-system
kubectl top nodes
```

## Performance Tuning

### Resource Optimization

```yaml
# performance.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sms-api
spec:
  template:
    spec:
      containers:
      - name: sms-api
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        env:
        - name: GOMAXPROCS
          value: "2"
```

### Database Tuning

```yaml
# postgres-tuning.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
data:
  postgresql.conf: |
    shared_buffers = 256MB
    effective_cache_size = 1GB
    maintenance_work_mem = 64MB
    checkpoint_completion_target = 0.9
    wal_buffers = 16MB
    default_statistics_target = 100
```

## Future Enhancements

### Planned Features

- **Auto-scaling**: Horizontal Pod Autoscaler (HPA)
- **Service Mesh**: Istio integration
- **GitOps**: ArgoCD deployment
- **Multi-cluster**: Cross-cluster deployment
- **Disaster Recovery**: Automated failover
- **Cost Optimization**: Resource optimization recommendations