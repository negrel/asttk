package main

import (
	"fmt"
)

type logger struct {
	prefix string
}

func (l *logger) Println(msg ...string) {
	fmt.Printf("%v %v\n", l.prefix, msg)
}

func main() {

	log := &logger{}

	log.Println("Hello world")
}
