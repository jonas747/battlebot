package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/battlebot/core"
)

var MiscCommands = []*core.CommandDef{
	&core.CommandDef{
		Name:        "help",
		Description: "Prints help info",
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			core.SendHelp(m.ChannelID, "")
		},
	},

	&core.CommandDef{
		Name:        "invite",
		Description: "Responds with a bot invite link",
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			go core.SendMessage(m.ChannelID, "**Invite link:** https://discordapp.com/oauth2/authorize?client_id=197048228099784704&scope=bot&permissions=101376")
		},
	},
}
