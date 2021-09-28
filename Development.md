# Development

This project is still in early stages of development, so any help is appreciated. To build and run the controller yourself locally, the following instructions may help.

## Build and run

As a requirement, you will need a working kubernetes cluster. Tools like [Minikube](https://minikube.sigs.k8s.io/docs/start/) and [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) provide an instant out-of-the-box Kubernetes you can use for testing. Make sure to configure your local KUBECONFIG file to use the desired cluster.

From the project's root directory, use the following targets from the `Makefile`:

```shell
## build the project
make

## generate and install CRDs
make manifests install

## run the controller
make run
```

## Take it for a test drive 

Check out the contents of `/config/samples` for some examples of resources.

Watched Secrets must contain an actual TLS certificate. Not necessarily a valid one (it can be expired), but one that can be read and parsed from a [PEM format](https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail).

Once a Secret and its related CertWatcher is created, you can easily trigger a controller reaction by just changing any label in the Secret. 

```shell
## adding any label triggers CertWatcher action
kubectl label --overwrite secret example-tls l1=v1

## changing any label triggers CertWatcher action
kubectl label --overwrite secret example-tls l1=v2
```

The reaction to labels is an included behavior to facilitate testing and development. The same reaction can be expected when the TLS certificate content is changed as well.

---

_Powered by [kubebuilder](https://book.kubebuilder.io)._
