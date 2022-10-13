FROM golang:1.18 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY base/ base/
COPY kfp/ kfp/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider kfp/main.go

FROM python:3.10.5-alpine3.16

ARG WHEEL_VERSION

RUN apk add jq

COPY dist/*-${WHEEL_VERSION}-*.whl /
RUN pip install /*-${WHEEL_VERSION}-*.whl && \
    rm /*-${WHEEL_VERSION}-*.whl

WORKDIR /
COPY --from=builder /workspace/ .
USER 65532:65532

ENTRYPOINT ["/provider"]
