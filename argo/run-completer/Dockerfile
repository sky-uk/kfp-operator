FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY bin/run-completer .
USER 65532:65532

ENTRYPOINT ["/run-completer"]
