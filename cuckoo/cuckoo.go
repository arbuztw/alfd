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
    "math"
)

// The key has to be uint32. But one can adjust the value to whatever type one wants.
type Entry struct {
    Key   uint64
    Size     uint64
    Num      uint64
    LastSeen uint64
}

type CuckooTable struct {
    Entries   []Entry
    seed      uint32
    idxBytes  uint32
    nEntries  uint32
    nRehashes uint32
    tableLen  uint32

    marking   []uint64
    serial    uint64
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
    c := &CuckooTable{
        Entries:  make([]Entry, size),
        idxBytes: uint32(math.Log2(float64(size))),
        tableLen: size,
        marking:  make([]uint64, size),
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

// Lookup the entry. Insert the key if it is not exist when `insert` is set to
// be true.
func (c *CuckooTable) LookupI(key uint64, insert bool) (*Entry, bool) {
    h1, h2 := c.getHashedKeys(key)
    if entry := &c.Entries[h1]; entry.Key == key {
        return entry, true
    }
    if entry := &c.Entries[h2]; entry.Key == key {
        return entry, true
    }

    if !insert {
        return nil, false
    }

    newEntry := Entry{key, 0, 0, 0}
    index := h1
    retIndex := index
    c.serial += 1

    for {
        oldEntry := c.Entries[index]
        if oldEntry.Key == 0 {
            c.Entries[index] = newEntry
            break
        }
        if c.marking[index] == c.serial {  // cycle detected
            break
        }
        c.Entries[index] = newEntry
        c.marking[index] = c.serial

        h1, h2 = c.getHashedKeys(oldEntry.Key)

        if index == h1 {
            index = h2
        } else {
            index = h1
        }

        newEntry = oldEntry
    }

    return &c.Entries[retIndex], true
}
