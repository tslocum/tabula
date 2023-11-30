package main

import (
	"log"
	"time"

	"code.rocket9labs.com/tslocum/tabula"
)

func main() {
	b := tabula.NewBoard()
	b[tabula.SpaceRoll1] = 5
	b[tabula.SpaceRoll2] = 3
	b.Print()

	t := time.Now()
	available := b.Available(1)
	t2 := time.Now()
	analysis := b.Analyze(1, available)
	t3 := time.Now()

	log.Println("AVAILABLE TOOK ", t2.Sub(t))
	log.Println("ANALYSIS TOOK ", t3.Sub(t2))

	log.Println("AVAILABLE", available)
	for i, a := range analysis {
		log.Printf("%+v %+v", i, a)
	}
}
