package main

import (
	_ "github.com/jonas747/battlebot/commands"
	"github.com/jonas747/battlebot/core"
	_ "github.com/jonas747/battlebot/items"
)

func main() {
	core.Run()
}
