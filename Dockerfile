FROM alpine:3.6 as alpine

ENV GODEBUG=netdns=go

ADD release/linux/amd64/drone-aws-device-farm-build /bin/

ENTRYPOINT ["/bin/drone-aws-device-farm-build"]