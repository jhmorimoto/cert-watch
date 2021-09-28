# User Guide - React by copying files over SCP

A CertWatcher can react by copying files to a remote host via SSH/SCP. Below, a full example:

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: scp-example
spec:
  secret:
    name: example-tls
    namespace: default
  actions:
    scp:
      hostname: 10.0.0.2
      port: 22
      authType: "password"
      credentialSecret: default/my-secret-credentials
      files:
        - name: tls.key
          remotePath: /tmp
          mode: "0640"
        - name: tls.crt
          remotePath: /tmp
          mode: "0644"

```

While `hostname` and `port` are the usual suspects, additional information must be provided for authentication.

`authType` can be either:

* `password` for username/password based authentication; or
* `key` for ssh key based authentication.

The value for `credentialSecret` must be in the form `<NAMESPACE>/<SECRET_NAME>`, which is a reference to a standard Kubernetes Secret of type `Opaque`. 

```yaml
## Secret with username and password for `password` authentication.
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: scp-credentials
  namespace: default
stringData:
  password: mysecret
  username: myuser
```

```yaml
## Secret with username and SSH key for `key` authentication.
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: scp-credentials-keys
  namespace: default
stringData:
  username: user
  key: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
  passphrase: mypassphrase
```

A passphrase for the SSH key is optional and can be provided if your key needs is protected with one.

The `files` list includes all certificate files that will be copied to the remote host.

```yaml
      files:
        - name: tls.key
          remotePath: /tmp
          mode: "0640"
```

In the file list, each entry refers to one file that must be copied:

| Configuration | Description                                                                                                          |
|---------------|----------------------------------------------------------------------------------------------------------------------|
| `name`        | Local file name, referring to one of the files included in the temporary workspace directory.                         |
| `remotePath`  | Directory in the remote host where the file will be copied to.                                                       |
| `mode`        | File mode the remote copy will have. Must be in the [numeric unix format](https://en.wikipedia.org/wiki/File-system_permissions#Numeric_notation), ex: `0644`. If omitted, defaults to `0600`. |
