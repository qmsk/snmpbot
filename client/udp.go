package client

import (
	"fmt"
	"io"
	"net"
	"syscall"
)

const UDPPort = "161"
const UDPSize uint = 64 * 1024

type UDPOptions struct {
	Size uint
}

func makeUDP(options UDPOptions) UDP {
	if options.Size == 0 {
		options.Size = UDPSize
	}

	return UDP{
		size: options.Size,
	}
}

func resolveUDP(addr string) (*net.UDPAddr, error) {
	if _, port, _ := net.SplitHostPort(addr); port == "" {
		addr = net.JoinHostPort(addr, UDPPort)
	}

	return net.ResolveUDPAddr("udp", addr)
}

func NewUDP(options UDPOptions) (*UDP, error) {
	var udp = makeUDP(options)

	if udpConn, err := net.ListenUDP("udp", &net.UDPAddr{}); err != nil {
		return nil, err
	} else {
		udp.conn = udpConn
	}

	if udpAddr, err := udp.LocalAddr(); err != nil {
		return nil, err
	} else {
		udp.addr = udpAddr
	}

	return &udp, nil
}

func ListenUDP(addr string, options UDPOptions) (*UDP, error) {
	var udp = makeUDP(options)

	if udpAddr, err := resolveUDP(addr); err != nil {
		return nil, err
	} else if udpConn, err := net.ListenUDP("udp", udpAddr); err != nil {
		return nil, err
	} else {
		udp.addr = udpAddr
		udp.conn = udpConn
	}

	return &udp, nil
}

func DialUDP(addr string, options UDPOptions) (*UDP, error) {
	var udp = makeUDP(options)

	if udpAddr, err := resolveUDP(addr); err != nil {
		return nil, err
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		return nil, err
	} else {
		udp.addr = udpAddr
		udp.conn = udpConn
	}

	return &udp, nil
}

type UDP struct {
	size uint
	addr *net.UDPAddr
	conn *net.UDPConn
}

func (udp *UDP) String() string {
	return fmt.Sprintf("%v", udp.addr)
}

func (udp *UDP) LocalAddr() (*net.UDPAddr, error) {
	switch localAddr := udp.conn.LocalAddr().(type) {
	case *net.UDPAddr:
		return localAddr, nil
	default:
		return nil, fmt.Errorf("Unknown LocalAddr type %T", localAddr)
	}
}

func (udp *UDP) Resolve(addr string) (net.Addr, error) {
	return resolveUDP(addr)
}

func (udp *UDP) send(buf []byte, addr net.Addr) error {
	var size int
	var err error

	if addr == nil {
		size, err = udp.conn.Write(buf)
	} else {
		size, err = udp.conn.WriteTo(buf, addr)
	}

	if err != nil {
		return err
	} else if size != len(buf) {
		return fmt.Errorf("short write: %d < %d", size, len(buf))
	}

	return nil
}

func (udp *UDP) Send(send IO) error {
	if err := send.Packet.PackPDU(send.PDUType, send.PDU); err != nil {
		return ProtocolError{fmt.Errorf("packet.PackPDU: %v", err)}
	} else if buf, err := send.Packet.Marshal(); err != nil {
		return ProtocolError{fmt.Errorf("packet.Marshal: %v", err)}
	} else if err := udp.send(buf, send.Addr); err != nil {
		return err
	}

	return nil
}

func (udp *UDP) Recv() (recv IO, err error) {
	var buf = make([]byte, udp.size)

	// recv
	if size, _, flags, addr, err := udp.conn.ReadMsgUDP(buf, nil); err != nil {
		return recv, err
	} else if size == 0 {
		return recv, io.EOF
	} else if flags&syscall.MSG_TRUNC != 0 {
		return recv, ProtocolError{fmt.Errorf("Packet truncated (>%d bytes)", udp.size)}
	} else {
		recv.Addr = addr
		buf = buf[:size]
	}

	if err := recv.Packet.Unmarshal(buf); err != nil {
		return recv, ProtocolError{fmt.Errorf("packet.Unmarshal: %v", err)}
	}

	if pduType, pdu, err := recv.Packet.UnpackPDU(); err != nil {
		return recv, ProtocolError{fmt.Errorf("packet.UnpackPDU: %v", err)}
	} else {
		recv.PDUType = pduType
		recv.PDU = pdu
	}

	return recv, nil
}

func (udp *UDP) Close() error {
	return udp.conn.Close()
}
