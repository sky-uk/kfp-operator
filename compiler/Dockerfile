FROM alpine:3.14.2
ADD compiler/compiler.py /compiler-src/
ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/compiler"]
