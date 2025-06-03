FROM alpine:3.22.0

WORKDIR /

COPY inject.sh inject.sh
COPY compile.sh compile.sh

USER 65534:65534
ENTRYPOINT ["/inject.sh"]
