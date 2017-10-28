package snmp

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func decode(str string) []byte {
	str = regexp.MustCompile(`\s+`).ReplaceAllString(str, "")

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
	var expected = decode(`
    30 21 02 01 01 04 06 70 75 62 6c 69 63 a1 14 02
    02 05 39 02 01 00 02 01 00 30 08 30 06 06 02 2b
    06 05 00
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
