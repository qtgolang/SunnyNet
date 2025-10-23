//go:build windows
// +build windows

package WinDivert

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// 公共序列化选项（复用，避免每次重复构造）
var serializeOpts = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

// helper: 序列化并发送 IPv4 包（tcp + 可选 payload）
// outbound 控制是否把 addr 标记为 outbound (true) 或 inbound (false)
func sendIPv4(h *Handle, ip *layers.IPv4, tcp *layers.TCP, payload []byte, addr *Address, outbound bool) error {
	_ = tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	if payload != nil && len(payload) > 0 {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp, gopacket.Payload(payload)); err != nil {
			return err
		}
	} else {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp); err != nil {
			return err
		}
	}
	out := addr.Clone()
	out.SetOutbound(outbound)
	_, err := h.Send(buf.Bytes(), out)
	return err
}

// helper: 序列化并发送 IPv6 包（tcp + 可选 payload）
func sendIPv6(h *Handle, ip *layers.IPv6, tcp *layers.TCP, payload []byte, addr *Address, outbound bool) error {
	_ = tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	if payload != nil && len(payload) > 0 {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp, gopacket.Payload(payload)); err != nil {
			return err
		}
	} else {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp); err != nil {
			return err
		}
	}
	out := addr.Clone()
	out.SetOutbound(outbound)
	_, err := h.Send(buf.Bytes(), out)
	return err
}

// SendSynAckToClient ：收到 client SYN 时注入 SYN/ACK（伪造 server 的 SYN/ACK）
func SendSynAckToClient(h *Handle, d *DevConn, addr *Address, clientISN uint32) error {
	// 复制关键字段（无需长时间持锁）
	d.mu.Lock()
	serverIP := d.serverIP
	clientIP := d.clientIP
	serverPort := d.serverPort
	clientPort := d.clientPort
	serverISN := d.serverISN
	v4 := d.v4
	d.mu.Unlock()
	if v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    serverIP,
			DstIP:    clientIP,
			Protocol: layers.IPProtocolTCP,
			TTL:      64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(serverPort),
			DstPort: layers.TCPPort(clientPort),
			Seq:     serverISN,
			Ack:     clientISN + 1,
			SYN:     true,
			ACK:     true,
			Window:  65535,
		}
		return sendIPv4(h, ip, tcp, nil, addr, false) // inbound from server -> client => outbound=false
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      serverIP,
		DstIP:      clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(serverPort),
		DstPort: layers.TCPPort(clientPort),
		Seq:     serverISN,
		Ack:     clientISN + 1,
		SYN:     true,
		ACK:     true,
		Window:  65535,
	}
	return sendIPv6(h, ip6, tcp6, nil, addr, false)
}

// SendAckToKernel ：向内核注入 ACK，通知内核我们已接收 client 的数据，避免内核重传
func SendAckToKernel(h *Handle, d *DevConn, clientNext uint32, addr *Address) error {
	// 只在短时间内持锁读取 seq
	d.mu.Lock()
	seq := d.serverSeqNext
	// copy v4 flag and endpoints
	v4 := d.v4
	serverIP := d.serverIP
	clientIP := d.clientIP
	serverPort := d.serverPort
	clientPort := d.clientPort
	d.mu.Unlock()
	if v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    serverIP,
			DstIP:    clientIP,
			Protocol: layers.IPProtocolTCP,
			TTL:      64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(serverPort),
			DstPort: layers.TCPPort(clientPort),
			Seq:     seq,
			Ack:     clientNext,
			ACK:     true,
			Window:  65535,
		}
		if err := sendIPv4(h, ip, tcp, nil, addr, true); err == nil {
			d.mu.Lock()
			if clientNext > d.highestClientAckSent {
				d.highestClientAckSent = clientNext
			}
			d.mu.Unlock()
			return nil
		} else {
			return err
		}
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      serverIP,
		DstIP:      clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(serverPort),
		DstPort: layers.TCPPort(clientPort),
		Seq:     d.serverSeqNext,
		Ack:     clientNext,
		ACK:     true,
		Window:  65535,
	}
	if err := sendIPv6(h, ip6, tcp6, nil, addr, true); err == nil {
		d.mu.Lock()
		if clientNext > d.highestClientAckSent {
			d.highestClientAckSent = clientNext
		}
		d.mu.Unlock()
		return nil
	} else {
		return err
	}
}

