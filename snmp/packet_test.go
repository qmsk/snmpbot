package snmp

import (
	"encoding/asn1"
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

// prepare a VarBind for use with assert.Equal() in tests
func testVarBind(oid OID, value interface{}) VarBind {
	varBind := MakeVarBind(oid, value)

	if varBind.RawValue.Bytes == nil {
		varBind.RawValue.Bytes = []byte{}
	}

	if data, err := asn1.Marshal(varBind.RawValue); err != nil {
		panic(err)
	} else {
		varBind.RawValue.FullBytes = data
	}

	return varBind
}

type packetTest struct {
	bytes   []byte
	packet  Packet
	pduType PDUType
	pdu     PDU
}

func testPacketMarshal(t *testing.T, test packetTest) {
	if err := test.packet.PackPDU(test.pduType, test.pdu); err != nil {
		t.Fatalf("pdu.pack: %v", err)
	}

	if bytes, err := test.packet.Marshal(); err != nil {
		t.Fatalf("packet.marshal: %v", err)
	} else {
		assert.Equal(t, test.bytes, bytes)
	}
}

func testPacketUnmarshal(t *testing.T, test packetTest) {
	var packet Packet
	var pdu PDU

	err := packet.Unmarshal(test.bytes)
	if err != nil {
		t.Errorf("packet.Unmarshal: %v", err)
		return
	}

	pduType, pdu, err := packet.UnpackPDU()
	if err != nil {
		t.Errorf("packet.UnpackPDU: %v", err)
		return
	}

	assert.Equal(t, test.packet.Version, packet.Version)
	assert.Equal(t, test.packet.Community, packet.Community)
	assert.Equal(t, test.pduType, pduType)
	assert.Equal(t, test.pdu, pdu)
}

func testPacket(t *testing.T, test packetTest) {
	testPacketMarshal(t, test)
	testPacketUnmarshal(t, test)
}

func TestPacketGetRequest(t *testing.T) {
	testPacket(t, packetTest{
		bytes: decodeTestPacket(`
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
		`),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetNextRequestType,
		pdu: GenericPDU{
			RequestID: 1337,
			VarBinds: []VarBind{
				testVarBind(OID{1, 3, 6}, nil),
			},
		},
	})
}

func TestPacketGetResponse(t *testing.T) {
	testPacket(t, packetTest{
		bytes: decodeTestPacket(`
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
		pdu: GenericPDU{
			RequestID: 24800755,
			VarBinds: []VarBind{
				testVarBind(OID{1, 3, 6, 1, 2, 1, 1, 5, 0}, []byte("UBNT EdgeSwitch")),
			},
		},
	})
}

func TestPacketCounter32(t *testing.T) {
	testPacket(t, packetTest{
		bytes: decodeTestPacket(`
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
		pdu: GenericPDU{
			RequestID: 698234863,
			VarBinds: []VarBind{
				testVarBind(OID{1, 3, 6, 1, 2, 1, 2, 2, 1, 10, 1}, Counter32(2833025851)),
			},
		},
	})
}

func TestPacketNoSuchInstance(t *testing.T) {
	testPacket(t, packetTest{
		bytes: decodeTestPacket(`
			30 29 02 01 01 04 06 70 75 62 6c 69 63 a2 1c 02
			04 47 6b 38 88 02 01 00 02 01 00 30 0e 30 0c 06
			08 2b 06 01 02 01 01 05 01 81 00
		`),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetResponseType,
		pdu: GenericPDU{
			RequestID: 1198209160,
			VarBinds: []VarBind{
				testVarBind(OID{1, 3, 6, 1, 2, 1, 1, 5, 1}, NoSuchInstanceValue),
			},
		},
	})
}

func TestPacketGetBulk(t *testing.T) {
	testPacket(t, packetTest{
		bytes: decodeTestPacket(`
			30 39 02 01 01 04 06 70 75 62 6c 69 63 a5 2c 02
			04 2c 6a 76 19 02 01 00 02 01 0a 30 1e 30 0d 06
			09 2b 06 01 02 01 02 02 01 01 05 00 30 0d 06 09
			2b 06 01 02 01 02 02 01 02 05 00
		`),
		packet: Packet{
			Version:   SNMPv2c,
			Community: []byte("public"),
		},
		pduType: GetBulkRequestType,
		pdu: BulkPDU{
			RequestID:      745174553,
			NonRepeaters:   0,
			MaxRepetitions: 10,
			VarBinds: []VarBind{
				testVarBind(MustParseOID(".1.3.6.1.2.1.2.2.1.1"), nil),
				testVarBind(MustParseOID(".1.3.6.1.2.1.2.2.1.2"), nil),
			},
		},
	})
}
