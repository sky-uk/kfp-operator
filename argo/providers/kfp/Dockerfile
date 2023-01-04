FROM golang:1.18 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY base/ base/
COPY kfp/ kfp/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider ./kfp/cmd

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ .
USER 65532:65532

ENTRYPOINT ["/provider"]
