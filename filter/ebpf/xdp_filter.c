//go:build ignore

#include <linux/bpf.h>
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
} blocked_ips SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} drop_counter SEC(".maps");

const __u32 *unused_blocked_key __attribute__((unused));
const __u8 *unused_blocked_value __attribute__((unused));
const __u32 *unused_counter_key __attribute__((unused));
const __u64 *unused_counter_value __attribute__((unused));

SEC("xdp")
int xdp_tor_filter(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return XDP_PASS;
    
    struct iphdr *ip = (struct iphdr *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    if (ip->protocol != IPPROTO_TCP)
        return XDP_PASS;

    struct tcphdr *tcp = (struct tcphdr *)((void *)ip + ip->ihl * 4);
    if ((void *)(tcp + 1) > data_end)
        return XDP_PASS;

    if (!(tcp->syn) || tcp->ack)
        return XDP_PASS;

    __u8 *blocked = bpf_map_lookup_elem(&blocked_ips, &ip->saddr); // src IP для ingress
    if (blocked) {
        __u32 key = 0;
        __u64 *counter = bpf_map_lookup_elem(&drop_counter, &key);
        if (counter)
            *counter += 1;
        
        bpf_printk("TOR_DETECT_EXIT: SYN from %pI4\n", &ip->saddr);
        return XDP_DROP;
    }
    return XDP_PASS; // XDP_PASS
}

char LICENSE[] SEC("license") = "GPL";
