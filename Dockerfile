# go backend
FROM golang:1.10.4-stretch as go-build

RUN curl -L -o /tmp/dep-linux-amd64 https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && install -m 0755 /tmp/dep-linux-amd64 /usr/local/bin/dep

WORKDIR /go/src/github.com/qmsk/snmpbot

COPY Gopkg.* ./
RUN dep ensure -vendor-only

COPY . ./
RUN go install -v ./cmd/...

# runtime
# must match with go-build base image
FROM debian:stretch

RUN adduser --system --home /opt/qmsk/snmpbot --uid 1000 --gid 100 qmsk-snmpbot

RUN mkdir -p \
  /opt/qmsk/snmpbot \
  /opt/qmsk/snmpbot/bin

COPY --from=go-build /go/bin/snmp* /opt/qmsk/snmpbot/bin/

USER qmsk-snmpbot
ENV PATH=$PATH:/opt/qmsk/snmpbot/bin
CMD ["/opt/qmsk/snmpbot/bin/snmpbot", \
  "--http-listen=:8286" \
]
