package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"code.rocket9labs.com/tslocum/tabula"
)

func main() {
	var pips bool
	flag.BoolVar(&pips, "pips", false, "Print table of pseudopip values")
	flag.Parse()

	if pips {
		fmt.Println("| Space | Pseudopips |")
		fmt.Println("| --- | --- |")
		for space := 1; space <= 25; space++ {
			fmt.Printf("| %d | %d |\n", space, tabula.PseudoPips(1, space))
		}
		return
	}

	b := tabula.NewBoard(true)
	b[tabula.SpaceRoll1] = 6
	b[tabula.SpaceRoll2] = 6
	b[tabula.SpaceRoll3] = 6
	b[tabula.SpaceRoll4] = 6
	b.Print()

	t := time.Now()
	available, _ := b.Available(1)
	t2 := time.Now()
	analysis := make([]*tabula.Analysis, 0, tabula.AnalysisBufferSize)
	b.Analyze(available, &analysis)
	t3 := time.Now()

	log.Println("AVAILABLE TOOK ", t2.Sub(t))
	log.Println("ANALYSIS TOOK ", t3.Sub(t2))

	log.Println("AVAILABLE", available)
	for i, a := range analysis {
		log.Printf("%+v %+v", i, a)
	}
}
