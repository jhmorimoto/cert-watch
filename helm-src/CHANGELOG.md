# Helm Chart Changelog

## 0.2.3

Fix helm chart template. Remove reference to autoscaling.

## 0.2.2

Include missing permissions to "events" from Event Recorder in RBAC rules.

## 0.2.1

Add `helm.sh/resource-policy: keep` annotation to CRDs so helm can keep them in the cluster after a `helm delete/uninstall`.

## 0.2.0

Include `emailConfiguration` in values.yaml for email configuration. Files are injected as ConfigMap volumes into `/etc/cert-watch/emailconfig`.

## 0.1.2

First functional release.
