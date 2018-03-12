package main

import (
    "../alfd"
    "fmt"
    "flag"
    "../flowgen"
    "os"
)

var (
    n int
    pktps int
    r float64
)

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
        i := 0
        for {
            pkt := g.Next()
            if dtctr.Recv(pkt.GetID(), pkt.GetSize(), pkt.GetTs()) {
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
}
