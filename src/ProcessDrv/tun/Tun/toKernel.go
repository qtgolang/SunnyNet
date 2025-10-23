//go:build android || darwin || linux
// +build android darwin linux

package Tun

import (
	"io"
	"runtime"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var serializeOpts = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

func sendIPv4(ip *layers.IPv4, tcp *layers.TCP, payload []byte) []byte {
	_ = tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	if payload != nil && len(payload) > 0 {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp, gopacket.Payload(payload)); err != nil {
			return nil
		}
	} else {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp); err != nil {
			return nil
		}
	}
	return buf.Bytes()
}

func sendIPv6(ip *layers.IPv6, tcp *layers.TCP, payload []byte) []byte {
	_ = tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	if payload != nil && len(payload) > 0 {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp, gopacket.Payload(payload)); err != nil {
			return nil
		}
	} else {
		if err := gopacket.SerializeLayers(buf, serializeOpts, ip, tcp); err != nil {
			return nil
		}
	}
	return buf.Bytes()
}

// SendSynAckToClient ：收到 client SYN 时注入 SYN/ACK（伪造 server 的 SYN/ACK）
func SendSynAckToClient(h io.ReadWriteCloser, d *DevConn, clientISN uint32) (int, error) {

	if d.v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			SrcIP:    d.serverIP,
			DstIP:    d.clientIP,
			Protocol: layers.IPProtocolTCP,
			TTL:      64,
		}
		tcp := &layers.TCP{
			SrcPort: layers.TCPPort(d.serverPort),
			DstPort: layers.TCPPort(d.clientPort),
			Seq:     d.serverISN,
			Ack:     clientISN + 1,
			SYN:     true,
			ACK:     true,
			Window:  65535,
		}
		return h.Write(sendIPv4(ip, tcp, nil))
	}

	ip6 := &layers.IPv6{
		Version:    6,
		SrcIP:      d.serverIP,
		DstIP:      d.clientIP,
		NextHeader: layers.IPProtocolTCP,
		HopLimit:   64,
	}
	tcp6 := &layers.TCP{
		SrcPort: layers.TCPPort(d.serverPort),
		DstPort: layers.TCPPort(d.clientPort),
		Seq:     d.serverISN,
		Ack:     clientISN + 1,
		SYN:     true,
		ACK:     true,
		Window:  65535,
	}
	return h.Write(sendIPv6(ip6, tcp6, nil))
}

func SendAckToKernel(d *DevConn, clientNext uint32) []byte {
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
		bs := sendIPv4(ip, tcp, nil)
		d.mu.Lock()
		if clientNext > d.highestClientAckSent {
			d.highestClientAckSent = clientNext
		}
		d.mu.Unlock()
		return bs
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
	bs := sendIPv6(ip6, tcp6, nil)
	d.mu.Lock()
	if clientNext > d.highestClientAckSent {
		d.highestClientAckSent = clientNext
	}
	d.mu.Unlock()
	return bs
}

// SendFinToClient ：注入一个 FIN/ACK（server -> client），并在成功后把 serverSeqNext 增 1（FIN 消耗 1 序号）。
func SendFinToClient(d *DevConn) []byte {
	// 复制需要的字段，避免在持锁时调用 h.Send 导致死锁或长时间阻塞
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
			FIN:     true,
			ACK:     true,
			Window:  65535,
		}
		bs := sendIPv4(ip, tcp, nil)
		d.mu.Lock()
		if seq == d.serverSeqNext {
			d.serverSeqNext = seq + 1
		}
		d.mu.Unlock()
		return bs
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
	bs := sendIPv6(ip6, tcp6, nil)
	d.mu.Lock()
	if seq == d.serverSeqNext {
		d.serverSeqNext = seq + 1
	}
	d.mu.Unlock()
	return bs
}

// SendRstToClient ：立刻强制断开（client 会收到 RST），通常客户端会马上重连
func SendRstToClient(d *DevConn) []byte {
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
		return sendIPv4(ip, tcp, nil)
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
	return sendIPv6(ip6, tcp6, nil)
}
func buildTCPReply(d *DevConn, tcp *layers.TCP, sendAck bool, sendFin bool, sendRst bool) (int, error) {
	var replyIP gopacket.NetworkLayer
	var replyIPop gopacket.SerializableLayer
	if d.v4 {
		ip := &layers.IPv4{
			Version: 4, IHL: 5,
			SrcIP: d.serverIP, DstIP: d.clientIP,
			Protocol: layers.IPProtocolTCP, TTL: 64,
		}
		replyIP = ip
		replyIPop = ip
	} else {
		ip6 := &layers.IPv6{
			Version:    6,
			SrcIP:      d.serverIP,
			DstIP:      d.clientIP,
			NextHeader: layers.IPProtocolTCP,
			HopLimit:   64,
		}
		replyIP = ip6
		replyIPop = ip6
	}

	replyTCP := *tcp

	replyTCP.SrcPort, replyTCP.DstPort = tcp.DstPort, tcp.SrcPort

	// 清空 payload
	replyTCP.Payload = nil

	// 重置标志
	replyTCP.SYN = false
	replyTCP.ACK = sendAck
	replyTCP.FIN = sendFin
	replyTCP.RST = sendRst
	replyTCP.PSH = false
	replyTCP.URG = false
	replyTCP.ECE = false
	replyTCP.CWR = false
	// 序号与确认号
	replyTCP.Seq = tcp.Ack
	replyTCP.Ack = tcp.Seq + uint32(len(tcp.Payload))
	if tcp.SYN || tcp.FIN {
		replyTCP.Ack++
	}
	_ = replyTCP.SetNetworkLayerForChecksum(replyIP)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	_ = gopacket.SerializeLayers(buf, opts, replyIPop, &replyTCP)
	return d.tun.Write(buf.Bytes())
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

		var pkt []byte // 序列化后的 IP 包
		if v4 {        // IPv4 分支
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
			pkt = sendIPv4(ip, tcp, seg) // 序列化完整 IPv4 包
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
			pkt = sendIPv6(ip6, tcp6, seg) // 序列化完整 IPv6 包
		}
		// 先在“状态上推进”下一发送序列（成功后生效；失败再回滚）
		nextSeq := seq + uint32(chunk) // 预计算下一 seq
		// 锁内不写 TUN，先解锁让收包线程有机会推进窗口

		// ---- 锁外：执行实际写入（避免长时间持锁阻塞收包）----
		if _, err := d.tun.Write(pkt); err != nil { // 写 TUN 失败
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
