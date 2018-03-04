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

func (client *Client) requestGeneric(requestType snmp.PDUType, varBinds []snmp.VarBind, responseType snmp.PDUType) ([]snmp.VarBind, error) {
	var send = IO{
		Addr: client.addr,
		Packet: snmp.Packet{
			Version:   SNMPVersion,
			Community: []byte(client.options.Community),
		},
		PDUType: requestType,
		PDU: snmp.GenericPDU{
			VarBinds: varBinds,
		},
	}

	if len(varBinds) == 0 {
		return nil, nil
	} else if recv, err := client.request(send); err != nil {
		return nil, err
	} else if recv.PDUType != responseType {
		return nil, fmt.Errorf("Invalid %v response type, expected %v, got %v", requestType, responseType, recv.PDUType)
	} else if responsePDU, ok := recv.PDU.(snmp.GenericPDU); !ok {
		return nil, fmt.Errorf("Invalid %v response type, expected %v, got %v with PDU of type %T", requestType, responseType, recv.PDUType, recv.PDU)
	} else if len(responsePDU.VarBinds) != len(varBinds) {
		return responsePDU.VarBinds, fmt.Errorf("Invalid %v response, expected %d vars, got %v with %d vars", requestType, len(varBinds), recv.PDUType, len(responsePDU.VarBinds))
	} else {
		return responsePDU.VarBinds, nil
	}
}

func makeGetVars(oids []snmp.OID) []snmp.VarBind {
	var varBinds = make([]snmp.VarBind, len(oids))

	for i, oid := range oids {
		varBinds[i] = snmp.MakeVarBind(oid, nil)
	}

	return varBinds
}

func (client *Client) Get(oids ...snmp.OID) ([]snmp.VarBind, error) {
	return client.requestGeneric(snmp.GetRequestType, makeGetVars(oids), snmp.GetResponseType)
}

func (client *Client) GetNext(oids ...snmp.OID) ([]snmp.VarBind, error) {
	return client.requestGeneric(snmp.GetNextRequestType, makeGetVars(oids), snmp.GetResponseType)
}
