package client

import (
	"fmt"
	"github.com/qmsk/snmpbot/snmp"
)

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
