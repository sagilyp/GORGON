package filter

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/cilium/ebpf/link"
	ebpfobjs "tor-filtering/filter/ebpf"
	"tor-filtering/logger"
)

type XDPManager struct {
	Logger logger.Logger
	objs   ebpfobjs.XdpFilterObjects
	links  []link.Link
}

func NewXDPManager(l logger.Logger) (*XDPManager, error) {
	m := &XDPManager{
		Logger: l,
	}

	if err := ebpfobjs.LoadXdpFilterObjects(&m.objs, nil); err != nil {
		return nil, fmt.Errorf("loading XDP eBPF objects: %w", err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		m.objs.Close()
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 { // || iface.Flags&net.FlagLoopback != 0
			continue
		}

		l, err := link.AttachXDP(link.XDPOptions{
			Program:   m.objs.XdpTorFilter,
			Interface: iface.Index,
		})
		if err != nil {
			m.Logger.Logf("WARN: Could not attach XDP to %s: %v", iface.Name, err)
			continue
		}

		m.links = append(m.links, l)
		m.Logger.Logf("XDP program successfully attached to %s", iface.Name)
	}

	if len(m.links) == 0 {
		m.objs.Close()
		return nil, fmt.Errorf("failed to attach XDP to any active network interface")
	}

	return m, nil
}

func (m *XDPManager) AddIPsToIPSet(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		val := uint8(1)
		if err := m.objs.BlockedIps.Put(key, val); err != nil {
			return fmt.Errorf("add %s to BPF map: %w", ipStr, err)
		}
	}
	m.Logger.Logf("Added %d IPs to XDP Blocked map", len(ips))
	return nil
}

func (m *XDPManager) RemoveIPsFromIPSet(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		_ = m.objs.BlockedIps.Delete(key)
	}
	return nil
}

func (m *XDPManager) ClearIPSet(_ string) {
	var key uint32
	it := m.objs.BlockedIps.Iterate()
	for it.Next(&key, nil) {
		_ = m.objs.BlockedIps.Delete(key)
	}
}

func (m *XDPManager) Close() error {
	for _, l := range m.links {
		l.Close()
	}
	m.objs.Close()
	return nil
}