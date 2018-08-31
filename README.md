# AMB TranX-2 golang implementation

[![Build Status](https://travis-ci.com/bitbrewers/tranx2.svg?branch=master)](https://travis-ci.com/bitbrewers/tranx2)
[![GoDoc](https://godoc.org/github.com/bitbrewers/tranx2?status.svg)](https://godoc.org/github.com/bitbrewers/tranx2)
[![codecov](https://codecov.io/gh/bitbrewers/tranx2/branch/master/graph/badge.svg)](https://codecov.io/gh/bitbrewers/tranx2)
[![Go Report Card](https://goreportcard.com/badge/github.com/bitbrewers/tranx2)](https://goreportcard.com/report/github.com/bitbrewers/tranx2)

This package provides Marshal and Unmarshal functionalities for TranX-2 encoded data and some tools for debugging and testing AMB TranX-2 decoder devices and software.

### Tools

- tranx2sim - generate tranx2 encoded data
- tranx2dump - show tranx2 encoded data in human readable format

## Examples

### Client

```go
package main

import (
	"log"

	"github.com/bitbrewers/tranx2"
)

type Handler struct{}

func (h *Handler) OnPassing(rec tranx2.Passing) { log.Println(rec) }
func (h *Handler) OnNoise(noise uint16)         { log.Println(noise) }
func (h *Handler) OnError(err error)            { log.Println(err) }

func main() {
	c := tranx2.NewClient("/dev/USBtty0", &Handler{})
	log.Fatal(c.Listen())
}

```

### Reader

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bitbrewers/tranx2"
)

func main() {
	r := tranx2.NewReader(os.Stdin)
	for {
		record, err := r.ReadPassing()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("%+v\n", record)
	}
}
```

### Writer

```go
package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bitbrewers/tranx2"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticker := time.NewTicker(time.Second)
	started := time.Now()

	w := tranx2.NewWriter(os.Stdout)
	for t := range ticker.C {
		record := tranx2.Passing{
			TransponderID: 12335,
			PassingTicks:  uint32(t.Sub(started).Nanoseconds() / int64(time.Millisecond)),
			Hits:          uint8(r.Uint32()%10) + 5,
			Strength:      uint8(r.Uint32()%50) + 80,
			Prefix:        uint16(r.Uint32()),
			Trailing:      uint8(r.Uint32()),
		}
		_, err := w.WritePassing(record)
		if err != nil {
			log.Fatal(err)
		}
	}
}
```
