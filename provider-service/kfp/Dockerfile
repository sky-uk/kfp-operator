FROM gcr.io/distroless/static-debian12:nonroot

COPY bin/kfp-provider-service /bin/kfp-provider-service

COPY --from=base-source ./base/pkg/config/config.yaml /etc/provider-service/config.yaml

CMD ["/bin/kfp-provider-service"]
