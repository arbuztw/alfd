package flowgen

import (
    "../rand"
)


type Packet struct {
    flowID uint64
    size   uint64
    ts     uint64    // timestamp in ns
}

type flow struct {
    id   uint64
    npkt uint64
    tick uint64
}

type flowQueue struct {
    buf []flow
    fr int
    bk int
}

type FlowGenerator struct {
    numFlow  uint64
    avail    []flow
    overuse  flowQueue
    numAvail int
    tick uint64
    timedelta uint64
    rate float64
    gamma uint64
    beta uint64
}

func NewFlowGenerator(numFlow, numPkt int, gamma, beta uint64, r float64) *FlowGenerator {
    g := FlowGenerator {
        numFlow: uint64(numFlow),
        avail: make([]flow, numFlow),
        numAvail: numFlow,
        timedelta: uint64(1000000000 / numPkt),
        beta: beta,
        rate: r,
    }
    g.overuse.setCap(numFlow+1)
    for i := 0; i < numFlow; i++ {
        g.avail[i] = flow { id: uint64(i+1) }
    }
    return &g
}

func (g *FlowGenerator) Next() Packet {
    for !g.overuse.isEmpty() {
        flow := g.overuse.top()
        if g.tick - flow.tick < g.numFlow {
            break
        }
        g.avail[g.numAvail] = g.overuse.pop()
        g.numAvail++
        flow = g.overuse.top()
    }

    k := int(rand.Rand()) % g.numAvail
    flow := &g.avail[k]
    if flow.npkt == 0 {
        flow.tick = g.tick
    }
    flow.npkt += 1
    size := uint64(10)
    if flow.id == 1 {
        size = uint64(10 * g.rate)
    }
    pkt := Packet{flow.id, size, g.tick * g.timedelta}

    if diff := g.tick - flow.tick; diff >= g.numFlow {
        leak := diff / g.numFlow
        if leak <= flow.npkt {
            flow.npkt -= leak
        } else {
            flow.npkt = 0
        }
        flow.tick += leak * g.numFlow
    }

    if flow.npkt >= g.beta {
        g.overuse.push(*flow)
        g.avail[k] = g.avail[g.numAvail-1]
        g.numAvail--
    }

    g.tick++

    return pkt
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

func (q *flowQueue) setCap(capacity int) {
    q.buf = make([]flow, capacity)
}

func (q *flowQueue) push(f flow) {
    q.buf[q.bk] = f
    q.bk++
    if q.bk >= len(q.buf) {
        q.bk = 0
    }
}

func (q *flowQueue) pop() flow {
    ret := q.buf[q.fr]
    q.fr++
    if q.fr >= len(q.buf) {
        q.fr = 0
    }
    return ret
}

func (q *flowQueue) top() *flow {
    return &q.buf[q.fr]
}

func (q *flowQueue) isEmpty() bool {
    return q.fr == q.bk
}
