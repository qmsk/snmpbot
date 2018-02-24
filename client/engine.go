package client

import (
	"fmt"
)

func NewEngine(options Options) (*Engine, error) {
	if udp, err := NewUDP(options.UDP); err != nil {
		return nil, err
	} else {
		var engine = makeEngine(udp)

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
	log.Debugf("%v teardown...", engine)

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

	if err := engine.transport.Send(request.send); err != nil {
		err = fmt.Errorf("SNMP %v send failed: %v", engine.transport, err)

		log.Debugf("%v request %d fail: %v", engine, request.id, err)

		request.fail(err)

		return err
	} else {
		log.Debugf("%v request %d...", engine, request.id)
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
				log.Errorf("%v recv: %v", engine, err)
				recvErr = err
				return
			} else {
				log.Debugf("%v recv: %#v", engine, recv)

				recvChan <- recv
			}
		}
	}()

	for {
		select {
		case request := <-engine.requestChan:
			request.id = engine.nextRequestID()

			log.Debugf("%v request %d send: %#v", engine, request.id, request.send)

			if err := engine.startRequest(request); err == nil {
				engine.requests[request.id] = request
			}

		case recv, ok := <-recvChan:
			if !ok {
				return recvErr
			}

			requestID := requestID(recv.PDU.RequestID)

			if request, ok := engine.requests[requestID]; !ok {
				log.Warnf("%v recv with unknown requestID=%d", engine, requestID)
			} else {
				log.Debugf("%v request %d done", engine, requestID)
				request.done(recv)
				delete(engine.requests, requestID)
			}

		case requestID := <-engine.timeoutChan:
			if request, ok := engine.requests[requestID]; !ok {
				log.Debugf("%v timeout with expired requestID=%d", engine, requestID)
			} else if request.retry <= 0 {
				log.Debugf("%v request %d timeout", engine, request.id)

				request.failTimeout(engine.transport)
				delete(engine.requests, requestID)
			} else {
				log.Debugf("%v request %d retry %d...", engine, request.id, request.retry)

				request.retry--

				if err := engine.startRequest(request); err != nil {
					// cleanup
					delete(engine.requests, requestID)
				}
			}
		}
	}
}

func (engine *Engine) Run() error {
	log.Debugf("%v Run...", engine)

	return engine.run()
}

func (engine *Engine) request(request *request) error {
	engine.requestChan <- request

	return request.wait()
}

// Closing the engine will cancel any requests, and cause Run() to return
func (engine *Engine) Close() error {
	log.Debugf("%v Close...", engine)

	return engine.transport.Close()
}
