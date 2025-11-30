//go:build windows
// +build windows

package WinDivert

import (
	"runtime"

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

// minInt 返回较小的整数
func minInt(a, b int) int {
	if a < b { // 如果 a 小于 b
		return a // 返回 a
	}
	return b // 否则返回 b
}
func calcMSS(v4 bool) int {
	if v4 { // 如果是 IPv4
		return 1460 // 典型以太网 MTU 1500 - IPv4 20 - TCP 20 = 1460
	}
	return 1440 // IPv6 40 + TCP 20，1500 - 40 - 20 = 1440
}

// SendDataToClient ：按 MSS 分段发送；写 TUN 时不持锁；仅最后一段置 PSH
func SendDataToClient(d *DevConn, payload []byte) (int, error) { // 对外发送函数
	if len(payload) == 0 { // 没有数据
		return 0, nil // 直接返回
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	mss := calcMSS(d.v4) // 计算单段最大负载
	// 发送循环变量
	total := len(payload) // 总长度
	offset := 0           // 当前偏移
	sent := 0             // 成功写入的 payload 字节数
	// 为了减轻内核队列压力，批量发送若干段后让出调度
	const burstSeg = 8 // 每发送 8 段让出一次

	// 主循环：直到全部发送完成
	for segIdx := 0; offset < total; segIdx++ {
		// 按段发送
		// 计算本段范围
		remain := total - offset     // 剩余字节
		chunk := minInt(mss, remain) // 本段大小不超过 MSS
		end := offset + chunk        // 本段结束位置
		psh := end == total          // 仅最后一段置 PSH
		// ---- 锁内：读取快照、构造首部并拿到本段 Seq/Ack ----
		seq := d.serverSeqNext     // 当前段 Seq 起点
		ack := d.clientNext        // 当前 Ack 值
		serverIP := d.serverIP     // 源 IP
		clientIP := d.clientIP     // 目的 IP
		serverPort := d.serverPort // 源端口
		clientPort := d.clientPort // 目的端口
		v4 := d.v4                 // 是否 IPv4
		seg := payload[offset:end] // 当前段载荷切片
		var err error
		if v4 { // IPv4 分支
			ip := &layers.IPv4{ // 构造 IPv4 首部
				Version:  4,                    // 版本
				IHL:      5,                    // 无选项 IHL=5
				SrcIP:    serverIP,             // 源 IP
				DstIP:    clientIP,             // 目的 IP
				Protocol: layers.IPProtocolTCP, // 上层协议 TCP
				TTL:      64,                   // TTL
			}
			tcp := &layers.TCP{ // 构造 TCP 首部
				SrcPort: layers.TCPPort(serverPort), // 源端口
				DstPort: layers.TCPPort(clientPort), // 目的端口
				Seq:     seq,                        // 本段 Seq
				Ack:     ack,                        // 本段 Ack
				ACK:     true,                       // ACK 位
				PSH:     psh,                        // 仅最后一段置 PSH
				Window:  65535,                      // 窗口（与接收窗口无关）
			}
			err = sendIPv4(d.h, ip, tcp, seg, d.lastAddr, false)
		} else { // IPv6 分支
			ip6 := &layers.IPv6{ // 构造 IPv6 首部
				Version:    6,                    // 版本
				SrcIP:      serverIP,             // 源 IP
				DstIP:      clientIP,             // 目的 IP
				NextHeader: layers.IPProtocolTCP, // 下一头部 TCP
				HopLimit:   64,                   // 跳限
			}
			tcp6 := &layers.TCP{ // 构造 TCP 首部
				SrcPort: layers.TCPPort(serverPort), // 源端口
				DstPort: layers.TCPPort(clientPort), // 目的端口
				Seq:     seq,                        // 本段 Seq
				Ack:     ack,                        // 本段 Ack
				ACK:     true,                       // ACK 位
				PSH:     psh,                        // 仅最后一段置 PSH
				Window:  65535,                      // 窗口
			}
			err = sendIPv6(d.h, ip6, tcp6, seg, d.lastAddr, false)
		}
		// 先在“状态上推进”下一发送序列（成功后生效；失败再回滚）
		nextSeq := seq + uint32(chunk) // 预计算下一 seq
		// 锁内不写 TUN，先解锁让收包线程有机会推进窗口

		// ---- 锁外：执行实际写入（避免长时间持锁阻塞收包）----
		if err != nil { // 写 TUN 失败
			// 写失败需要回到锁内回滚 serverSeqNext
			d.serverSeqNext = seq // 回滚到当前段起始 Seq
			return sent, err      // 返回已发送字节及错误
		}

		// ---- 锁内：确认写成功后，更新状态并推进 offset/sent ----
		d.serverSeqNext = nextSeq // 提交推进后的 Seq
		// 推进偏移与累计成功字节
		offset = end  // 偏移前移到下一段起点
		sent += chunk // 累计成功写入的 payload 字节数

		// 每写若干段，让出一下调度，减少内核队列压力
		if segIdx%burstSeg == burstSeg-1 { // 达到一批次
			runtime.Gosched() // 让出调度给收包 goroutine
		}
	}
	// 全部成功
	return sent, nil // 返回成功写入的 payload 字节总数
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
