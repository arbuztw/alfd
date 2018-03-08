package main

import (
    "../alfd"
    "fmt"
    "flag"
    // "math/rand"
    "../flowgen"
    // "log"
    // "time"
    "os"
    // "github.com/netsec-ethz/lfd/eardet"
)

const (
    //n     = 131072   // number of flowID
    // pktps = 40000000 // number of packets per sec.
    // sec   = 10       // how many secs.
    // r     = 2        // large flow rate
    // trial = 50
)

var (
    n int
    pktps int
    r float64
)
/*
type packet struct {
    flowID uint64
    size   uint64
    ts     uint64    // timestamp in ns
}

var rng = getRng(uint32(time.Now().UnixNano()))
*/
func main() {
    nptr := flag.Int("n", 131072, "Number of flows")
    pptr := flag.Int("pkt", 305, "Number of packets per second per legal flow")
    rptr := flag.Float64("r", 2, "Large flow ratio")
    tptr := flag.Int("t", 50, "Number of trials")
    optr := flag.String("o", "", "Log file")
    flag.Parse()

    n = *nptr
    pktps = *pptr
    r = *rptr
    trial := *tptr

    if *optr == "" {
        fmt.Println("Please specify log file name")
        return
    }
    f, err := os.OpenFile(*optr, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        fmt.Println("Cannot open file:", *optr)
        return
    }
    defer f.Close()
    fmt.Fprintln(f, "r =", r, "n =", n, "pkt/s =", pktps)

    delay := 0.0
    for t := 0; t < trial; t++ {
        g := flowgen.NewFlowGenerator(n, pktps, uint64(pktps / n), 1, r)
        dtctr := alfd.NewAlfd(uint32(n), 1000000000 / 16, 1000000000 / 4)
        // dtctr := eardet.NewConfigedEardetDtctr(1024, 10000, 640, 1000, 10000000)
        i := 0
        for {
            pkt := g.Next()//genPacket(i)
            if dtctr.Recv(pkt.GetID(), pkt.GetSize(), pkt.GetTs()) {
            // if dtctr.Detect(uint32(pkt.GetID()), uint32(pkt.GetSize()), time.Duration(pkt.GetTs())) {
                ts := float64(pkt.GetTs()) / 1000000000.0
                fmt.Println(t, ts)
                fmt.Fprintln(f, ts)
                delay += ts
                break
            }
            i = i + 1
        }
    }

    fmt.Println("r =", r, "n =", n, "pkt/s = ", pktps)
    fmt.Println(delay / float64(trial))

    // elapsed := time.Since(start)
    // fmt.Println(elapsed)
}
/*
func genPacket(i int) packet {
    interval := 1000000000 / pktps

    mod := uint32(float64(n) + r - 1) * 10

    var flowID uint64
    if id := rng() % mod; id < (n - 1) * 10 {
        flowID = uint64(id / 10 + 2)
    } else {
        flowID = 1
    }
    size := uint64(1) //func() uint64 { if flowID == 0 { return 1 } else { return 10 } }()
    ts := uint64(i * interval)
    flowID := uint64(rng() % n + 1)
    size := flowID % 2
    if flowID == 1 {
        size = uint64(2)
    }
    ts := uint64(i * interval)

    return packet{flowID, size, ts}
}

func genPackets(pkts []packet, num int) {
    for i := 0; i < num; i++ {
        pkts[i] = genPacket(i)
    }
}

func shuffle(s []packet) {
    rand.Shuffle(len(s), func(i, j int) {
        s[i], s[j] = s[j], s[i]
    })
}

func getRng(seed uint32) func() uint32 {
    return func() uint32 {
        seed ^= seed << 13
        seed ^= seed >> 17
        seed ^= seed << 5
        return seed
    }
}
*/
