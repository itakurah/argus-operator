# Argus-Operator

Argus is a lightweight Kubernetes operator that automates rolling updates for your applications. It monitors referenced ConfigMaps and Secrets, triggering a native rollout whenever their data changes so your pods are always in sync with their configuration.

> [!NOTE]
> Currently supports **Deployments**, **StatefulSets**, and **DaemonSets**.

### Key Features

- **Smart Hashing:** Only restarts pods if the data actually changes (ignores metadata updates).
- **Zero-Downtime:** Leverages native Kubernetes Deployment controllers for safe rollouts.
- **Opt-in Only:** Only touches workloads you've explicitly annotated.
- **Lightweight:** Written in Go with minimal overhead.

## Enable a workload

Add this annotation on the workload’s `metadata.annotations`:

```yaml
argus.io/rollout-on-update: "true"
```

Workloads without that annotation are ignored.

## Install (Helm)

The namespace must exist, or create it with Helm:

```bash
helm  install  argus  ./charts/argus-operator  \
-n argus-system \
--create-namespace
```

Optional: `image.registry`, `image.pullSecrets` for private registries.
See `charts/argus-operator/values.yaml` for the full list of options.

**Uninstall**

```bash
helm  uninstall  argus  -n  argus-system
```

## Build the image

```bash
docker  build  -t  argus-operator:dev  .
```

## Development

```bash
go  test  ./...
go  build  -o  bin/manager  ./cmd/manager
```

Run locally against a cluster (e.g. `~/.kube/config`):

```bash
./bin/manager
```

## License

See [LICENSE](LICENSE).