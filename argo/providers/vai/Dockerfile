FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY bin/provider .
USER 65532:65532

ENTRYPOINT ["/provider"]
