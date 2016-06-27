package main

import (
	"fmt"
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
	&CommandDef{
		Name:        "whois",
		Description: "WHO IS THIS",
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := p.Args[0].DiscordUser()
			out := fmt.Sprintf("**%s#%s**", user.Username, user.Discriminator)
			dgo.ChannelMessageSend(m.ChannelID, out)
		},
	},
	&CommandDef{
		Name:        "panic",
		Description: "This will panic",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			var u *CommandDef
			u.Name = "wont happen"
		},
	},
	&CommandDef{
		Name:        "invite",
		Description: "Responds with discord invite",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			dgo.ChannelMessageSend(m.ChannelID, "**Invite link:** https://discordapp.com/oauth2/authorize?client_id=197048228099784704&scope=bot&permissions=101376")
		},
	},
}

func SendHelp(channel string) {
	out := "**BattleBot help**\n\n"

	for _, cmd := range commands {
		out += " - " + cmd.String() + "\n"
	}

	out += "\n" + VERSION

	dgo.ChannelMessageSend(channel, out)
}
