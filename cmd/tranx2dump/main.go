package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitbrewers/tranx2"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			os.Exit(1)
		}

		switch line[0] {
		case tranx2.NoisePrefix:
			if noise, err := tranx2.UnmarshalNoise(line); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s", err)
			} else {
				enc.Encode(struct{ Noise uint16 }{noise})
			}
		case tranx2.PassingPrefix:
			if rec, err := tranx2.UnmarshalPassing(line); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s", err)
			} else {
				enc.Encode(rec)
			}
		}
	}
}
