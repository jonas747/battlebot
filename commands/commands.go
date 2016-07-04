package commands

import (
	"github.com/jonas747/battlebot/core"
)

func init() {
	core.RegisterCommands(MiscCommands...)
	core.RegisterCommands(BattleCommands...)
	core.RegisterCommands(InventoryCommands...)
	core.RegisterCommands(PlayerCommands...)
}
