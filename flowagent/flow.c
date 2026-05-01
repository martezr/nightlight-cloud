//go:build ignore
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

char LICENSE[] SEC("license") = "GPL";

#define TC_ACT_OK 0

#define ETH_P_IP     0x0800
#define ETH_P_ARP    0x0806
#define ETH_P_8021Q  0x8100
#define ETH_P_8021AD 0x88A8

#define IPPROTO_TCP  6
#define IPPROTO_UDP  17
#define IPPROTO_ICMP 1

#define MAX_PAYLOAD 256

struct flow_event {
    __u8  src_mac[6];
    __u8  dst_mac[6];

    __u32 src_ip;
    __u32 dst_ip;

    __u16 src_port;
    __u16 dst_port;

    __u8  protocol;
    __u8  icmp_type;
    __u8  icmp_code;
    __u8  pad;

    __u32 bytes;

    __u16 payload_len;
    __u8  payload[MAX_PAYLOAD];
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

SEC("tc")
int flow_monitor(struct __sk_buff *skb)
{
    void *data     = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    // ---- Ethernet ----
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;

    struct flow_event *evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
    if (!evt)
        return TC_ACT_OK;

    __builtin_memset(evt, 0, sizeof(*evt));

    // MACs (allowed for fixed size)
    __builtin_memcpy(evt->src_mac, eth->h_source, 6);
    __builtin_memcpy(evt->dst_mac, eth->h_dest, 6);

    evt->bytes = skb->len;

    __u16 proto = bpf_ntohs(eth->h_proto);
    void *cursor = eth + 1;

    // ---- VLAN ----
    if (proto == ETH_P_8021Q || proto == ETH_P_8021AD) {
        struct vlan_hdr {
            __be16 tci;
            __be16 encap_proto;
        };

        struct vlan_hdr *vh = cursor;
        if ((void *)(vh + 1) > data_end)
            goto submit;

        proto = bpf_ntohs(vh->encap_proto);
        cursor = vh + 1;
    }

    // ---- ARP ----
    if (proto == ETH_P_ARP) {
        evt->protocol = 0;
        goto submit;
    }

    // ---- IPv4 only ----
    if (proto != ETH_P_IP)
        goto submit;

    struct iphdr *ip = cursor;
    if ((void *)(ip + 1) > data_end)
        goto submit;

    __u32 ip_len = ip->ihl * 4;
    if (ip_len < sizeof(*ip))
        goto submit;

    if ((void *)ip + ip_len > data_end)
        goto submit;

    evt->src_ip = bpf_ntohl(ip->saddr);
    evt->dst_ip = bpf_ntohl(ip->daddr);
    evt->protocol = ip->protocol;

    void *l4 = (void *)ip + ip_len;

    // ---- ICMP ----
    if (ip->protocol == IPPROTO_ICMP) {
        struct icmphdr *icmp = l4;
        if ((void *)(icmp + 1) > data_end)
            goto submit;

        evt->icmp_type = icmp->type;
        evt->icmp_code = icmp->code;
        goto submit;
    }

// ---- TCP (ADD PAYLOAD EXTRACTION) ----
if (ip->protocol == IPPROTO_TCP) {
    struct tcphdr *tcp = l4;
    if ((void *)(tcp + 1) > data_end)
        goto submit;

    evt->src_port = bpf_ntohs(tcp->source);
    evt->dst_port = bpf_ntohs(tcp->dest);

    __u32 tcp_len = tcp->doff * 4;
    if (tcp_len < sizeof(*tcp))
        goto submit;

    void *payload = (void *)tcp + tcp_len;
    if (payload >= data_end)
        goto submit;

    __u64 offset = payload - data;
    if (offset >= skb->len)
        goto submit;

    __u32 remaining = skb->len - offset;

    if (remaining < 1)
        goto submit;

    if (remaining > MAX_PAYLOAD)
        remaining = MAX_PAYLOAD;

    if (bpf_skb_load_bytes(skb, offset, evt->payload, remaining) == 0) {
        evt->payload_len = remaining;
    }

    goto submit;
}

// ---- UDP (payload extraction - verifier safe) ----
if (ip->protocol == IPPROTO_UDP) {
    struct udphdr *udp = l4;
    if ((void *)(udp + 1) > data_end)
        goto submit;

    evt->src_port = bpf_ntohs(udp->source);
    evt->dst_port = bpf_ntohs(udp->dest);

    // Compute payload offset
    __u64 payload_offset = (void *)(udp + 1) - data;

    if (payload_offset >= skb->len)
        goto submit;

    __u32 remaining = skb->len - payload_offset;

    // 🔒 Hard clamp range for verifier
    if (remaining < 1)
        goto submit;

    if (remaining > MAX_PAYLOAD)
        remaining = MAX_PAYLOAD;

    // 🔒 Now remaining is guaranteed: [1, MAX_PAYLOAD]
    if (bpf_skb_load_bytes(skb, payload_offset, evt->payload, remaining) == 0) {
        evt->payload_len = remaining;
    }
}

submit:
    bpf_ringbuf_submit(evt, 0);
    return TC_ACT_OK;
}