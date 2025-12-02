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
    iface     *net.Interface
    direction TCDirection
}

func NewTCManager(l logger.Logger, ifaceName string, dir TCDirection) (*TCManager, error) {
    m := &TCManager{
        Logger:    l,
        direction: dir,
    }

    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        return nil, fmt.Errorf("get interface %s: %w", ifaceName, err)
    }
    m.iface = iface

    if err := ebpfobjs.LoadTcFilterObjects(&m.objs, nil); err != nil {
        return nil, fmt.Errorf("loading TC eBPF: %w", err)
    }

    qdisc := &netlink.GenericQdisc{
        QdiscAttrs: netlink.QdiscAttrs{
            LinkIndex: iface.Index,
            Handle:    netlink.MakeHandle(0xffff, 0),
            Parent:    netlink.HANDLE_CLSACT,
        },
        QdiscType: "clsact",
    }
    if err := netlink.QdiscReplace(qdisc); err != nil {
        m.objs.Close()
        return nil, fmt.Errorf("create clsact qdisc: %w", err)
    }

    parent := uint32(netlink.HANDLE_MIN_INGRESS)
    if dir == TCDirectionEgress {
        parent = netlink.HANDLE_MIN_EGRESS
    }

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
        m.objs.Close()
        return nil, fmt.Errorf("attach TC filter: %w", err)
    }

    dirStr := "ingress"
    if dir == TCDirectionEgress {
        dirStr = "egress"
    }
    m.Logger.Logf("TC eBPF attached to %s (%s)", ifaceName, dirStr)
    return m, nil
}

// IPSet API над BPF map

func (m *TCManager) AddIPsToIPSet(_ string, ips []string) error {
    for _, ipStr := range ips {
        ip := net.ParseIP(ipStr)
        if ip == nil || ip.To4() == nil {
            continue
        }
        key := binary.BigEndian.Uint32(ip.To4())
        val := uint8(1)
        if err := m.objs.BlockedIps.Put(key, val); err != nil {
            return fmt.Errorf("add %s: %w", ipStr, err)
        }
    }
    return nil
}

func (m *TCManager) RemoveIPsFromIPSet(_ string, ips []string) error {
    for _, ipStr := range ips {
        ip := net.ParseIP(ipStr)
        if ip == nil || ip.To4() == nil {
            continue
        }
        key := binary.BigEndian.Uint32(ip.To4())
        _ = m.objs.BlockedIps.Delete(key)
    }
    return nil
}

func (m *TCManager) ClearIPSet(_ string) {
    var key uint32
    it := m.objs.BlockedIps.Iterate()
    for it.Next(&key, nil) {
        _ = m.objs.BlockedIps.Delete(key)
    }
}

func (m *TCManager) Close() error {
    m.objs.Close()
    return nil
}