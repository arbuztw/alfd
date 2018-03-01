package flowgen

import (
    "math/rand"
    // "../rand"
    "time"
)


type Packet struct {
    flowID uint64
    size   uint64
    ts     uint64    // timestamp in ns
}

type FlowGenerator struct {
    flowIDs []uint64
    idx int
    timestamp uint64
    timedelta uint64
    rate float64
}

func ResetSeed() {
    rand.Seed(time.Now().UnixNano())
}

func NewFlowGenerator(numFlow, numPkt int, r float64) *FlowGenerator {
    k := 3
    flowIDs := make([]uint64, numFlow*k)

    idx := 0
    for i := 1; i <= numFlow; i++ {
        for j := 0; j < k; j++ {
            flowIDs[idx] = uint64(i)
            idx++
        }
    }
    return &FlowGenerator{
        flowIDs: flowIDs,
        idx: 0,
        timestamp: 0,
        timedelta: uint64(1000000000 / numPkt),
        rate: r,
    }
}

func (g *FlowGenerator) Next() Packet {
    if g.idx == 0 {
        shuffle(g.flowIDs)
    }

    pkt := Packet{}
    pkt.flowID = g.flowIDs[g.idx]
    if pkt.flowID == 1 {
        pkt.size = uint64(g.rate * 10)
    } else {
        pkt.size = 10
    }
    pkt.ts = g.timestamp

    g.timestamp += g.timedelta
    g.idx++
    if g.idx >= len(g.flowIDs) {
        g.idx = 0
    }

    return pkt
}

func shuffle(s []uint64) {
    rand.Shuffle(len(s), func(i, j int) {
        s[i], s[j] = s[j], s[i]
    })
}

func (p *Packet) GetID() uint64 {
    return p.flowID
}

func (p *Packet) GetSize() uint64 {
    return p.size
}

func (p *Packet) GetTs() uint64 {
    return p.ts
}

/*
type heapElem struct {
    key uint32
    val uint64
}

type randHeap struct {
    data []heapElem
}

func (hp *randHeap) Len() int {
    return len(hp.data)
}

func Less(i, j int) bool {
    return hp.data[i] < hp.data[j]
}

func Swap(i, j int) {
    tmp := hp.data[i]
    hp.data[i] = hp.data[j]
    hp.data[j] = tmp
}

func (hp *randHeap) Push(x uint64) {
    hp.data = append(hp.data, heapElem{x, rand.rand()})
}

func (hp *randHeap) Pop() uint64 {
    last := len(hp.data) - 1
    ret := hp.data[last]
    hp.data = hp.data[:last]
    return ret
}
*/
