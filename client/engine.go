package client

import (
	"fmt"
	"github.com/qmsk/go-logging"
)

func NewEngine(options Options) (*Engine, error) {
	if udp, err := NewUDP(options.UDP); err != nil {
		return nil, err
	} else {
		var engine = makeEngine(udp)

		engine.log = logging.WithPrefix(log, fmt.Sprintf("Engine<%v>", &engine))

		return &engine, nil
	}
}

func makeEngine(transport Transport) Engine {
	return Engine{
		transport: transport,

		requestID:   1, // TODO: randomize
		requests:    make(map[requestID]*request),
		requestChan: make(chan *request),
		timeoutChan: make(chan requestID),
	}
}

type Engine struct {
	log       logging.PrefixLogging
	transport Transport

	requestID   requestID
	requests    map[requestID]*request
	requestChan chan *request
	timeoutChan chan requestID
}

func (engine *Engine) String() string {
	return fmt.Sprintf("%v", engine.transport)
}

func (engine *Engine) nextRequestID() requestID {
	var requestID = engine.requestID

	engine.requestID++

	return requestID
}

func (engine *Engine) teardown() {
	engine.log.Debugf("teardown...")

	close(engine.requestChan)

	// cancel any queued requests
	for request := range engine.requestChan {
		request.close()
	}

	// cancel any active requests
	for _, request := range engine.requests {
		request.close()
	}
}

func (engine *Engine) startRequest(request *request) error {
	request.send.PDU.RequestID = int(request.id)

	engine.log.Debugf("Send: %#v", request.send)

	if err := engine.transport.Send(request.send); err != nil {
		err = fmt.Errorf("SNMP<%v> send failed: %v", engine.transport, err)

		request.fail(err)

		return err
	}

	request.start(request.timeout, engine.timeoutChan)

	return nil
}

func (engine *Engine) run() error {
	var recvChan = make(chan IO)
	var recvErr error

	defer engine.teardown()

	go func() {
		defer close(recvChan)

		for {
			if recv, err := engine.transport.Recv(); err != nil {
				engine.log.Errorf("Recv: %v", err)
				recvErr = err
				return
			} else {
				engine.log.Debugf("Recv: %#v", recv)

				recvChan <- recv
			}
		}
	}()

	for {
		select {
		case request := <-engine.requestChan:
			request.id = engine.nextRequestID()

			if err := engine.startRequest(request); err != nil {
				engine.log.Debugf("Start request %v failed: %v", request, err)
			} else {
				engine.log.Debugf("Start request: %v", request)

				engine.requests[request.id] = request
			}

		case recv, ok := <-recvChan:
			if !ok {
				return recvErr
			}

			requestID := requestID(recv.PDU.RequestID)

			if request, ok := engine.requests[requestID]; !ok {
				engine.log.Debugf("Recv for unknown requestID=%d", requestID)
			} else {
				engine.log.Debugf("Request done: %v", request)
				request.done(recv)
				delete(engine.requests, requestID)
			}

		case requestID := <-engine.timeoutChan:
			if request, ok := engine.requests[requestID]; !ok {
				engine.log.Debugf("Timeout with expired requestID=%d", requestID)

			} else if request.retry <= 0 {
				engine.log.Debugf("Timeout request: %v", request)

				request.failTimeout(engine.transport)

				delete(engine.requests, requestID)

			} else {
				request.retry--

				engine.log.Debugf("Retry on timeout (%d attempts remaining): %v", request.retry, request)

				if err := engine.startRequest(request); err != nil {
					engine.log.Debugf("Retry request %v failed: %v", request, err)

					// cleanup
					delete(engine.requests, requestID)
				}
			}
		}
	}
}

func (engine *Engine) Run() error {
	engine.log.Debugf("Run...")

	return engine.run()
}

func (engine *Engine) request(request *request) error {
	engine.requestChan <- request

	return request.wait()
}

// Closing the engine will cancel any requests, and cause Run() to return
func (engine *Engine) Close() error {
	engine.log.Debugf("Close...")

	return engine.transport.Close()
}
