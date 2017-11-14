package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
	"time"
)

const (
	SNMPVersion    = snmp.SNMPv2c
	DefaultTimeout = 1 * time.Second
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
		timeoutChan: make(chan requestID),
	}
}

type Client struct {
	log Logging

	version   snmp.Version
	community []byte
	timeout   time.Duration
	retry     int

	transport Transport

	requestID   requestID
	requests    map[requestID]*request
	requestChan chan *request
	timeoutChan chan requestID
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

	if config.Timeout == 0 {
		client.timeout = DefaultTimeout
	} else {
		client.timeout = config.Timeout
	}
	client.retry = config.Retry

	if udp, err := DialUDP(config.Addr, UDPOptions{}); err != nil {
		return fmt.Errorf("DialUDP %v: %v", config.Addr, err)
	} else {
		client.log.Infof("Connect UDP %v: %v", config.Addr, udp)

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

	// cancel any queued requests
	for request := range client.requestChan {
		request.close()
	}

	// cancel any active requests
	for _, request := range client.requests {
		request.close()
	}
}

func (client *Client) startRequest(request *request) error {
	request.send.PDU.RequestID = int(request.id)

	if err := client.transport.Send(request.send); err != nil {
		client.log.Debugf("request %d fail: %v", request.id, err)

		request.fail(err)

		return err
	} else {
		client.log.Debugf("request %d...", request.id)
	}

	request.start(request.timeout, client.timeoutChan)

	return nil
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
			request.id = client.nextRequestID()

			client.log.Debugf("request %d send: %#v", request.id, request.send)

			if err := client.startRequest(request); err == nil {
				client.requests[request.id] = request
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

		case requestID := <-client.timeoutChan:
			if request, ok := client.requests[requestID]; !ok {
				client.log.Debugf("timeout with expired requestID=%d", requestID)
			} else if request.retry <= 0 {
				client.log.Debugf("request %d timeout", request.id)

				request.failTimeout(client.transport)
				delete(client.requests, requestID)
			} else {
				client.log.Debugf("request %d retry %d...", request.id, request.retry)

				request.retry--

				if err := client.startRequest(request); err != nil {
					// cleanup
					delete(client.requests, requestID)
				}
			}
		}
	}
}

func (client *Client) Run() error {
	client.log.Debugf("Run...")

	return client.run()
}

// Closing the client will cancel any requests, and cause Run() to return
func (client *Client) Close() error {
	client.log.Debugf("Close...")

	return client.transport.Close()
}
