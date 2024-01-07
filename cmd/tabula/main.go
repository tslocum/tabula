package main

import (
	"flag"
	"fmt"
	"log"

	"code.rocket9labs.com/tslocum/tabula"
)

func main() {
	var beiAddress string
	var pips bool
	flag.StringVar(&beiAddress, "bei", "", "Listen for BEI connections on specified address (TCP)")
	flag.BoolVar(&pips, "pips", false, "Print table of pseudopip values")
	flag.BoolVar(&tabula.Verbose, "verbose", false, "Print state of each request")
	flag.Parse()

	if pips {
		fmt.Println("| Space | Pseudopips |")
		fmt.Println("| --- | --- |")
		for space := int8(1); space <= 25; space++ {
			fmt.Printf("| %d | %d |\n", space, tabula.PseudoPips(1, space))
		}
		return
	}

	if beiAddress != "" {
		s := tabula.NewBEIServer()
		s.Listen(beiAddress)
	}

	//b := tabula.Board{0, 0, 0, 0, 0, -1, 8, 0, 4, 0, 0, 0, 0, 0, -1, -1, 0, -1, -1, -1, 1, -2, -2, -3, -2, 0, 2, 0, 0, 0, 0, 4, 1, 1, 0}

	//b := tabula.Board{0, 0, -2, -2, -2, 4, 0, -1, 0, 0, -2, 4, 0, -2, -1, 0, -2, 4, 0, 2, 0, 0, 0, 0, -1, 0, 1, 0, 4, 1, 0, 0, 1, 1, 1}

	analysis := make([]*tabula.Analysis, 0, tabula.AnalysisBufferSize)
	for r1 := 1; r1 <= 6; r1++ {
		for r2 := 1; r2 <= 6; r2++ {
			b := tabula.NewBoard(tabula.VariantBackgammon)
			b[tabula.SpaceRoll1] = int8(r1)
			b[tabula.SpaceRoll2] = int8(r2)
			if r1 == r2 {
				b[tabula.SpaceRoll3] = int8(r1)
				b[tabula.SpaceRoll4] = int8(r2)
			}
			available, _ := b.Available(1)
			b.Analyze(available, &analysis)
			log.Println("ROLL", r1, r2, analysis[0].Moves)
		}
	}
}
