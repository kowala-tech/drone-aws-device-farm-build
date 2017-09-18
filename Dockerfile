FROM byrnedo/alpine-curl as alpine
RUN apk add -U --no-cache ca-certificates

ENV GODEBUG=netdns=go

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD release/linux/amd64/drone-aws-device-farm-build /bin/

ENTRYPOINT ["/bin/drone-aws-device-farm-build"]