// Copyright 2017 Nicolas Forster
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file implements a cuckoo hash table using two different hashes of the
// key to find an appropriate place for the element to be inserted.
// The maximum table size is 2^16 elements.

package cuckoo

import (
    "crypto/rand"
    "../murmur3"
    "encoding/binary"
    "fmt"
    "os"
    // "runtime"
    "math"
    // "unsafe"
    //"bitbucket/cuckoohash/murmur"
)

const (
    maxLen      = 100
    minIdxBytes = 20
    maxLoadFact = 0.5
)

// The key has to be uint32. But one can adjust the value to whatever type one wants.
type Entry struct {
    Key   uint64
    Value float64
}

type CuckooTable struct {
    Entries   []Entry
    seed      uint32
    idxBytes  uint32
    nEntries  uint32
    nRehashes uint32
    tableLen  uint32
}

// resetSeed() resets the current seed. Used during a rehash of the table.
func (c *CuckooTable) resetSeed() {
    s := make([]byte, 4)
    _, err := rand.Read(s)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
    }

    c.seed = binary.LittleEndian.Uint32(s)
}

func NewCuckoo(size uint32) *CuckooTable {
    entries := make([]Entry, size)
    c := &CuckooTable{
        Entries:  entries,
        idxBytes: uint32(math.Log2(float64(size))),
        tableLen: size,
    }
    c.resetSeed()

    return c
}

// getHashedKeys() generates the two hashed keys. Important to note is that
// only one hash is generated. This hash is then split up into the two
// hashed key values used for inserting/finding an object.
func (c *CuckooTable) getHashedKeys(key uint64) (uint32, uint32) {
    hash := murmur3.Murmur3_32(key, 0)
    h1 := hash >> (32 - c.idxBytes)
    h2 := hash & ((1<<c.idxBytes)-1)
    return h1, h2
}

// LookUp() looks an element up in the table.
func (c *CuckooTable) LookUp(key uint64) (float64, bool) {
    h1, h2 := c.getHashedKeys(key)
    if entry := &c.Entries[h1]; entry.Key == key {
        return entry.Value, true
    }

    if entry := &c.Entries[h2]; entry.Key == key {
        return entry.Value, true
    }

    return 0, false
}

// Insert() inserts an element at the appropriate position in the table.
func (c *CuckooTable) Insert(key uint64/*, value float64*/) bool {
    h1, h2 := c.getHashedKeys(key)

    if entry := &c.Entries[h1]; entry.Key == key {
        return true
    }
    if entry := &c.Entries[h2]; entry.Key == key {
        return true
    }

    newEntry := Entry{key, 0}
    index := h1

    for {
        oldEntry := c.Entries[index]
        if oldEntry.Key == 0 {
            c.Entries[index] = newEntry
            return true
        }
        if oldEntry.Key == key {
            return false
        }
        c.Entries[index] = newEntry

        h1, h2 = c.getHashedKeys(oldEntry.Key)

        if index == h1 {
            index = h2
        } else {
            index = h1
        }

        newEntry = oldEntry
    }
}

// Delete() deletes an element.
/*func (c *CuckooTable) Delete(key uint32) {
    h1, h2 := c.getHashedKeys(key)
    if entry := c.Entries[h1]; entry != nil && entry.Key == key {
        c.Entries[h1] = nil
        c.nEntries -= 1
    }

    if entry := c.Entries[h2]; entry != nil && entry.Key == key {
        c.Entries[h2] = nil
        c.nEntries -= 1
    }

    // If the load factor of the table is too low, shrink the table.
    if c.LoadFactor() < maxLoadFact/2 {
        c.shrink()
    }

}

func (c *CuckooTable) rehash() {
    c.nEntries = 0
    c.nRehashes += 1
    c.reorganize()
}

func (c *CuckooTable) grow() {
    c.idxBytes += 1
    c.nEntries = 0
    c.nRehashes = 0

    if c.idxBytes > maxLen {
        panic("Too many elements")
    }

    c.reorganize()
}

func (c *CuckooTable) shrink() {
    if c.idxBytes <= minIdxBytes {
        return
    }
    c.idxBytes -= 1
    c.nEntries = 0
    c.nRehashes = 0

    c.reorganize()
}

func (c *CuckooTable) reorganize() {
    tempTab := &CuckooTable{}
    *tempTab = *c
    c.resetSeed()

    c.Entries = make([]*Entry, 1<<c.idxBytes)

    for _, val := range tempTab.Entries {
        if val != nil {
            c.Insert(val.Key, val.Value)
        }
    }

    defer func() {
        tempTab = nil
        runtime.GC()
    }()
}

func (c *CuckooTable) LoadFactor() float64 {
    tLen := 1 << c.idxBytes
    return float64(c.nEntries) / float64(tLen)
}

func (c *CuckooTable) GetNEntries() uint32 {
    return c.nEntries
}
*/
