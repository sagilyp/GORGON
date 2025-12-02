package ebpf

// XDP
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-global-types xdpFilter xdp_filter.c

// TC
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-global-types tcFilter tc_filter.c
