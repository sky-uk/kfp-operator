FROM alpine:3.17.0

WORKDIR /
COPY inject.sh inject.sh
COPY compile.sh compile.sh
COPY bin/provider .
USER 65532:65532

ENTRYPOINT ["/bin/sh", "/inject.sh"]
