# `cert-watch`
Watch and react to change in Kubernetes TLS Secrets.

## What is `cert-watch`?

Kubernetes has introduced a number of different ways to keep certificates generated, renewed and updated. Tools like [cert-manager](https://cert-manager.io/docs/) provide an easy way to issue and renew TLS certificates _inside the cluster_ The only drawback is exactly that last bit: **_inside the cluster_**.

While certificates are easily managed inside your Kubernetes cluster, the tools that issue them do not provide a straight forward way to distribute certificates to the outside world. As we enter a new age of cloud computing, we still live in a mixed era where, sometimes, shiny new Kubernetes clusters need to play ball and integrate with older legacy infrastructure.

`cert-watch` provides a way to distribute certificates provisioned and renewed inside a Kubernetes cluster. While conected to the apiserver, it watches for native changes in Secrets resources (type `kubernetes.io/tls`). Whenever TLS Secrets change (ie: a cert is renewed) it reacts to perform actions that can distribute them into other environments.

Actions can vary from sending an e-mail with certificates attached, copying them into a remote host via SSH/SCP or running a [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) to perform a custom set of operations.

## Roadmap

- [x] React with dummy echo
- [x] React sending an e-mail
- [x] React copying files over SCP
- [x] Publish Docker image
- [x] Publish helm chart
- [ ] Add more features and fix all the bugs...
- [ ] Release a stable 1.0.0

## Installation via Helm Chart

A helm chart is provided for quick installation. To use it, include the helm repository to your local helm installation.

```shell
helm repo add cert-watch https://raw.githubusercontent.com/jhmorimoto/cert-watch/main/helm
helm repo up
```

Install a release with:

```shell
helm upgrade --create-namespace --install -n cert-watch cert-watch cert-watch/cert-watch
```

Be aware there are no stable releases yet. In practice, that means the helm chart is prepared to use the `latest` Docker image and will promptly pull new versions of this tag on every restart (`imagePullPolicy: Always`). Not only that, the compiled binary has debugging flags and log levels enabled for active development. This will continue to be the case for mostly 0.x.x releases, until a stable 1.x.x is reached.

### CRDs managed by Helm

The helm chart also manages the CRDs. At this point, it lacks the annotation to prevent removal of CRDs on `helm delete`. By uninstalling a release, all CRDs and existing CertWatchers in the cluster will be removed as well.

## User Guide

For details on how to use `cert-watch`, check out the [User Guide](UserGuide.md).

## Development

If you wish to contribute or would like to run the controller yourself locally, checkout the [Development](Development.md) quick start guide.

---

_Powered by [kubebuilder](https://book.kubebuilder.io)._
