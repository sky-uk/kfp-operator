FROM gcr.io/distroless/static-debian12:nonroot

COPY bin/run-completion-event-trigger /bin/run-completion-event-trigger

COPY ./config/config.yaml /etc/run-completion-event-trigger/config.yaml

CMD ["/bin/run-completion-event-trigger"]
