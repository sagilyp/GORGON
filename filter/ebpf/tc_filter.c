//go:build ignore

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <bpf/bpf_helpers.h>

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 20000);
    __type(key, __u32);
    __type(value, __u8);
} blocked_ips SEC(".maps");

// ✅ Глобальная ссылка на типы для bpf2go
const __u32 *unused_blocked_key __attribute__((unused));
const __u8 *unused_blocked_value __attribute__((unused));

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
    
    // Проверка dst IP (для egress фильтрации guard/bridge)
    // Для ingress используем saddr, для egress - daddr
    __u32 check_ip = skb->ifindex == 1 ? ip->saddr : ip->daddr; // 1 = ingress
    
    __u8 *blocked = bpf_map_lookup_elem(&blocked_ips, &check_ip);
    if (blocked) {
        bpf_printk("TOR_BLOCK: Dropped packet to %pI4\n", &check_ip);
        return TC_ACT_SHOT; // drop
    }
    
    return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";
