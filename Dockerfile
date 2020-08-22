## build go backend
FROM golang:1.15-buster as go-build

WORKDIR /go/src/github.com/qmsk/snmpbot

# dependencies
COPY go.mod go.sum ./
RUN go mod download

# source code
COPY . ./
RUN go install -v ./cmd/...


## download mibs
FROM buildpack-deps:stretch-scm as get-mibs

ARG SNMPBOT_MIBS_VERSION=0.1.0

RUN curl -fsSL https://github.com/qmsk/snmpbot-mibs/archive/v${SNMPBOT_MIBS_VERSION}.tar.gz | tar -C /tmp -xzv


## runtime
# must match with go-build base image
FROM debian:stretch

RUN adduser --system --home /opt/qmsk/snmpbot --uid 1000 --gid 100 qmsk-snmpbot

RUN mkdir -p \
  /opt/qmsk/snmpbot \
  /opt/qmsk/snmpbot/bin \
  /opt/qmsk/snmpbot/mibs

COPY --from=go-build /go/bin/snmp* /opt/qmsk/snmpbot/bin/
COPY --from=get-mibs /tmp/snmpbot-mibs-* /opt/qmsk/snmpbot/mibs/

USER qmsk-snmpbot
ENV \
  PATH=$PATH:/opt/qmsk/snmpbot/bin \
  SNMPBOT_MIBS=/opt/qmsk/snmpbot/mibs

CMD ["/opt/qmsk/snmpbot/bin/snmpbot", \
  "--http-listen=:8286", \
  "--verbose" \
]
EXPOSE 8286/tcp
