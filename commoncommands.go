package main

import (
	"github.com/jonas747/discordgo"
)

var CommonCommands = []*CommandDef{
	&CommandDef{
		Name:        "help",
		Description: "Prints help info",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			SendHelp(m.ChannelID)
		},
	},
	&CommandDef{
		Name:        "echo",
		Description: "Make me say stuff ;)",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			dgo.ChannelMessageSend(m.ChannelID, m.ContentWithMentionsReplaced())
		},
	},
}

func SendHelp(channel string) {
	out := ""

	for _, cmd := range commands {
		out += " - " + cmd.String() + "\n"
	}

	out += "\n" + VERSION
}
