package snmp

import (
    "fmt"
    wapsnmp "github.com/cdevr/WapSNMP"
)

const (
    UDP_SIZE   = 16 * 1024
)

type Packet struct {
    Version     wapsnmp.SNMPVersion
    Community   string
    PduType     wapsnmp.BERType
}

type PDU struct {
    RequestID   int32
    ErrorStatus int
    ErrorIndex  int

    VarBinds    []VarBind
}

type VarBind struct {
    Name        wapsnmp.Oid
    Value       interface{}
}

func parsePacket(buf []byte) (packet Packet, pduSeq []interface{}, err error) {
    // decode
    seq, err := wapsnmp.DecodeSequence(buf)
    if err != nil {
        return
    }

    // parse
    if len(seq) != 4 || seq[0] != wapsnmp.Sequence {
        err = fmt.Errorf("invalid 3-sequence")
        return
    }

    if seqVersion, ok := seq[1].(int64); !ok {
        err = fmt.Errorf("invalid version: %#v", seq[1])
        return
    } else {
        packet.Version = wapsnmp.SNMPVersion(seqVersion)
    }

    if seqCommunity, ok := seq[2].(string); !ok {
        err = fmt.Errorf("invalid community: %#v", seq[2])
        return
    } else {
        packet.Community = seqCommunity
    }

    if seqPdu, ok := seq[3].([]interface{}); !ok {
        err = fmt.Errorf("invalid PDU: %#v", seq[3])
        return
    } else if seqPduType, ok := seqPdu[0].(wapsnmp.BERType); !ok {
        return packet, pduSeq, fmt.Errorf("invalid PDU Type: %#v", seqPdu[0])
    } else {
        packet.PduType = seqPduType

        return packet, seqPdu, nil
    }
}

func parsePDU(seq []interface{}) (pdu PDU, err error) {
    if len(seq) != 5 {
        err = fmt.Errorf("invalid 4-sequence")
    } else if seqType, ok := seq[0].(wapsnmp.BERType); !ok {
        err = fmt.Errorf("invalid PDU type: %#v", seq[0])
    } else {
        _ = seqType
    }

    if seqRequestID, ok := seq[1].(int64); !ok {
        err = fmt.Errorf("invalid request-id")
        return
    } else {
        pdu.RequestID = int32(seqRequestID)
    }

    if seqErrorStatus, ok := seq[2].(int64); !ok {
        return pdu, fmt.Errorf("invalid error-status")
    } else {
        pdu.ErrorStatus = int(seqErrorStatus)
    }

    if seqErrorIndex, ok := seq[3].(int64); !ok {
        return pdu, fmt.Errorf("invalid error-index")
    } else {
        pdu.ErrorIndex = int(seqErrorIndex)
    }

    if seqVarBinds, ok := seq[4].([]interface{}); !ok {
        return pdu, fmt.Errorf("invalid variable-bindings")
    } else if pduVarBinds, err := parseVarBinds(seqVarBinds); err != nil {
        return pdu, fmt.Errorf("invalid variable-bindings: %s", err)
    } else {
        pdu.VarBinds = pduVarBinds
    }

    return
}

func parseVarBinds(seq []interface{}) (vars []VarBind, err error) {
    if len(seq) < 1 || seq[0] != wapsnmp.Sequence {
        err = fmt.Errorf("invalid sequence")
        return
    }

    for _, seqItem := range seq[1:] {
        if varSeq, ok := seqItem.([]interface{}); !ok {
            return vars, fmt.Errorf("invalid varbind sequence")
        } else if varBind, err := parseVarBind(varSeq); err != nil {
            return vars, err
        } else {
            vars = append(vars, varBind)
        }
    }

    return
}

func parseVarBind(seq []interface{}) (varBind VarBind, err error) {
    if len(seq) != 3 || seq[0] != wapsnmp.Sequence {
        err = fmt.Errorf("invalid 2-sequence")
        return
    }

    if seqName, ok := seq[1].(wapsnmp.Oid); !ok {
        err = fmt.Errorf("invalid name")
        return
    } else {
        varBind.Name = seqName
    }

    varBind.Value = seq[2]

    return
}