// SendDataToClient ：把伪 server 要发送的数据注入给 client，正确设置 Seq/Ack 并更新 serverSeqNext
func SendDataToClient(h *Handle, d *DevConn, payload []byte, addr *Address) (int, error) {
	// 读取必要字段（缩短锁持有时间）
	d.mu.Lock()
	seq := d.serverSeqNext
	ack := d.clientNext
	serverIP := d.serverIP
	clientIP := d.clientIP
	serverPort := d.serverPort
	clientPort := d.clientPort
	v4 := d.v4
	d.mu.Unlock()

	if v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    serverIP,
			DstIP:    clientIP,
			Protocol: layers.IPProtocolTCP,
			TTL:      64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(serverPort),
			DstPort: layers.TCPPort(clientPort),
			Seq:     seq,
			Ack:     ack,
			ACK:     true,
			PSH:     true,
			Window:  65535,
		}
		if err := sendIPv4(h, ip, tcp, payload, addr, false); err != nil {
			return 0, err
		}
		// 更新 serverSeqNext
		d.mu.Lock()
		d.serverSeqNext += uint32(len(payload))
		d.mu.Unlock()
		return len(payload), nil
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      serverIP,
		DstIP:      clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(serverPort),
		DstPort: layers.TCPPort(clientPort),
		Seq:     seq,
		Ack:     ack,
		ACK:     true,
		PSH:     true,
		Window:  65535,
	}
	if err := sendIPv6(h, ip6, tcp6, payload, addr, false); err != nil {
		return 0, err
	}
	d.mu.Lock()
	d.serverSeqNext += uint32(len(payload))
	d.mu.Unlock()
	return len(payload), nil
}

// SendFinToClient ：注入一个 FIN/ACK（server -> client），并在成功后把 serverSeqNext 增 1（FIN 消耗 1 序号）。
func SendFinToClient(h *Handle, d *DevConn) error {
	// 复制需要的字段，避免在持锁时调用 h.Send 导致死锁或长时间阻塞
	d.mu.Lock()
	seq := d.serverSeqNext
	ack := d.clientNext
	serverIP := d.serverIP
	clientIP := d.clientIP
	serverPort := d.serverPort
	clientPort := d.clientPort
	lastAddr := d.lastAddr
	v4 := d.v4
	d.mu.Unlock()

	if lastAddr == nil {
		return nil
	}

	if v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    serverIP,
			DstIP:    clientIP,
			Protocol: layers.IPProtocolTCP,
			TTL:      64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(serverPort),
			DstPort: layers.TCPPort(clientPort),
			Seq:     seq,
			Ack:     ack,
			FIN:     true,
			ACK:     true,
			Window:  65535,
		}
		if err := sendIPv4(h, ip, tcp, nil, lastAddr, false); err != nil {
			return err
		}
		d.mu.Lock()
		if seq == d.serverSeqNext {
			d.serverSeqNext = seq + 1
		}
		d.mu.Unlock()
		return nil
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      serverIP,
		DstIP:      clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(serverPort),
		DstPort: layers.TCPPort(clientPort),
		Seq:     seq,
		Ack:     ack,
		FIN:     true,
		ACK:     true,
		Window:  65535,
	}
	if err := sendIPv6(h, ip6, tcp6, nil, lastAddr, false); err != nil {
		return err
	}
	d.mu.Lock()
	if seq == d.serverSeqNext {
		d.serverSeqNext = seq + 1
	}
	d.mu.Unlock()
	return nil
}

// SendRstToClient ：立刻强制断开（client 会收到 RST），通常客户端会马上重连
func SendRstToClient(h *Handle, d *DevConn) error {
	d.mu.Lock()
	seq := d.serverSeqNext
	ack := d.clientNext
	serverIP := d.serverIP
	clientIP := d.clientIP
	serverPort := d.serverPort
	clientPort := d.clientPort
	lastAddr := d.lastAddr
	v4 := d.v4
	d.mu.Unlock()

	if lastAddr == nil {
		return nil
	}

	if v4 {
		ip := &layers.IPv4{
			Version: 4, IHL: 5,
			SrcIP: serverIP, DstIP: clientIP,
			Protocol: layers.IPProtocolTCP, TTL: 64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(serverPort),
			DstPort: layers.TCPPort(clientPort),
			Seq:     seq,
			Ack:     ack,
			RST:     true,
			Window:  0,
		}
		return sendIPv4(h, ip, tcp, nil, lastAddr, false)
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      serverIP,
		DstIP:      clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(serverPort),
		DstPort: layers.TCPPort(clientPort),
		Seq:     seq,
		Ack:     ack,
		RST:     true,
		Window:  0,
	}
	return sendIPv6(h, ip6, tcp6, nil, lastAddr, false)
}
