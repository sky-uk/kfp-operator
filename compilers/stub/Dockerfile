FROM alpine:3.20.1

WORKDIR /

COPY inject.sh inject.sh
COPY compile.sh compile.sh

USER 65534:65534
ENTRYPOINT ["/inject.sh"]
