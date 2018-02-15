package main

import (
    "../murmur3"
    "fmt"
    "time"
    "../cuckoo"
)


const (
    numPkt = 40000000
    n = 131072
    m = 128
    d = 64
    r = 2
)

type packet struct {
    flowID uint64
    size uint64
    ts uint64
}

type LLFD struct {
    buf []uint64
    counter []uint64
    entries [n]cuckoo.Entry

    fastCyc uint32
    fastCycSz uint64
    fastStart uint64

    slowCyc uint32
    slowCycSz uint64
    slowStart uint64
    slowFStart uint32
    slowCounter []uint64

    sampled uint64

    ctable *cuckoo.CuckooTable
}

type perf struct {
    start time.Time
    elapsed time.Duration
}

func (p *perf) Start() {
    p.start = time.Now()
}
func (p *perf) End() {
    p.elapsed += time.Since(p.start)
}
func (p *perf) GetElapsed() time.Duration {
    return p.elapsed
}

var rng = getRng(uint32(time.Now().UnixNano()))

func main() {
    // var pkts [numPkt]packet
    // genPackets(&pkts)
    var p1 perf
    var pkt packet
    // var counter [128]uint64

    p1.Start()
    dtctr := NewLLFD()
    for i := 0; i < numPkt+1; i++ {
        pkt.flowID = uint64(rng() & (n-1))
        pkt.size = func() uint64 { if pkt.flowID == 0 { return r } else { return 1 } }()
        pkt.flowID++
        pkt.ts = uint64(i*25)
        // idx := murmur3.Murmur3_32(pkt.flowID, dtctr.fastCyc) & (m - 1)
        // counter[idx] += pkt.size
        if dtctr.Recv(&pkt) {
            break
        }
    }
    // elapsed := time.Since(start)
    // fmt.Println(elapsed)
    // start = time.Now()
    // dtctr.fastCyc++
    // dtctr.detect()
    p1.End()
    fmt.Println(p1.GetElapsed())
}

func (dtctr *LLFD) Recv(pkt *packet) bool {
    if pkt.ts - dtctr.fastStart >= dtctr.fastCycSz {
        // fmt.Println(dtctr.counter[:m])
        dtctr.counter = dtctr.nextCounter(dtctr.counter)
        for i := 0; i < m; i++ {
            dtctr.counter[i] = 0
        }
        dtctr.fastCyc += 1
        dtctr.fastStart = pkt.ts
    }
    if pkt.ts - dtctr.slowStart >= dtctr.slowCycSz {
        if dtctr.detect() {
            // return true
        }
        dtctr.slowCyc += 1
        dtctr.slowStart = pkt.ts
        dtctr.slowFStart = dtctr.fastCyc
        dtctr.sampled = 0
        dtctr.slowCounter = dtctr.counter
    }
    idx := murmur3.Murmur3_32(pkt.flowID, dtctr.fastCyc) & (m - 1)
    dtctr.counter[idx] += pkt.size

    expected := (pkt.ts - dtctr.slowStart) >> 11
    if dtctr.sampled < expected {
        dtctr.ctable.Insert(pkt.flowID)
        dtctr.sampled += 1
    }
    return false
}

func (dtctr *LLFD) detect() bool {
    entries := dtctr.entries[:]
    cc := 0
    for i := range dtctr.ctable.Entries {
        entry := dtctr.ctable.Entries[i]
        if entry.Key == 0 {
            continue
        }
        entries[cc] = entry
        cc++
    }
    entries = entries[:cc]

    counter := dtctr.slowCounter
    for t := dtctr.slowFStart; t < dtctr.fastCyc; t++ {
        var counterSize [m]uint64
        for i := range entries {
            idx := murmur3.Murmur3_32(entries[i].Key, t) & (m - 1)
            counterSize[idx]++
        }
        for i := range entries {
            entry := &entries[i]
            idx := murmur3.Murmur3_32(entry.Key, t) & (m - 1)
            entry.Value += float64(counter[idx]) / float64(counterSize[idx])
        }
        // fmt.Println(counter[:m])
        // fmt.Println(counterSize)
        fmt.Println(counter[0] / counterSize[0])
        counter = dtctr.nextCounter(counter)
    }
    max := 0.0
    for i := range dtctr.ctable.Entries {
        entry := &dtctr.ctable.Entries[i]
        if entry.Key == 0 {
            continue
        }
        *entry = entries[0]
        entries = entries[1:]
        if entry.Value >= max {
            max = entry.Value
        }
    }
    // for _, entry := range entries {
        // if entry.Value >= max {
            // max = entry.Value
        // }
    // }
    val, _ := dtctr.ctable.LookUp(1)
    fmt.Println(max, val)
    return max <= val
}
func NewLLFD() *LLFD {
    dtctr := LLFD{
        buf: make([]uint64, m*d*2),
        fastCycSz: 15625000,
        slowCycSz: 1000000000,// / 4,
        ctable: cuckoo.NewCuckoo(n*4),
    }
    dtctr.counter = dtctr.buf[:]
    dtctr.slowCounter = dtctr.counter
    return &dtctr
}

func (dtctr *LLFD) nextCounter(counter []uint64) []uint64 {
    counter = counter[m:]
    if cap(counter) < m {
        counter = dtctr.buf[:]
    }
    return counter
}

func genPacket(i int) *packet {
    flowID := uint64(rng() & (n-1))
    size := func() uint64 { if flowID == 0 { return 10 } else { return 1 } }()
    return &packet{flowID+1, size, uint64(i*25)}
}

func genPackets(pkts *[numPkt]packet) {
    // cnt := 0
    //rng := getRng(uint32(time.Now().UnixNano()))
    for i := 0; i < numPkt; i++ {
        flowID := uint64(rng() & (n-1))
        // if flowID == 0 {
            // cnt++
        // }
        size := func() uint64 { if flowID == 0 { return 10 } else { return 1 } }()
        pkts[i] = packet{flowID+1, size, uint64(i * 25)}
    }
    // fmt.Println(cnt)
}

func getRng(seed uint32) func() uint32 {
    return func() uint32 {
        seed ^= seed << 13
        seed ^= seed >> 17
        seed ^= seed << 5
        return seed
    }
}
