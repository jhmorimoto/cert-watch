# cert-watch
Watch and react to change in Kubernetes TLS Secrets.

## What is cert-watch?

Kubernetes has introduced a number of different ways to keep certificates generated, renewed and updated. Tools like [cert-manager](https://cert-manager.io/docs/) provide an easy way to issue and renew TLS certificates _inside the cluster_ The only drawback is exactly that last bit: **_inside the cluster_**.

While certificates are easily managed inside your Kubernetes cluster, the tools that issue them do not provide a straight forward way to distribute certificates to the outside world. As we enter a new age of cloud computing, we still live in a mixed era where, sometimes, shiny new Kubernetes clusters need to play ball and integrate with older legacy infrastructure.

`cert-watch` provides a way to "expose" certificates managed inside the Kubernetes. While conected to the apiserver, it watches for native Secrets resources (type `kubernetes.io/tls`). Whenever they change (ie: a cert is renewed) it reacts to perform actions that can distribute them into other environments.

Actions can vary from sending an e-mail with certificates attached, copying them into a remote host via SSH/SCP or running a shell script to perform a custom set of operations.

## State of the matter

This project is still in early stages of development and no release has been made just yet. Currently, the following actions are implemented:

* Sending an e-mail via standard SMTP
* Copying to remote hosts via SSH/SCP

However, there running this controller still involves working out the details of a development IDE for local testing.

## Build and run

To build the project and run the controller, you will need a working kubernetes cluster. Tools like [Minikube](https://minikube.sigs.k8s.io/docs/start/) provide an instant out-of-the-box Kubernetes you can use for testing. Make sure you configure your local KUBECONFIG file to use the desired cluster and use the following commands from the `Makefile`:

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

---

_Powered by [kubebuilder](https://book.kubebuilder.io)._
