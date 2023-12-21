package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"code.rocket9labs.com/tslocum/tabula"
)

func main() {
	var address string
	var pips bool
	flag.StringVar(&address, "address", "", "Listen for BEI connections on specified address (TCP)")
	flag.BoolVar(&pips, "pips", false, "Print table of pseudopip values")
	flag.Parse()

	if pips {
		fmt.Println("| Space | Pseudopips |")
		fmt.Println("| --- | --- |")
		for space := int8(1); space <= 25; space++ {
			fmt.Printf("| %d | %d |\n", space, tabula.PseudoPips(1, space))
		}
		return
	}

	if address != "" {
		s := tabula.NewBEIServer()
		s.Listen(address)
	}

	//b := tabula.Board{0, 0, 0, 0, 0, -1, 8, 0, 4, 0, 0, 0, 0, 0, -1, -1, 0, -1, -1, -1, 1, -2, -2, -3, -2, 0, 2, 0, 0, 0, 0, 4, 1, 1, 0}

	b := tabula.Board{0, 0, -2, -2, -2, 4, 0, -1, 0, 0, -2, 4, 0, -2, -1, 0, -2, 4, 0, 2, 0, 0, 0, 0, -1, 0, 1, 0, 4, 1, 0, 0, 1, 1, 1}

	t := time.Now()
	available, _ := b.Available(1)

	//log.Println(b.Move(15, 19, 2))
	//log.Println(b.Move(15, 19, 2).UseRoll(15, 19, 2))

	log.Println(available)
	//os.Exit(0)
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

	//t4 := time.Now()
	//_ = b.ChooseDoubles(&analysis)
	//log.Println("CHOOSE DOUBLES TOOK ", time.Since(t4))

	/*

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
		}*/
}
