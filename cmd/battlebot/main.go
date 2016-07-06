package main

import (
	_ "github.com/jonas747/battlebot/commands"
	"github.com/jonas747/battlebot/core"
	"github.com/jonas747/battlebot/items"
)

func main() {
	items.RegisterGenericItems()
	core.Run()
}
