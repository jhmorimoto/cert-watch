# User Guide - React by sending an e-mail

A CertWatcher can react by sending an e-mail. Below, a full example:

```yaml
apiVersion: certwatch.morimoto.net.br/v1
kind: CertWatcher
metadata:
  name: email-example
spec:
  secret:
    name: example-tls
    namespace: default
  actions:
    email:
      configFile: /path/to/email.properties
      to: john.doe@example.com
      from: "CertWatch <no-reply@example.com>"
      subject: "Certificate has changed"
      bodyContentType: text/html
      bodyTemplate: |-
        <h1>The certificate has been renewed.</h1>.
      attachments:
        - tls.zip
        - tls.p12.zip
```

Aside from the usual values used in any e-mail setup (from, to, subject, body, etc.), a few other things are worth noting here. A path to a configuration file must be provided, which contains details about the SMTP server and authentication to use. The format is quite simplistic:

```
host: smtp.example.com
port: 25
username: someuser
password: sompassword
encryption: SSL / TLS / SSLTLS / STARTTLS
from: NoReply <me@host.com>
```

If `encryption` is omitted, no encryption will be used during authentication. Likewise, if username and password are omitted, no authentication will be used.

The value of `from` is an overall requirement for e-mail communication. In this file, it can be considered a default and will be overridden if redefined in your CertWatcher spec.

## Sources for `configFile`

At the moment, the contents of this file are not yet included in the CertWatcher CRD specification, but they can be easily injected in the `cert-watch` controller Pod as a volume.  Like any Kubernetes volume, its source source can be a `ConfigMap` or a `Secret`. You only need to match the volume `mountPath` to the `configFile` path in your CertWatchers. There are no limits as to how many configuration files can be mounted in your controller instance.

The controller process itself can also receive the command line argument `--emailconfig=/path/to/email.properties`. If present, it will work as a default to all CertWatchers, overridden by `configFile` in each instance.
