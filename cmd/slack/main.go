package main

import (
	"github.com/mattmoor/knobots/pkg/handler"
	"github.com/mattmoor/knobots/pkg/handler/slack"
)

func main() {
	handler.Main(slack.New)
}
