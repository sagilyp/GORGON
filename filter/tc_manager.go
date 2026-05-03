package filter

import (
    "encoding/binary"
    "fmt"
    "net"

    "github.com/vishvananda/netlink"
    "golang.org/x/sys/unix"

    ebpfobjs "tor-filtering/filter/ebpf"
    "tor-filtering/logger"
)

type TCDirection int

const (
    TCDirectionIngress TCDirection = 0
    TCDirectionEgress  TCDirection = 1
)

type TCManager struct {
    Logger    logger.Logger
    objs      ebpfobjs.TcFilterObjects
    direction TCDirection
}

func NewTCManager(l logger.Logger, dir TCDirection) (*TCManager, error) {
	m := &TCManager{
		Logger:    l,
		direction: dir,
	}

	if err := ebpfobjs.LoadTcFilterObjects(&m.objs, nil); err != nil {
		return nil, fmt.Errorf("loading TC eBPF: %w", err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		m.objs.Close()
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	attachedCount := 0
	dirStr := "ingress"
	parent := uint32(netlink.HANDLE_MIN_INGRESS)
	if dir == TCDirectionEgress {
		parent = netlink.HANDLE_MIN_EGRESS
		dirStr = "egress"
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 { // || iface.Flags&net.FlagLoopback != 0
			continue
		}

		qdisc := &netlink.GenericQdisc{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: iface.Index,
				Handle:    netlink.MakeHandle(0xffff, 0),
				Parent:    netlink.HANDLE_CLSACT,
			},
			QdiscType: "clsact",
		}
		_ = netlink.QdiscReplace(qdisc)

		filter := &netlink.BpfFilter{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: iface.Index,
				Parent:    parent,
				Handle:    1,
				Protocol:  unix.ETH_P_ALL,
			},
			Fd:           m.objs.TcTorFilter.FD(),
			Name:         "tc_tor_filter",
			DirectAction: true,
		}

		if err := netlink.FilterReplace(filter); err != nil {
			m.Logger.Logf("WARN: Could not attach TC to %s: %v", iface.Name, err)
			continue
		}

		attachedCount++
		m.Logger.Logf("TC eBPF attached to %s (%s)", iface.Name, dirStr)
	}

	if attachedCount == 0 {
		m.objs.Close()
		return nil, fmt.Errorf("failed to attach TC filter to any active interface")
	}

	return m, nil
}

func (m *TCManager) AddGuardIPs(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		val := uint8(1)
		if err := m.objs.GuardIps.Put(key, val); err != nil {
			return fmt.Errorf("add guard %s: %w", ipStr, err)
		}
	}
	return nil
}

func (m *TCManager) AddBridgeIPs(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		val := uint8(1)
		if err := m.objs.BridgeIps.Put(key, val); err != nil {
			return fmt.Errorf("add bridge %s: %w", ipStr, err)
		}
	}
	return nil
}

func (m *TCManager) RemoveGuardIPs(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		_ = m.objs.GuardIps.Delete(key)
	}
	return nil
}

func (m *TCManager) RemoveBridgeIPs(_ string, ips []string) error {
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue
		}
		key := binary.LittleEndian.Uint32(ip.To4())
		_ = m.objs.BridgeIps.Delete(key)
	}
	return nil
}

func (m *TCManager) ClearGuardIPs(_ string) {
	var key uint32
	it := m.objs.GuardIps.Iterate()
	for it.Next(&key, nil) {
		_ = m.objs.GuardIps.Delete(key)
	}
}

func (m *TCManager) ClearBridgeIPs(_ string) {
	var key uint32
	it := m.objs.BridgeIps.Iterate()
	for it.Next(&key, nil) {
		_ = m.objs.BridgeIps.Delete(key)
	}
}

func (m *TCManager) Close() error {
	m.objs.Close()
	return nil
}