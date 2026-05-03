//go:build ignore

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h> 
#include <linux/in.h> 
#include <bpf/bpf_helpers.h>

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 20000);
    __type(key, __u32);
    __type(value, __u8);
} guard_ips SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 20000);
    __type(key, __u32);
    __type(value, __u8);
} bridge_ips SEC(".maps");

const __u32 *unused_guard_key __attribute__((unused));
const __u8  *unused_guard_val __attribute__((unused));
const __u32 *unused_bridge_key __attribute__((unused));
const __u8  *unused_bridge_val __attribute__((unused));

SEC("tc")
int tc_tor_filter(struct __sk_buff *skb) {
    void *data_end = (void *)(long)skb->data_end;
    void *data = (void *)(long)skb->data;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;
    
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return TC_ACT_OK;
    
    struct iphdr *ip = (struct iphdr *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return TC_ACT_OK;
    
    if (ip->protocol != IPPROTO_TCP)
        return TC_ACT_OK;

    struct tcphdr *tcp = (struct tcphdr *)((void *)ip + ip->ihl * 4);
    if ((void *)(tcp + 1) > data_end)
        return TC_ACT_OK;

    if (!(tcp->syn) || tcp->ack)
        return TC_ACT_OK;

    
    __u32 dst = ip->daddr;
    __u8 *is_guard  = bpf_map_lookup_elem(&guard_ips, &dst);
    __u8 *is_bridge = bpf_map_lookup_elem(&bridge_ips, &dst);

    if (is_guard) {
        bpf_printk("TOR_DETECT_GUARD: SYN to %pI4\n", &dst);
        return TC_ACT_SHOT;
    } else if (is_bridge) {
        bpf_printk("TOR_DETECT_BRIDGE: SYN to %pI4\n", &dst);
        return TC_ACT_SHOT;
    }
    return TC_ACT_OK; // TC_ACT_OK
}

char LICENSE[] SEC("license") = "GPL";
