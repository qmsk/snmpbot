package snmp

import (
    "fmt"
    "net"
    "time"
    wapsnmp "github.com/cdevr/WapSNMP"
)

const (
    UDP_SIZE   = 16 * 1024
)

type GenericTrap int

const (
    TrapColdStart   GenericTrap = 0
    TrapWarmStart   GenericTrap = 1
    TrapLinkDown    GenericTrap = 2
    TrapLinkUp      GenericTrap = 3
    TrapAuthenticationFailure   GenericTrap = 4
    TrapEgpNeighborLoss         GenericTrap = 5
    TrapEnterpriseSpecific      GenericTrap = 6
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

// SNMPv1 Trap-PDU
type TrapPDU struct {
    Enterprise      wapsnmp.Oid
    AgentAddr       net.IP
    GenericTrap     GenericTrap
    SpecificTrap    int
    TimeStamp       time.Duration
    VarBinds        []VarBind
}

type VarBind struct {
    Name        wapsnmp.Oid
    Value       interface{}
}

func packSeq(items... interface{}) []interface{} {
    seq := []interface{}{wapsnmp.Sequence}
    seq = append(seq, items...)

    return seq
}

func decodePacket(buf []byte) (packet Packet, pduSeq []interface{}, err error) {
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

func encodePacket(packet Packet, pduSeq []interface{}) ([]byte, error) {
    seq := packSeq(
        int(packet.Version),
        packet.Community,
        pduSeq,
    )

    return wapsnmp.EncodeSequence(seq)
}

func unpackPDU(seq []interface{}) (pdu PDU, err error) {
    if len(seq) != 5 {
        err = fmt.Errorf("invalid 4-sequence")
    } else if seqType, ok := seq[0].(wapsnmp.BERType); !ok {
        err = fmt.Errorf("invalid PDU type: %#v", seq[0])
    } else {
        _ = seqType
    }

    if seqRequestID, ok := seq[1].(int64); !ok {
        err = fmt.Errorf("invalid request-id: %#v", seq[1])
        return
    } else {
        pdu.RequestID = int32(seqRequestID)
    }

    if seqErrorStatus, ok := seq[2].(int64); !ok {
        return pdu, fmt.Errorf("invalid error-status: %#v", seq[2])
    } else {
        pdu.ErrorStatus = int(seqErrorStatus)
    }

    if seqErrorIndex, ok := seq[3].(int64); !ok {
        return pdu, fmt.Errorf("invalid error-index: %#v", seq[3])
    } else {
        pdu.ErrorIndex = int(seqErrorIndex)
    }

    if seqVarBinds, ok := seq[4].([]interface{}); !ok {
        return pdu, fmt.Errorf("invalid variable-bindings: %#v", seq[4])
    } else if pduVarBinds, err := unpackVarBinds(seqVarBinds); err != nil {
        return pdu, fmt.Errorf("invalid variable-bindings: %s", err)
    } else {
        pdu.VarBinds = pduVarBinds
    }

    return
}

func packPDU(pduType wapsnmp.BERType, pdu PDU) []interface{} {
    return []interface{}{pduType,
        int(pdu.RequestID),
        int(pdu.ErrorStatus),
        int(pdu.ErrorIndex),
        packVarBinds(pdu.VarBinds),
    }
}

func unpackVarBinds(seq []interface{}) (vars []VarBind, err error) {
    if len(seq) < 1 || seq[0] != wapsnmp.Sequence {
        err = fmt.Errorf("invalid sequence")
        return
    }

    for _, seqItem := range seq[1:] {
        if varSeq, ok := seqItem.([]interface{}); !ok {
            return vars, fmt.Errorf("invalid varbind sequence: %#v", seqItem)
        } else if varBind, err := unpackVarBind(varSeq); err != nil {
            return vars, err
        } else {
            vars = append(vars, varBind)
        }
    }

    return
}

func packVarBinds(vars []VarBind) (seq []interface{}) {
    seq = packSeq()

    for _, varBind := range vars {
        seq = append(seq, packVarBind(varBind))
    }

    return seq
}

func unpackVarBind(seq []interface{}) (varBind VarBind, err error) {
    if len(seq) != 3 || seq[0] != wapsnmp.Sequence {
        err = fmt.Errorf("invalid 2-sequence")
        return
    }

    if seqName, ok := seq[1].(wapsnmp.Oid); !ok {
        err = fmt.Errorf("invalid name: %#v", seq[1])
        return
    } else {
        varBind.Name = seqName
    }

    varBind.Value = seq[2]

    return
}

func packVarBind(varBind VarBind) (seq []interface{}) {
    return packSeq(varBind.Name, varBind.Value)
}

func unpackTrapPDU(seq []interface{}) (pdu TrapPDU, err error) {
    if len(seq) != 7 {
        return pdu, fmt.Errorf("invalid 6-sequence")
    } else if seqType, ok := seq[0].(wapsnmp.BERType); !ok {
        return pdu, fmt.Errorf("invalid PDU type: %#v", seq[0])
    } else {
        _ = seqType
    }

    if seqEnterpriseOid, ok := seq[1].(wapsnmp.Oid); !ok {
        return pdu, fmt.Errorf("invalid enterprise oid: %#v", seq[1])
    } else {
        pdu.Enterprise = seqEnterpriseOid
    }

    if seqAgentAddr, ok := seq[2].(net.IP); !ok {
        return pdu, fmt.Errorf("invalid agent-address: %#v", seq[2])
    } else {
        pdu.AgentAddr = seqAgentAddr
    }

    if seqGenericTrap, ok := seq[3].(int64); !ok {
        return pdu, fmt.Errorf("invalid generic-trap: %#v", seq[3])
    } else {
        pdu.GenericTrap = GenericTrap(seqGenericTrap)
    }

    if seqSpecificTrap, ok := seq[4].(int64); !ok {
        return pdu, fmt.Errorf("invalid specific-trap: %#v", seq[4])
    } else {
        pdu.SpecificTrap = int(seqSpecificTrap)
    }

    if seqTimeStamp, ok := seq[5].(time.Duration); !ok {
        return pdu, fmt.Errorf("invalid time-stamp: %#v", seq[5])
    } else {
        pdu.TimeStamp = seqTimeStamp
    }

    if seqVarBinds, ok := seq[6].([]interface{}); !ok {
        return pdu, fmt.Errorf("invalid variable-bindings: %#v", seq[6])
    } else if pduVarBinds, err := unpackVarBinds(seqVarBinds); err != nil {
        return pdu, fmt.Errorf("invalid variable-bindings: %s", err)
    } else {
        pdu.VarBinds = pduVarBinds
    }

    return pdu, nil
}
