package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/bitbrewers/tranx2"
	flags "github.com/jessevdk/go-flags"
)

type options struct {
	Transponders int64  `short:"t" long:"transponders" description:"Number of transponders" required:"true"`
	Interval     uint32 `short:"i" long:"interval" description:"Average interval between writing passings per transponder in ms" required:"true"`
	Jitter       uint32 `short:"j" long:"jitter" description:"Maximum jitter between writing passings per transponder in ms"`
	Clock        uint32 `short:"c" long:"clock" description:"Set start value for internal ticker"`
	Seed         int64  `short:"s" long:"seed" description:"Seed for generating transponder IDs"`
}

func main() {
	opts := options{}
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		fmt.Println()
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	w := tranx2.NewWriter(os.Stdout)

	go generateNoise(w)

	started := time.Now()
	ponders := generateTransponders(opts.Transponders, opts.Seed)
	for _, tID := range ponders {
		go generatePassings(w, opts.Clock, opts.Interval, opts.Jitter, tID, time.Now().UnixNano(), started)
	}
	select {}
}

func generatePassings(w *tranx2.Writer, clock, interval, jitter, transponderID uint32, seed int64, started time.Time) {
	r := rand.New(rand.NewSource(seed))
	for i := uint32(1); true; i++ {
		passingTicks := (interval * i) + (r.Uint32() % (jitter + 1)) - (jitter / 2)
		time.Sleep(time.Until(started.Add(time.Millisecond * time.Duration(passingTicks))))
		if _, err := w.WritePassing(
			tranx2.Passing{
				TransponderID: transponderID,
				PassingTicks:  clock + passingTicks,
				Hits:          uint8(r.Uint32()%10) + 5,
				Strength:      uint8(r.Uint32()%50) + 80,
				Prefix:        uint16(r.Uint32()),
				Trailing:      uint8(r.Uint32()),
			}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func generateNoise(w *tranx2.Writer) {
	for {
		time.Sleep(time.Second * 5)
		if _, err := w.WriteNoise(2000); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func generateTransponders(amount, seed int64) (transponderIDs []uint32) {
	r := rand.New(rand.NewSource(seed))
	for i := int64(0); i < amount; i++ {
		transponderIDs = append(transponderIDs, r.Uint32()%(tranx2.MaxTransponderID+1))
	}
	return
}
