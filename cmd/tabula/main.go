package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	// should be 3/1 then 5/0 in acey, then fix backgammon movement
	// test cases for issues first then fix until tests pass
	b := tabula.Board{0, 5, 2, 5, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, -1, 0, 0, -2, -2, -2, -3, -3, -2, 0, 1, 0, 0, 0, 5, 2, 0, 0, 1, 1, 1}
	log.Println(b[24])
	b.Print()
	available, _ := b.Available(1)
	for i := range available {
		log.Println(available[i])
	}
	b.Analyze(available, &analysis)
	for i := range analysis {
		log.Println(analysis[i])
	}
	log.Println(b)
	b = b.UseRoll(3, 1, 1).Move(3, 1, 1)
	log.Println("NEW AVAILABLE")
	log.Println(b)
	available, _ = b.Available(1)
	for i := range available {
		log.Println(available[i])
	}
	log.Println(b)
	os.Exit(0)

	// Print opening moves.
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
