package main

import (
    "../alfd"
    "fmt"
    "time"
)

const (
    n     = 131072   // number of flowID
    pktps = 40000000 // number of packets per sec.
    sec   = 10       // how many secs.
    r     = 2        // large flow rate
    trial = 50
)

type packet struct {
    flowID uint64
    size   uint64
    ts     uint64    // timestamp in ns
}

var rng = getRng(uint32(time.Now().UnixNano()))

func main() {
    // numPkt := pktps * sec + 1

    // var pkts [numPkt]packet
    // genPackets(pkts[:], numPkt)

    // start := time.Now()

    delay := 0.0

    for t := 0; t < trial; t++ {
        dtctr := alfd.NewAlfd(1000000000 / 64, 1000000000 / 4)
        i := 0
        for {
            pkt := genPacket(i)
            if dtctr.Recv(pkt.flowID, pkt.size, pkt.ts) {
                ts := float64(pkt.ts)
                fmt.Println(ts / 1000000000.0)
                delay += ts
                break
            }
            i = i + 1
        }
    }

    fmt.Println(delay / float64(trial))

    // elapsed := time.Since(start)
    // fmt.Println(elapsed)
}

func genPacket(i int) packet {
    interval := 1000000000 / pktps

    flowID := /*uint64(i & (n - 1)) //*/uint64(rng() & (n - 1))
    size := func() uint64 { if flowID == 0 { return r } else { return 1 } }()
    ts := uint64(i * interval)

    return packet{flowID+1, size, ts}
}

func genPackets(pkts []packet, num int) {
    for i := 0; i < num; i++ {
        pkts[i] = genPacket(i)
    }
}

func getRng(seed uint32) func() uint32 {
    return func() uint32 {
        seed ^= seed << 13
        seed ^= seed >> 17
        seed ^= seed << 5
        return seed
    }
}
