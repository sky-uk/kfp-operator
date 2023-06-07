FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY bin/status-updater .
USER 65532:65532

ENTRYPOINT ["/status-updater"]
