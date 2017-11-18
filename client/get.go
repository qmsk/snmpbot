package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

func (client *Client) requestRead(requestType snmp.PDUType, varBinds []snmp.VarBind) (snmp.PDUType, []snmp.VarBind, error) {
	var retType snmp.PDUType
	var retVars = make([]snmp.VarBind, len(varBinds))
	var retLen = 0
	var send = IO{
		Packet: snmp.Packet{
			Version:   client.version,
			Community: client.community,
		},
		PDUType: requestType,
		PDU:     snmp.PDU{},
	}

	for retLen < len(varBinds) {
		var reqVars = make([]snmp.VarBind, client.maxVars)
		var reqLen = 0

		for retLen+reqLen < len(varBinds) && reqLen < client.maxVars {
			reqVars[reqLen] = varBinds[retLen+reqLen]
			reqLen++
		}

		send.PDU.VarBinds = reqVars[:reqLen]

		// TODO: handle snmp.TooBigError
		if recv, err := client.request(send); err != nil {
			return recv.PDUType, recv.PDU.VarBinds, err
		} else if len(recv.PDU.VarBinds) > len(reqVars) {
			return retType, retVars, fmt.Errorf("Invalid %v with %d vars for %v with %d vars", recv.PDUType, len(recv.PDU.VarBinds), requestType, len(retVars))
		} else {
			retType = recv.PDUType

			for _, varBind := range recv.PDU.VarBinds {
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

	if responseType, responseVars, err := client.requestRead(snmp.GetRequestType, requestVars); err != nil {
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

	if responseType, responseVars, err := client.requestRead(snmp.GetNextRequestType, requestVars); err != nil {
		return responseVars, err
	} else if responseType != snmp.GetResponseType {
		return responseVars, fmt.Errorf("Unexpected response type %v for GetNextRequest", responseType)
	} else if len(responseVars) != len(oids) {
		return nil, fmt.Errorf("Incorrect number of response vars %d for GetRequest with %d OIDs", len(responseVars), len(oids))
	} else {
		return responseVars, nil
	}
}
