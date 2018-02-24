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
		requests:    make(requestMap),
		requestChan: make(chan *request),
		timeoutChan: make(chan ioKey),
		recvChan:    make(chan IO),
	}
}

type Engine struct {
	log       logging.PrefixLogging
	transport Transport

	requestID   requestID
	requests    requestMap
	requestChan chan *request
	timeoutChan chan ioKey

	recvChan chan IO
	recvErr  error
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

func (engine *Engine) receiver() {
	defer close(engine.recvChan)

	for {
		if recv, err := engine.transport.Recv(); err != nil {
			engine.log.Errorf("Recv: %v", err)

			engine.recvErr = err
			return
		} else {
			engine.log.Debugf("Recv: %#v", recv)

			engine.recvChan <- recv
		}
	}
}

func (engine *Engine) sendRequest(request *request) error {
	engine.log.Debugf("Send: %#v", request.send)

	if err := engine.transport.Send(request.send); err != nil {
		return fmt.Errorf("SNMP<%v> send failed: %v", engine.transport, err)
	}

	return nil
}

func (engine *Engine) startRequest(request *request) {
	// initialize request with next request ID to get the request key used to track send/recv/timeout
	requestKey := request.init(engine.nextRequestID())

	if err := engine.sendRequest(request); err != nil {
		engine.log.Debugf("Start request %v failed: %v", requestKey, err)

		request.fail(err)
	} else {
		engine.log.Debugf("Start request %v: %v", requestKey, request)

		engine.requests[requestKey] = request

		request.startTimeout(engine.timeoutChan, requestKey)
	}
}

func (engine *Engine) recvRequest(recv IO) {
	requestKey := recv.key()

	if request, ok := engine.requests[requestKey]; !ok {
		engine.log.Debugf("Unknown request %v recv", requestKey)
	} else {
		engine.log.Debugf("Request %v done: %v", requestKey, request)

		request.done(recv)

		delete(engine.requests, requestKey)
	}
}

func (engine *Engine) timeoutRequest(requestKey ioKey) {
	if request, ok := engine.requests[requestKey]; !ok {
		engine.log.Debugf("Unknown request %v timeout", requestKey)

	} else if request.retry <= 0 {
		engine.log.Debugf("Timeout %v request: %v", requestKey, request)

		request.failTimeout(engine.transport)

		delete(engine.requests, requestKey)

	} else {
		request.retry--

		engine.log.Debugf("Retry request %v on timeout (%d attempts remaining): %v", requestKey, request.retry, request)

		if err := engine.sendRequest(request); err != nil {
			engine.log.Debugf("Retry request %v failed: %v", requestKey, err)

			// cleanup
			delete(engine.requests, requestKey)

			request.fail(err)
		} else {
			request.startTimeout(engine.timeoutChan, requestKey)
		}
	}
}

func (engine *Engine) run() error {
	defer engine.teardown()

	for {
		select {
		case request := <-engine.requestChan:
			engine.startRequest(request)

		case recvIO, ok := <-engine.recvChan:
			if !ok {
				return engine.recvErr
			}

			engine.recvRequest(recvIO)

		case requestKey := <-engine.timeoutChan:
			engine.timeoutRequest(requestKey)
		}
	}
}

func (engine *Engine) Run() error {
	engine.log.Debugf("Run...")

	go engine.receiver()

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
