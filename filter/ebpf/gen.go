package ebpf

// XDP
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-global-types XdpFilter xdp_filter.c

// TC
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-global-types TcFilter tc_filter.c
