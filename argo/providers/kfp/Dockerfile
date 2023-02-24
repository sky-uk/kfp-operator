FROM golang:1.18 as builder

WORKDIR /workspace

COPY argo argo
COPY go.mod go.mod
COPY go.sum go.sum
COPY apis apis

WORKDIR /workspace/argo/providers

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider ./kfp/cmd
RUN ls
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/argo/providers .
USER 65532:65532

ENTRYPOINT ["/provider"]
