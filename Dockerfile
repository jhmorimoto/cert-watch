# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y openssl zip
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
