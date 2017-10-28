package snmp

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func decodeTestHex(str string) []byte {
	str = regexp.MustCompile(`\s+|--.+`).ReplaceAllString(str, "")

	if buf, err := hex.DecodeString(str); err != nil {
		panic(err)
	} else {
		return buf
	}
}

func TestPacketMarshal(t *testing.T) {
	var packet = Packet{Version: SNMPv2c, Community: []byte("public")}
	var pdu = PDU{
		RequestID: 1337,
		VarBinds: []VarBind{
			VarBind{Name: OID{1, 3, 6}},
		},
	}
	var expected = decodeTestHex(`
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

	if packedPDU, err := pdu.pack(GetNextRequestType); err != nil {
		t.Errorf("pdu.pack: %v", err)
		return
	} else {
		packet.PDU = packedPDU
	}

	if bytes, err := packet.marshal(); err != nil {
		t.Errorf("packet.marshal: %v", err)
	} else {
		assert.Equal(t, expected, bytes)
	}
}
