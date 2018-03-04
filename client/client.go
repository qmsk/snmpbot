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

func (client *Client) requestPDU(requestType snmp.PDUType, pdu snmp.PDU, responseType snmp.PDUType) ([]snmp.VarBind, error) {
	var send = IO{
		Addr: client.addr,
		Packet: snmp.Packet{
			Version:   SNMPVersion,
			Community: []byte(client.options.Community),
		},
		PDUType: requestType,
		PDU:     pdu,
	}

	if recv, err := client.request(send); err != nil {
		return nil, err
	} else if recv.PDUType != responseType {
		return nil, fmt.Errorf("Invalid %v response type, expected %v, got %v", requestType, responseType, recv.PDUType)
	} else if responsePDU, ok := recv.PDU.(snmp.GenericPDU); !ok {
		return nil, fmt.Errorf("Invalid %v response type, expected %v, got %v with PDU of type %T", requestType, responseType, recv.PDUType, recv.PDU)
	} else {
		return responsePDU.VarBinds, nil
	}
}

func (client *Client) requestGeneric(requestType snmp.PDUType, varBinds []snmp.VarBind, responseType snmp.PDUType) ([]snmp.VarBind, error) {
	var pdu = snmp.GenericPDU{
		VarBinds: varBinds,
	}

	if len(varBinds) == 0 {
		return nil, nil
	} else if varBinds, err := client.requestPDU(requestType, pdu, responseType); err != nil {
		return nil, err
	} else if len(varBinds) != len(varBinds) {
		return varBinds, fmt.Errorf("Invalid %v response, expected %d vars, got %v with %d vars", requestType, len(varBinds), responseType, len(varBinds))
	} else {
		return varBinds, nil
	}
}

// Split request OIDs into multiple requests of options.MaxVars each.
//
// Override response varbinds outside of rootOIDs with snmp.EndOfMibViewValue
//
// TODO: automatically handle snmp.TooBigError?
func (client *Client) requestSplit(requestType snmp.PDUType, varBinds []snmp.VarBind, responseType snmp.PDUType) ([]snmp.VarBind, error) {
	var maxVars = DefaultMaxVars
	var retVars = make([]snmp.VarBind, len(varBinds))
	var retLen = uint(0)

	if client.options.MaxVars > 0 {
		maxVars = client.options.MaxVars
	}

	for retLen < uint(len(varBinds)) {
		var reqOffset = retLen
		var reqVars = make([]snmp.VarBind, maxVars)
		var reqLen = uint(0)

		for retLen+reqLen < uint(len(varBinds)) && reqLen < maxVars {
			reqVars[reqLen] = varBinds[reqOffset+reqLen]
			reqLen++
		}

		if varBinds, err := client.requestGeneric(requestType, reqVars[:reqLen], responseType); err != nil {
			return nil, err
		} else {
			for _, varBind := range varBinds {
				retVars[retLen] = varBind
				retLen++
			}
		}
	}

	return retVars, nil
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

func (client *Client) GetNextSplit(oids []snmp.OID) ([]snmp.VarBind, error) {
	return client.requestSplit(snmp.GetNextRequestType, makeGetVars(oids), snmp.GetResponseType)
}

func makeBulkVars(scalars []snmp.OID, entries []snmp.OID) []snmp.VarBind {
	var varBinds = make([]snmp.VarBind, len(scalars)+len(entries))

	for i, oid := range scalars {
		varBinds[i] = snmp.MakeVarBind(oid, nil)
	}
	for i, oid := range entries {
		varBinds[len(scalars)+i] = snmp.MakeVarBind(oid, nil)
	}

	return varBinds
}

func unpackBulkVars(scalarCount int, entryLen int, varBinds []snmp.VarBind) ([]snmp.VarBind, [][]snmp.VarBind, error) {
	var scalarVars = varBinds[:scalarCount]
	var entryCount = (len(varBinds) - scalarCount) / entryLen
	var entryList = make([][]snmp.VarBind, entryCount)

	if len(varBinds) < scalarCount+entryLen {
		return nil, nil, fmt.Errorf("Invalid bulk response for %d+%d => %d vars", scalarCount, entryLen, len(varBinds))
	}

	for i := 0; i < entryCount; i++ {
		var enrtryVars = make([]snmp.VarBind, entryLen)

		for j := 0; j < entryLen; j++ {
			enrtryVars[j] = varBinds[scalarCount+i*entryLen+j]
		}

		entryList[i] = enrtryVars
	}

	return scalarVars, entryList, nil

}

func (client *Client) GetBulk(scalars []snmp.OID, entries []snmp.OID) ([]snmp.VarBind, [][]snmp.VarBind, error) {
	var scalarLen = uint(len(scalars))
	var entriesLen = uint(len(entries))
	var maxRepetitions = DefaultMaxRepetitions

	if client.options.MaxRepetitions != 0 {
		maxRepetitions = client.options.MaxRepetitions
	}

	if scalarLen >= client.options.MaxVars || entriesLen >= client.options.MaxVars-scalarLen {
		maxRepetitions = 1
	} else if scalarLen+maxRepetitions*entriesLen > client.options.MaxVars {
		maxRepetitions = (client.options.MaxVars - scalarLen) / entriesLen
	}

	var pdu = snmp.BulkPDU{
		NonRepeaters:   len(scalars),
		MaxRepetitions: int(maxRepetitions),
		VarBinds:       makeBulkVars(scalars, entries),
	}

	if len(pdu.VarBinds) == 0 {
		return nil, nil, nil
	}

	if varBinds, err := client.requestPDU(snmp.GetBulkRequestType, pdu, snmp.GetResponseType); err != nil {
		return nil, nil, err
	} else {
		return unpackBulkVars(len(scalars), len(entries), varBinds)
	}
}
