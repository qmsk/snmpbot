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
		packet.PDU = packedPDU
	}

	if bytes, err := packet.marshal(); err != nil {
		t.Errorf("packet.marshal: %v", err)
	} else {
		assert.Equal(t, testGetNextRequestPacket, bytes)
	}
}

func TestPacketUnmarshal(t *testing.T) {
	var packet Packet
	var pdu PDU

	var expectedPacket = Packet{
		Version:   SNMPv2c,
		Community: []byte("public"),
	}
	var expectedPDU = PDU{
		RequestID: 24800755,
		VarBinds: []VarBind{
			VarBind{
				Name:  OID{1, 3, 6, 1, 2, 1, 1, 5, 0},
				Value: []byte("UBNT EdgeSwitch"),
			},
		},
	}

	if err := packet.unmarshal(testGetResponsePacket); err != nil {
		t.Errorf("packet.unmarshal: %v", err)
		return
	}

	if err := pdu.unpack(packet.PDU); err != nil {
		t.Errorf("pdu.unpack: %v", err)
		return
	}

	assert.Equal(t, expectedPacket.Version, packet.Version)
	assert.Equal(t, expectedPacket.Community, packet.Community)
	assert.Equal(t, GetResponseType, packet.PDUType())
	assert.Equal(t, expectedPDU, pdu)
}
