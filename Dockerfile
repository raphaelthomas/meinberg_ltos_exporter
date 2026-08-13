FROM alpine:3.23 AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
ARG TARGETPLATFORM

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY $TARGETPLATFORM/meinberg_ltos_exporter /bin/meinberg_ltos_exporter

USER 65534:65534

EXPOSE 10123
ENTRYPOINT ["/bin/meinberg_ltos_exporter"]
