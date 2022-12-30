FROM golang:1.18 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

#RUN apt-get update && apt-get install -y ca-certificates openssl

#ARG cert_location=/usr/local/share/ca-certificates

# Get certificate from "github.com"
#RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
# Get certificate from "proxy.golang.org"
#RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
# Update certificates
#RUN update-ca-certificates

#RUN GOPROX=direct GOFLAGS="-insecure" go mod download
RUN go mod download

COPY base/ base/
COPY kfp/ kfp/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o provider ./kfp/cmd

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ .
USER 65532:65532

ENTRYPOINT ["/provider"]
