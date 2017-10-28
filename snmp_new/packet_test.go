package snmp

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func decodeTestPacket(str string) []byte {
	str = regexp.MustCompile(`\s+|--.+`).ReplaceAllString(str, "")

	if buf, err := hex.DecodeString(str); err != nil {
		panic(err)
	} else {
		return buf
	}
}

var testGetNextRequestPacket = decodeTestPacket(`
30 21 											-- SEQUENCE
02 01 01 										-- INTEGER version
04 06 70 75 62 6c 69 63 		-- OCTET STRING community
a1 14												-- GetNextRequest-PDU
  02 02 05 39									-- INTEGER request-id
  02 01 00										-- INTEGER error-status
  02 01 00										-- INTEGER error-index
  30 08												-- SEQUENCE variable-bindings
    30 06												-- SEQUENCE
      06 02 2b 06									-- OID name
      05 00												-- NULL value
`)

var testGetResponsePacket = decodeTestPacket(`
30 38 02 01 01 04 06 70 75 62 6c 69 63 a2 2b 02
04 01 7a 6d f3 02 01 00 02 01 00 30 1d 30 1b 06
08 2b 06 01 02 01 01 05 00 04 0f 55 42 4e 54 20
45 64 67 65 53 77 69 74 63 68
`)

func TestPacketMarshal(t *testing.T) {
	var packet = Packet{Version: SNMPv2c, Community: []byte("public")}
	var pdu = PDU{
		RequestID: 1337,
		VarBinds: []VarBind{
			VarBind{Name: OID{1, 3, 6}},
		},
	}

	if packedPDU, err := pdu.pack(GetNextRequestType); err != nil {
		t.Errorf("pdu.pack: %v", err)
		return
	} else {
		packet.RawPDU = packedPDU
	}

	if bytes, err := packet.marshal(); err != nil {
		t.Errorf("packet.marshal: %v", err)
	} else {
		assert.Equal(t, testGetNextRequestPacket, bytes)
	}
}

type packetUnmarshalTest struct {
	data    []byte
	packet  Packet
	pduType PDUType
	pdu     PDU
	values  []interface{}
}

func testPacketUnmarshal(t *testing.T, test packetUnmarshalTest) {
	var packet Packet
	var pdu PDU

	if err := packet.unmarshal(test.data); err != nil {
		t.Errorf("packet.unmarshal: %v", err)
		return
	}

	if err := pdu.unpack(packet.RawPDU); err != nil {
		t.Errorf("pdu.unpack: %v", err)
		return
	}

	assert.Equal(t, test.packet.Version, packet.Version)
	assert.Equal(t, test.packet.Community, packet.Community)
	assert.Equal(t, test.pduType, packet.PDUType())
	assert.Equal(t, test.pdu.RequestID, pdu.RequestID)
	assert.Equal(t, test.pdu.ErrorStatus, pdu.ErrorStatus)
	assert.Equal(t, test.pdu.ErrorIndex, pdu.ErrorIndex)

	for i, varBind := range pdu.VarBinds {
		if i >= len(test.pdu.VarBinds) {
			t.Errorf("extra varBind[%d]: %#v", i, varBind)
			continue
		}

		assert.Equal(t, test.pdu.VarBinds[i].Name, varBind.Name, "VarBinds[i].Name", i)

		if value, err := varBind.Value(); err != nil {
			t.Errorf("varBind[%d].Value: %s", i, err)
			continue
		} else if i >= len(test.values) {
			t.Fatalf("missing test.values for varBind[%d]", i)
		} else {
			assert.Equal(t, test.values[i], value, "VarBinds[i].Value", i)
		}
	}
}

func TestPacketUnmarshal(t *testing.T) {
	testPacketUnmarshal(t, packetUnmarshalTest{
		data: decodeTestPacket(`
        30 38 02 01 01 04 06 70 75 62 6c 69 63 a2 2b 02
        04 01 7a 6d f3 02 01 00 02 01 00 30 1d 30 1b 06
        08 2b 06 01 02 01 01 05 00 04 0f 55 42 4e 54 20
        45 64 67 65 53 77 69 74 63 68
    `),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetResponseType,
		pdu: PDU{
			RequestID: 24800755,
			VarBinds: []VarBind{
				VarBind{
					Name: OID{1, 3, 6, 1, 2, 1, 1, 5, 0},
				},
			},
		},
		values: []interface{}{
			[]byte("UBNT EdgeSwitch"),
		},
	})
}

func TestPacketUnmarshalCounter32(t *testing.T) {
	testPacketUnmarshal(t, packetUnmarshalTest{
		data: decodeTestPacket(`
			30 30 02 01 01 04 06 70 75 62 6c 69 63 a2 23 02
			04 29 9e 37 ef 02 01 00 02 01 00 30 15 30 13 06
			0a 2b 06 01 02 01 02 02 01 0a 01 41 05 00 a8 dc
			8b 3b
		`),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetResponseType,
		pdu: PDU{
			RequestID: 698234863,
			VarBinds: []VarBind{
				VarBind{
					Name: OID{1, 3, 6, 1, 2, 1, 2, 2, 1, 10, 1},
				},
			},
		},
		values: []interface{}{
			Counter32(2833025851),
		},
	})
}

func TestPacketUnmarshalNoSuchInstance(t *testing.T) {
	testPacketUnmarshal(t, packetUnmarshalTest{
		data: decodeTestPacket(`
			30 29 02 01 01 04 06 70 75 62 6c 69 63 a2 1c 02
			04 47 6b 38 88 02 01 00 02 01 00 30 0e 30 0c 06
			08 2b 06 01 02 01 01 05 01 81 00
		`),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetResponseType,
		pdu: PDU{
			RequestID: 1198209160,
			VarBinds: []VarBind{
				VarBind{
					Name: OID{1, 3, 6, 1, 2, 1, 1, 5, 1},
				},
			},
		},
		values: []interface{}{
			NoSuchInstanceValue,
		},
	})
}
