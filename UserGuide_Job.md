# User Guide - React by running a Kubernetes Job

A CertWatcher can react by running a standard [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/). Below, a full example:

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: job-example
spec:
  secret:
    name: example-tls
    namespace: default
  actions:
    job:
      name: myjob
      spec:
        template:
          spec:
            containers:
              - name: app
                image: ubuntu:20.04
                command: ["cat"]
                tty: true
            restartPolicy: Never
```
Inside the `job` key, the value for `spec` follows the native Kubernetes [`batch/v1` sepificiation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#jobspec-v1-batch), where you can declare any Pod to run when the action is triggered.

This should give you a high level of flexibility to distribute certificates. Any container image can be executed with custom commands and specilized scripts or programs. Being a standard Kubernetes Job, the Pod itself is even allowed to have multiple containers, each one with a custom processing logic of their own.

## Configuring volume name and mount path

The original TLS Secret containing the certificada will be injected into all containers of the Pod as an additional Volume named `certs` and `mountPath: /workspace`. You can use the values of `volumeName` and `mountPath` to change these defaults:

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: job-example
spec:
  secret:
    name: example-tls
    namespace: default
  actions:
    job:
      name: myjob
      volumeName: my-cert-volume files
      mountPath: /path/to/cert/files 
      spec:
        template:
          ...
```

## Limitations

Contrary to `email` and `scp` actions, where the temporary workspace directory is readily available for the controller process, **the running Pod will not have all the same files available in various formats**. Because `cert-watch` mounts the original TLS Secret as a Volume, only `tls.key` and `tls.crt` will be available.

Jobs instances are created using its given name, suffixed by a random hash to avoid conflicts between multiple and subsquent executions. There is currently no provision to clean up job instances after they complete. So, right now, be aware that these jobs will add up in the etcd database and must be manually removed.
