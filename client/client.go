package client

import (
	"fmt"
	snmp "github.com/qmsk/snmpbot/snmp_new"
	"net"
)

const (
	SNMPVersion = snmp.SNMPv2c
)

func (config Config) Client() (*Client, error) {
	var client = makeClient(config.Logging)

	if err := client.connectUDP(config); err != nil {
		return nil, err
	}

	return &client, nil
}

func makeClient(logging Logging) Client {
	return Client{
		log: logging,

		requestID:   1, // TODO: randomize
		requests:    make(map[requestID]*request),
		requestChan: make(chan *request),
	}
}

type Client struct {
	log Logging

	version   snmp.Version
	community []byte

	addr      net.Addr
	transport Transport

	requestID   requestID
	requests    map[requestID]*request
	requestChan chan *request
}

func (client *Client) String() string {
	if client.transport == nil {
		return "<disconnected>"
	}

	return fmt.Sprintf("%s@%v", client.community, client.transport)
}

func (client *Client) connectUDP(config Config) error {
	client.version = SNMPVersion
	client.community = []byte(config.Community)

	if udp, err := DialUDP(config.Addr, UDPOptions{}); err != nil {
		return fmt.Errorf("DialUDP %v: %v", config.Addr, err)
	} else {
		client.log.Infof("Connect UDP %v: %v", config.Addr, udp)

		client.addr = udp.addr
		client.transport = udp
	}

	return nil
}

func (client *Client) nextRequestID() requestID {
	var requestID = client.requestID

	client.requestID++

	return requestID
}

func (client *Client) teardown() {
	client.log.Debugf("teardown...")

	close(client.requestChan)
	for _, request := range client.requests {
		request.cancel()
	}
}

func (client *Client) run() error {
	var recvChan = make(chan IO)
	var recvErr error

	defer client.teardown()

	go func() {
		defer close(recvChan)

		for {
			if recv, err := client.transport.Recv(); err != nil {
				client.log.Errorf("recv: %v", err) // TODO: store
				recvErr = err
				return
			} else if string(recv.Packet.Community) != string(client.community) {
				client.log.Warnf("wrong community=%s", string(recv.Packet.Community))
			} else {
				client.log.Debugf("recv: %#v", recv)

				recvChan <- recv
			}
		}
	}()

	for {
		select {
		case request := <-client.requestChan:
			requestID := client.nextRequestID()

			request.send.PDU.RequestID = int(requestID)

			client.log.Debugf("send: %#v", request.send)

			if err := client.transport.Send(request.send); err != nil {
				client.log.Debugf("request %d fail: %v", requestID, err)
				request.fail(err)
			} else {
				client.log.Debugf("request %d...", requestID)

				client.requests[requestID] = request
			}

		case recv, ok := <-recvChan:
			if !ok {
				return recvErr
			}

			requestID := requestID(recv.PDU.RequestID)

			if request, ok := client.requests[requestID]; !ok {
				client.log.Warnf("recv with unknown requestID=%d", requestID)
			} else {
				client.log.Debugf("request %d done", requestID)
				request.done(recv)
				delete(client.requests, requestID)
			}
		}
	}
}

func (client *Client) Run() error {
	client.log.Debugf("Run...")

	return client.run()
}

func (client *Client) Close() error {
	client.log.Debugf("Close...")

	return client.transport.Close()
}
