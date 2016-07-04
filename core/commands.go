package core

import ()

var CommonCommands = []*CommandDef{}

func SendHelp(channel string, cmd string) {
	out := "**BattleBot help**\n\n"

	for _, cmd := range Commands {
		if cmd.HideFromHelp {
			continue
		}
		out += " - " + cmd.String() + "\n"
	}

	out += "\n" + VERSION

	go SendMessage(channel, out)
}
