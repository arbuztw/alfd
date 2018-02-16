package alfd

import (
    "../cuckoo"
    "../murmur3"
    "fmt"
)

const (
    n = 131072 // number of flowID
    m = 128    // number of counter per fast cycle
)

type Alfd struct {
    buf     []uint64
    counter []uint64
    entries [n]cuckoo.Entry

    fastCyc   uint32  // id of current fast cycle
    fastCycSz uint64  // duration of a fast cycle
    fastStart uint64  // starting time of current fast cycle

    slowCyc     uint32  // id of current slow cycle
    slowCycSz   uint64  // duration of a slow cycle
    slowStart   uint64  // starting time of current slow cycle
    slowFStart  uint32  // id of the first fast cycle in current slow cycle
    slowCounter []uint64

    sampled uint64

    ctable *cuckoo.CuckooTable
}

func NewAlfd(fastCycSz, slowCycSz uint64) *Alfd {
    d := slowCycSz / fastCycSz + 1 // number of fast cycles in a slow cycle
    dtctr := Alfd{
        buf:       make([]uint64, m*d),
        fastCycSz: fastCycSz,
        slowCycSz: slowCycSz,
        ctable:    cuckoo.NewCuckoo(n * 4),
    }
    dtctr.counter = dtctr.buf[:]
    dtctr.slowCounter = dtctr.counter
    return &dtctr
}

func (dtctr *Alfd) Recv(flowID, size, ts uint64) {
    // advanced to next fast cycle
    if ts - dtctr.fastStart >= dtctr.fastCycSz {
        dtctr.counter = dtctr.nextCounter(dtctr.counter)
        for i := 0; i < m; i++ {
            dtctr.counter[i] = 0
        }
        dtctr.fastCyc += 1
        dtctr.fastStart = ts
    }
    // advanced to next slow cycle
    if ts - dtctr.slowStart >= dtctr.slowCycSz {
        dtctr.detect(ts)

        dtctr.slowCyc += 1
        dtctr.slowStart = ts
        dtctr.slowFStart = dtctr.fastCyc
        dtctr.sampled = 0
        dtctr.slowCounter = dtctr.counter
    }

    // counting
    idx := murmur3.Murmur3_32(flowID, dtctr.fastCyc) & (m - 1)
    dtctr.counter[idx] += size

    // sub-sample flowIDs
    expected := (ts - dtctr.slowStart) >> 9
    if dtctr.sampled < expected {
        value, _ := dtctr.ctable.LookupI(flowID, true)
        value.LastSeen = ts
        dtctr.sampled += 1
    }
}

func (dtctr *Alfd) detect(now uint64) bool {
    entries := dtctr.entries[:]
    cc := 0
    for i := range dtctr.ctable.Entries {
        entry := &dtctr.ctable.Entries[i]
        if entry.Key == 0 {
            continue
        }
        /*if now - entry.LastSeen > dtctr.slowCycSz * 2 {
            entry.Key = 0
            continue
        }*/
        entries[cc] = *entry
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
            entry.Size += counter[idx]
            entry.Num += counterSize[idx]
        }
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
        val := float64(entry.Size) / float64(entry.Num)
        if val >= max {
            max = val
        }
    }
    entry, _ := dtctr.ctable.LookupI(1, false) // the large flow
    val := 0.0
    if entry != nil {
        val = float64(entry.Size) / float64(entry.Num)
    }
    fmt.Println(max, val, max <= val)
    return max <= val
}

func (dtctr *Alfd) nextCounter(counter []uint64) []uint64 {
    counter = counter[m:]
    if cap(counter) < m {
        counter = dtctr.buf[:]
    }
    return counter
}
