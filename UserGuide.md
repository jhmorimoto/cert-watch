# User Guide

The starting point is to have a native [Kubernetes TLS Secret](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets) that you can use as a target. You can create those in any number of ways. If you need help, [cert-manager](https://cert-manager.io/docs/) is a good place to start. To improvise a certificate, [cert-manager can issue some from your own self-signed CA](https://cert-manager.io/docs/configuration/selfsigned/#bootstrapping-ca-issuers).

## Quick Start

Assuming you already have a TLS Secret, it would look similar to this:

```yaml
apiVersion: v1
stringData:
  tls.crt: ...
  tls.key: ...
kind: Secret
metadata:
  name: example-tls
  namespace: default
type: kubernetes.io/tls
```

Whatever solution you have to provision and renew certificates does not really matter. `cert-watch` is only interested in the contents of the Secret.

To watch for changes, create your own `CertWatcher`.

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: echo-example
spec:
  secret:
    name: example-tls
    namespace: default
  actions:
    echo: {}
```

The `echo` action type is the simplest example. It reacts by simply sending messages to the api Event Recorder.  After a `CertWatcher` is created, you can get details about its status with the usual `kubectl get`.

```shell
$ kubectl get certwatch echo

NAME   SECRET_NS   SECRET_NAME   STATUS   ACTION_STATUS   LAST_UPDATE            LAST_CHECKSUM                                  MESSAGE
echo   default     example-tls   Ready                    2021-09-28T03:44:02Z   b6XJqnEPw2HYaRjTcQcQ1r8fcf3J1gRslQu9dpY6x5g=   CertWatcher successfully initialized
```

A status `Ready` indicates that this CertWatcher was successfully initialized. The value of `LAST_CHECKSUM` will be populated with a hash of the current contents of the watched Secret.

Now, you can experiment by simply changing a label on the watched Secret:

```shell
kubectl label --overwrite secret example-tls somelabel=somevalue
```

Checking the status of the CertWatcher again:

```shell
NAME   SECRET_NS   SECRET_NAME   STATUS   ACTION_STATUS   LAST_UPDATE            LAST_CHECKSUM                                  MESSAGE
echo   default     example-tls   Ready    Ready           2021-09-28T03:55:12Z   6dWsBXVpAzz5Ms0LFLjw-uvGSZ5Bn6cKzB5W0wrHNm0=   Waiting for next Secret change
```

Take a closer look at more details using `kubectl describe`.

```shell
kubectl describe certwatcher echo
```

```yaml
Name:         echo
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  certwatch.morimoto.net.br/v1
Kind:         CertWatcher
Metadata:
  ...
Spec:
  Actions:
    Echo:
  Secret:
    Name:       example-tls
    Namespace:  default
Status:
  Action Status:  Ready
  Last Checksum:  6dWsBXVpAzz5Ms0LFLjw-uvGSZ5Bn6cKzB5W0wrHNm0=
  Last Update:    2021-09-28T03:55:12Z
  Message:        Waiting for next Secret change
  Status:         Ready
Events:
  Type    Reason                 Age   From                   Message
  ----    ------                 ----  ----                   -------
  Normal  CertWatcherInit        13m   CertWatcherReconciler  CertWatcher successfully initialized
  Normal  SecretChanged          2m    SecretReconciler       Updating CertWatcher status.
  Normal  CertWatcherProcessing  2m    CertWatcherReconciler  Processing pending actions
  Normal  CertWatcherProcessing  2m    CertWatcherReconciler  ECHO: Good morning to default/example-tls
  Normal  CertWatcherProcessing  2m    CertWatcherReconciler  Action processing finished successfully
```

The echo message will be in the Event list:

```
ECHO: Good morning to default/example-tls
```

You can also see events live, as actions are performed by using `kubectl get events` on a separate terminal window.

```shell
$ kubectl get events -w
0s          Normal    SecretChanged           certwatcher/echo                      Updating CertWatcher status.
0s          Normal    CertWatcherProcessing   certwatcher/echo                      Processing pending actions
0s          Normal    CertWatcherProcessing   certwatcher/echo                      ECHO: Good morning to default/example-tls
0s          Normal    CertWatcherProcessing   certwatcher/echo                      Action processing finished successfully
```

There are two ways to introduce a change in the Secret:

* Change the contents of the certificate data (`tls.key` or `tls.crt`), which will eventually happen if you wait long enough for your provisioner.
* Change the labels on the Secret metadata.

Either will cause the checksum to change and trigger a reaction in the related CertWatcher.


## Actions that a CertWatcher can perform

Depending on how your CertWatcher is configured, a few actions can be performed:

* [Sending an e-mail](UserGuide_Email.md)
* [Copying files over SCP](UserGuide_Scp.md)
* [Running a Kubernetes Job](UserGuide_Job.md)

A given CertWatcher can be configured to execute any combination of those, but each CertWatcher can only have one of each. That means, for example, that one CertWatcher can run all three actions. But, if you need to send multiple e-mails, you will need to create multiple CertWatcher objects.

## Certificate files ready to use

Before executing any actions, the Secret contents are copied as files into a temporary workspace directory. That directory is unique for each CertWatcher instance and is promptly removed after all actions are performed. But, while they are being performed, the following will be available to your CertWatcher:

| Filename        | Description                                                                                         |
|-----------------|-----------------------------------------------------------------------------------------------------|
| `tls.key`         | Certificate private key in PEM format                                                               |
| `tls.key.zip`     | Zip file containing `tls.key`                                                                         |
| `tls.crt`         | Public certificate in PEM format                                                                |
| `tls.crt.zip`     | Zip file containing `tls.crt`                                                                         |
| `tls.zip`         | Zip file containing `tls.key` and `tls.crt`                                                             |
| `tls.p12`         | PKCS#12 envelope containing `tls.key` and `tls.crt`                                                     |
| `tls.p12.zip`     | Zip file containing `tls.p12`                                                                         |
| `tls.crt.p12`     | PKCS#12 envelope containing `tls.crt`                                                                 |
| `tls.crt.p12.zip` | Zip file containing `tls.crt.p12`                                                                     |
| `tls.all.zip`     | Zip file containing all of the above: `tls.key`, `tls.crt`, `tls.p12` and `tls.crt.p12` |
                                                  

Aside from the original `tls.key` and `tls.crt` in conventional PEM format, a number of other variants are included, such as PKCS#12 and zipped versions. Zip files with both private/public keys, and isolated public certificates are included. The primary reason is to give you, `cert-watch` user, some of the most popular options to chose from. Files can be referenced in a CertWatcher by their filenames, always relative to the temporary directory.

## Additional CertWatcher options

Each CertWatcher can be configured in a few different ways. It is possible to change the filename prefix and protect files with a password.  This might be necessary for some recipient systems.

For example, an e-mail server running an anti-virus check might refuse an e-mail with an open raw certificate attached. Quite often, these attachments need to be zipped and password protected. While processing actions, `cert-watch` will provide these files ready-to-use.  The prefix `tls` is used in all filenames by default. That means, files will follow the convention `tls.key`, `tls.crt`, `tlz.zip`, etc.

You can use `filenamesPrefix` to change the prefix to any other string value.

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: echo
spec:
  secret:
    name: example-tls
    namespace: default
  filenamesPrefix: mycert
  actions:
    email:
      ...
```

The example above causes `cert-watch` to create temporary files named `mycert.key`, `mycert.crt`, `mycert.p12`, `mycert.zip`, etc.

To protect PKCS#12 envelopes with a password, use `pkcs12Password`.

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: echo
spec:
  secret:
    name: example-tls
    namespace: default
  pkcs12Password: changeit
  actions:
    email:
      ...
```

> _TIP: PKCS#12 can be considered a legacy format and may not provide sufficient security in its password encryption. As legacy systems might only accept this format, it is recommended to set PKCS#12 password to a standard value of `changeit` and use other means of protecting the certificate contents in transit, such as a password-protected zip file._

In a similar fashion, use `zipFilesPassword` to protect zip files with a password.

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: echo
spec:
  secret:
    name: example-tls
    namespace: default
  zipFilesPassword: my-password
  actions:
    email:
      ...
```
