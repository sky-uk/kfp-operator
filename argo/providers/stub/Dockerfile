FROM golang:1.18 as builder

WORKDIR /workspace

COPY argo argo
COPY go.mod go.mod
COPY go.sum go.sum
COPY apis apis

WORKDIR /workspace/argo/providers

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider ./stub/cmd

FROM alpine:3.17.0
WORKDIR /
COPY argo/providers/stub/inject.sh /inject.sh
COPY argo/providers/stub/compile.sh /compile.sh
COPY --from=builder /workspace/argo/providers .
USER 65532:65532

ENTRYPOINT ["/bin/sh", "/inject.sh"]
