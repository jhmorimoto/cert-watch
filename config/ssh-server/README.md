# SSH/SCP Testing

The `Makefile` in this directory provides a _quick-and-easy_ OpenSSH server that can be used for testing.  The only requirement for this setup is to have recent versions of `docker-compose` and Docker installed.

A couple of disposable keys are provided in the files:

* `ssh.key` and `ssh.key.pub`
* `ssh-nopassphrase.key` and `ssh-nopassphrase.key.pub`

Please, consider those keys insecure for anything else other than testing this particular component. In case they need to be recreated from scratch, run:

```
make keys
```

> NOTE: If you create new keys, remember that they are copy-referenced into the sample manifests under `/config/samples/`, so make sure to update those manually as well.

* `/config/samples/scp-credentials-keys.yaml`
* `/config/samples/scp-credentials-keys-nopass.yaml`

## Start/Stop the OpenSSH Server

The targets in the `Makefile` are quite obvious:

```
## start the server
make start

## watch the logs
make logs

## obtain a shell inside the container
make shell

## stop the server and cleanup
make stop
```

Upon starting up, the keys will be loaded into `~/.ssh/authorized_keys` inside the container.

If you make sure private and public keys match with the samples you are testing with, everything should work and the controller will copy the certificate files into the SSH container and the Secret changes.
