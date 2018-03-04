package client

import (
	"fmt"
	"github.com/qmsk/go-logging"
	"github.com/qmsk/snmpbot/snmp"
	"net"
)

func NewClient(engine *Engine, config Config) (*Client, error) {
	var client = makeClient(engine, config.Options)

	if addr, err := engine.transport.Resolve(config.Address); err != nil {
		return nil, fmt.Errorf("Resolve Config.Address=%v: %v", config.Address, err)
	} else {
		client.addr = addr
	}

	client.log = logging.WithPrefix(log, fmt.Sprintf("Client<%v>", &client))

	return &client, nil
}

func makeClient(engine *Engine, options Options) Client {
	return Client{
		engine:  engine,
		options: options,
	}
}

type Client struct {
	engine  *Engine
	options Options
	log     logging.PrefixLogging

	addr net.Addr // host or host:port
}

func (client *Client) String() string {
	return fmt.Sprintf("%v@%v", string(client.options.Community), client.addr)
}

func (client *Client) request(send IO) (IO, error) {
	var request = NewRequest(client.options, send)

	if err := client.engine.Request(request); err != nil {
		client.log.Infof("Request %v: %v", request, err)

		return IO{}, err

	} else if recv, err := request.Result(); err != nil {
		client.log.Infof("Request %v: %v", request, err)

		return recv, err
	} else {
		client.log.Infof("Request %v", request)

		return recv, nil
	}
}

func (client *Client) requestGeneric(requestType snmp.PDUType, varBinds []snmp.VarBind) (snmp.PDUType, []snmp.VarBind, error) {
	var maxVars = DefaultMaxVars
	var retType snmp.PDUType
	var retVars = make([]snmp.VarBind, len(varBinds))
	var retLen = uint(0)
	var send = IO{
		Addr: client.addr,
		Packet: snmp.Packet{
			Version:   SNMPVersion,
			Community: []byte(client.options.Community),
		},
		PDUType: requestType,
	}

	if client.options.MaxVars > 0 {
		maxVars = client.options.MaxVars
	}

	for retLen < uint(len(varBinds)) {
		var reqVars = make([]snmp.VarBind, maxVars)
		var reqLen = uint(0)

		for retLen+reqLen < uint(len(varBinds)) && reqLen < maxVars {
			reqVars[reqLen] = varBinds[retLen+reqLen]
			reqLen++
		}

		send.PDU = snmp.GenericPDU{
			VarBinds: reqVars[:reqLen],
		}

		// TODO: handle snmp.TooBigError
		if recv, err := client.request(send); err != nil {
			return recv.PDUType, nil, err
		} else if responsePDU, ok := recv.PDU.(snmp.GenericPDU); !ok {
			return recv.PDUType, nil, fmt.Errorf("Invalid %v with PDU of type %T", recv.PDUType, recv.PDU)
		} else if len(responsePDU.VarBinds) > len(reqVars) {
			return retType, retVars, fmt.Errorf("Invalid %v with %d vars for %v with %d vars", recv.PDUType, len(responsePDU.VarBinds), requestType, len(retVars))
		} else {
			retType = recv.PDUType

			for _, varBind := range responsePDU.VarBinds {
				retVars[retLen] = varBind
				retLen++
				reqLen++
			}
		}
	}

	return retType, retVars, nil
}

func (client *Client) Get(oids ...snmp.OID) ([]snmp.VarBind, error) {
	var requestVars = make([]snmp.VarBind, len(oids))

	for i, oid := range oids {
		requestVars[i] = snmp.MakeVarBind(oid, nil)
	}

	if len(oids) == 0 {
		return nil, nil
	} else if responseType, responseVars, err := client.requestGeneric(snmp.GetRequestType, requestVars); err != nil {
		return responseVars, err
	} else if responseType != snmp.GetResponseType {
		return responseVars, fmt.Errorf("Unexpected response type %v for GetRequest", responseType)
	} else if len(responseVars) != len(oids) {
		return nil, fmt.Errorf("Incorrect number of response vars %d for GetRequest with %d OIDs", len(responseVars), len(oids))
	} else {
		return responseVars, nil
	}
}

func (client *Client) GetNext(oids ...snmp.OID) ([]snmp.VarBind, error) {
	var requestVars = make([]snmp.VarBind, len(oids))

	for i, oid := range oids {
		requestVars[i] = snmp.MakeVarBind(oid, nil)
	}

	if len(oids) == 0 {
		return nil, nil
	} else if responseType, responseVars, err := client.requestGeneric(snmp.GetNextRequestType, requestVars); err != nil {
		return responseVars, err
	} else if responseType != snmp.GetResponseType {
		return responseVars, fmt.Errorf("Unexpected response type %v for GetNextRequest", responseType)
	} else if len(responseVars) != len(oids) {
		return nil, fmt.Errorf("Incorrect number of response vars %d for GetRequest with %d OIDs", len(responseVars), len(oids))
	} else {
		return responseVars, nil
	}
}
