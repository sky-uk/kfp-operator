FROM gcr.io/distroless/static-debian12:nonroot

COPY bin/vai-provider-service /bin/vai-provider-service

COPY --from=base-source ./base/pkg/config/config.yaml /etc/provider-service/config.yaml

CMD ["/bin/vai-provider-service"]
