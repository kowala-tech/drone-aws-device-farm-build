FROM alpine:3.6 as alpine

ENV GODEBUG=netdns=go

ADD release/linux/amd64/drone-device-farm /bin/

ENTRYPOINT ["/bin/drone-device-farm"]