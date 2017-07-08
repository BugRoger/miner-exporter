FROM alpine:3.6
MAINTAINER Michael Schmidt <michael.j.schmidt@gmail.com>

ADD bin/linux/miner-exporter miner-exporter

ENTRYPOINT ["/miner-exporter"]
