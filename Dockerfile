FROM alpine:latest AS ca
RUN apk add -U ca-certificates

FROM scratch
COPY --from=ca /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY bin/travisci-exporter-linux /bin/travisci-exporter
EXPOSE 9099
ENTRYPOINT ["/bin/travisci-exporter"]
CMD [""]
