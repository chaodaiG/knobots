package main

import (
	"github.com/mattmoor/knobots/pkg/handler"
	"github.com/mattmoor/knobots/pkg/whitespace"
)

func main() {
	handler.Main(whitespace.New())
}
