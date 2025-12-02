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
    Logger    logger.Logger
    objs      ebpfobjs.XdpFilterObjects
    link      link.Link
    ifaceName string
}

func NewXDPManager(l logger.Logger, ifaceName string) (*XDPManager, error) {
    m := &XDPManager{
        Logger:    l,
        ifaceName: ifaceName,
    }

    if err := ebpfobjs.LoadXdpFilterObjects(&m.objs, nil); err != nil {
        return nil, fmt.Errorf("loading XDP eBPF objects: %w", err)
    }

    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        m.objs.Close()
        return nil, fmt.Errorf("get interface %s: %w", ifaceName, err)
    }

    m.link, err = link.AttachXDP(link.XDPOptions{
        Program:   m.objs.XdpTorFilter,
        Interface: iface.Index,
    })
    if err != nil {
        m.objs.Close()
        return nil, fmt.Errorf("attach XDP: %w", err)
    }

    m.Logger.Logf("XDP program attached to %s", ifaceName)
    return m, nil
}


func (m *XDPManager) AddIPsToIPSet(_ string, ips []string) error {
    for _, ipStr := range ips {
        ip := net.ParseIP(ipStr)
        if ip == nil || ip.To4() == nil {
            continue
        }
        key := binary.BigEndian.Uint32(ip.To4())
        val := uint8(1)
        if err := m.objs.BlockedIps.Put(key, val); err != nil {
            return fmt.Errorf("add %s to BPF map: %w", ipStr, err)
        }
    }
    m.Logger.Logf("Added %d IPs to BPF map", len(ips))
    return nil
}

func (m *XDPManager) RemoveIPsFromIPSet(_ string, ips []string) error {
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

func (m *XDPManager) ClearIPSet(_ string) {
    var key uint32
    it := m.objs.BlockedIps.Iterate()
    for it.Next(&key, nil) {
        _ = m.objs.BlockedIps.Delete(key)
    }
}

func (m *XDPManager) Close() error {
    if m.link != nil {
        m.link.Close()
    }
    m.objs.Close()
    return nil
}