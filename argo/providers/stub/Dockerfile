FROM golang:1.18 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY base/ base/
COPY stub/ stub/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider ./stub/cmd

FROM alpine:3.17.0
WORKDIR /
COPY --from=builder /workspace/ .
COPY stub/inject.sh /inject.sh
COPY stub/compile.sh /compile.sh
USER 65532:65532

ENTRYPOINT ["/bin/sh", "/inject.sh"]
