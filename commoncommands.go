package main

import (
	"fmt"
	"github.com/jonas747/discordgo"
	"log"
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
		Name:         "stats",
		Description:  "Shows stats for a user",
		OptionalArgs: true,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := m.Author
			if len(p.Args) > 0 && p.Args[0] != nil {
				user = p.Args[0].DiscordUser()
			}

			player := playerManager.GetCreatePlayer(user.ID, user.Username)
			player.RLock()

			level := GetLevelFromXP(player.XP)
			next := GetXPForLevel(level + 1)
			diff := next - player.XP

			out := fmt.Sprintf("**%s**\n - Level: %d\n - XP: %d\n - Wins: %d\n - Losses: %d",
				player.Name, GetLevelFromXP(player.XP), diff, player.Wins, player.Losses)

			player.RUnlock()
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
	&CommandDef{
		Name:        "battle",
		Description: "Requests a battle with another player",
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := p.Args[0].DiscordUser()
			if m.Author.ID == user.ID {
				dgo.ChannelMessageSend(m.ChannelID, "Can't fight yourself you idiot")
				return
			}

			attacker := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			defender := playerManager.GetCreatePlayer(user.ID, user.Username)

			battle := NewBattle(attacker, defender, m.ChannelID)
			log.Println("Calling maybeaddbattle")
			if battleManager.MaybeAddBattle(battle) {
				dgo.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> Has requested a battle with <@%s>, you got 60 seconds.\nRepond with `@BattleBot accept`", m.Author.ID, user.ID))
			} else {
				dgo.ChannelMessageSend(m.ChannelID, "Did not request battle")
			}
			log.Println("after maybeaddbattle")
		},
	},
	&CommandDef{
		Name:        "accept",
		Description: "Accepts the pending battle",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			if !battleManager.MaybeAcceptBattle(m.Author.ID) {
				dgo.ChannelMessageSend(m.ChannelID, "You have no pending battles")
			}
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
