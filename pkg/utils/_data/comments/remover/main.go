package main

import (
	"fmt"
)

type logger struct {
	// prefix of each log
	prefix string
}

// Println method log the given message.
func (l *logger) Println(msg ...string) {
	fmt.Printf("%v %v\n", l.prefix, msg)
}

// Main function of our program
func main() {
	// our logger
	log := &logger{}
	// log "Hello world"
	log.Println("Hello world")
}
