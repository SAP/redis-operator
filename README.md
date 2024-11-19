# Kubernetes Operator For Redis™

[![REUSE status](https://api.reuse.software/badge/github.com/SAP/redis-operator)](https://api.reuse.software/info/github.com/SAP/redis-operator)

Disclaimer: Redis is a registered trademark of Redis Ltd. Any rights therein are reserved to Redis Ltd. Any use by SAP is for referential purposes only and does not indicate any sponsorship, endorsement, or affiliation between Redis Ltd. and SAP.

## About this project
This repository adds a new resource type `Redis` (`redis.cache.cs.sap.com`) to Kubernetes clusters,
which can be used to deploy redis caches for cluster-internal usage. For example:

```yaml
apiVersion: cache.cs.sap.com/v1alpha1
kind: Redis
metadata:
  name: test
spec:
  replicas: 3
  sentinel:
    enabled: true
  metrics:
    enabled: true
  tls:
    enabled: true
```

The controller contained in this repository under the hood uses the [bitnami redis chart](https://github.com/bitnami/charts/tree/main/bitnami/redis)
to install redis in the cluster. As a consequence of this fact, the following topologies are supported:
- statically configured master with optional read replicas
- sentinel cluster (i.e. dynamic master with read replicas, master elected by sentinel).

Sharding (redis-cluster) scenarios are not supported.

### Sentinel mode
If `spec.sentinel.enabled` is false, one redis master node will be deployed, and `spec.replicas - 1` read replicas.
Both master and read nodes are reachable at dedicated services; since the master statefulset currently cannot be scaled beyound 1, 
only the read part is truly highly available. 

If `spec.sentinel.enabled` is true, then an ensemble of `spec.replicas` nodes will be deployed, each of which runs the actual redis service, and a sentinel sidecar. As long as a quorum of sentinels is available (more than 50%), they will form a consensus about which of the redis services has the master role, and configure the redis instances accordingly. There will be one service, exposing the sentinels at port `26379`, and the redis caches at port `6379`; clients which just want to perform read operations, can directly connect to the service at `6379`; in order to write to redis, clients have to connect to the sentinel port of the service first, in order to detect the address of the current master, and then connect to the retrieved address at `6379`.

Note that the field `spec.sentinel.enabled` is immutable.

### Encryption
TLS encryption can be turned on by setting `spec.tls.enabled`. Without further configuration, a self-signed certificate will be created.
As an alternative, if available, certificate and key can be retrieved from [cert-manager](https://cert-manager.io). With

```yaml
spec:
  tls:
    enabled: true
    certManager: {}
```

a self-signing issuer will be generated; an existing issuer could be referenced as well, such as:

```yaml
spec:
  tls:
    enabled: true
    certManager:
      issuer:
        # group: cert-manager.io
        kind: ClusterIssuer
        name: cluster-ca
```

### Persistence

AOF persistence can be enabled by setting `spec.persistence.enabled` to true. It may be tweaked by setting
`spec.persistence.storageClass` and `spec.persistence.size`; note that the latter fields are immutable.

### Metrics

If `spec.metrics.enabled` is set to true, an prometheus exporter sidecar will be added to the pods, which can be scraped
at port `9121` (optionally via the corresponding service and `ServiceMonitor`, if [prometheus-operator](https://prometheus-operator.dev) is used).

### Binding secret

By default, a binding secret like the following will be generated:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-test-binding
type: Opaque
stringData:
  caData: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  host: redis-test.testns.svc.cluster.local
  masterName: mymaster
  password: BM5vR1ziGE
  port: "6379"
  sentinelEnabled: "true"
  sentinelHost: redis-test.testns.svc.cluster.local
  sentinelPort: "26379"
  tlsEnabled: "true"
```

The format of the secret data can be overridden by specifying a go temmplate as `spec.binding.template`.
In that go template, the following variables may be used:
- `.sentinelEnabled` (whether sentinel mode is enabled or not)
- `.masterHost`, `.masterPort`, `.replicaHost`, `.replicaPort` (only if sentinel is disabled)
- `.host`, `.port`, `.sentinelHost`, `.sentinelPort`, `.masterName` (only if sentinel is enabled)
- `.tlsEnabled` (whether TLS encryption is enabled or not)
- `.caData` (CA certificate that clients may use to connect to redis)

### Customize pod settings

The following attributes allow to tweak the created pods/containers:
- `spec.nodeSelector`
- `spec.affinity`
- `spec.topologySpreadConstraints`
- `spec.tolerations`
- `spec.priorityClassName`
- `spec.podSecurityContext`
- `spec.podLabels`
- `spec.podAnnotations`
- `spec.resources`
- `spec.securityContext`
- `spec.sentinel.resources`
- `spec.sentinel.securityContext`
- `spec.metrics.resources`
- `spec.metrics.securityContext`

For topology spread constraints, a special logic applies: if undefined, then
some weak spread constraints will be generated, such as
```yaml
topologySpreadConstraints:
- labelSelector:
    matchLabels:
      app.kubernetes.io/component: node
      app.kubernetes.io/instance: test
      app.kubernetes.io/name: redis
  maxSkew: 1
  nodeAffinityPolicy: Honor
  nodeTaintsPolicy: Honor
  topologyKey: kubernetes.io/hostname
  whenUnsatisfiable: ScheduleAnyway
  matchLabelKeys:
  - controller-revision-hash
```
This does not harm but helps to ensure proper spreading of the redis pods across Kubernetes nodes.
In addition, if a supplied constraint misses both `labelSelector` and `matchLabelKeys`, then
these attributes will be automatically populated by the controller, as in the above example.

## Requirements and Setup

The recommended deployment method is to use the [Helm chart](https://github.com/sap/redis-operator-helm):

```bash
helm upgrade -i redis-operator oci://ghcr.io/sap/redis-operator-helm/redis-operator
```

## Documentation

The API reference is here: [https://pkg.go.dev/github.com/sap/redis-operator](https://pkg.go.dev/github.com/sap/redis-operator).

## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/SAP/redis-operator/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](CONTRIBUTING.md).

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/SAP/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2024 SAP SE or an SAP affiliate company and redis-operator contributors. Please see our [LICENSE](LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/SAP/redis-operator).
