package main

import (
	"fmt"
	"github.com/jonas747/discordgo"
	"strings"
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
			go SendMessage(m.ChannelID, m.ContentWithMentionsReplaced())
		},
	},
	&CommandDef{
		Name:        "stats",
		Aliases:     []string{"s"},
		Description: "Shows stats for a user",
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

			out := "**Stats**\n" + player.GetPrettyDiscordStats()

			player.RUnlock()
			go SendMessage(m.ChannelID, out)
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
		Description: "Responds with a bot invite link",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			go SendMessage(m.ChannelID, "**Invite link:** https://discordapp.com/oauth2/authorize?client_id=197048228099784704&scope=bot&permissions=101376")
		},
	},
	&CommandDef{
		Name:         "battle",
		Description:  "Requests a battle with another player",
		Aliases:      []string{"b"},
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := p.Args[0].DiscordUser()
			if m.Author.ID == user.ID {
				go SendMessage(m.ChannelID, "Can't fight yourself you idiot")
				return
			}

			attacker := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			defender := playerManager.GetCreatePlayer(user.ID, user.Username)

			battle := NewBattle(attacker, defender, m.ChannelID)
			if battleManager.MaybeAddBattle(battle) {
				go SendMessage(m.ChannelID, fmt.Sprintf("<@%s> Has requested a battle with <@%s>, you got 60 seconds.\nRepond with `@BattleBot accept`", m.Author.ID, user.ID))
			} else {
				go SendMessage(m.ChannelID, "Did not request battle")
			}
		},
	},
	&CommandDef{
		Name:        "accept",
		Description: "Accepts the pending battle",
		Aliases:     []string{"a"},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			if !battleManager.MaybeAcceptBattle(m.Author.ID) {
				go SendMessage(m.ChannelID, "You have no pending battles")
			}
		},
	},
	&CommandDef{
		Name:         "up",
		Description:  "Increases an attribute",
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "attribute", Type: ArgumentTypeString},
			&ArgumentDef{Name: "amount", Type: ArgumentTypeNumber},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			num := 1
			if len(p.Args) > 1 {
				num = p.Args[1].Int()
			}
			if num < 1 {
				go SendMessage(m.ChannelID, "You can't increase attributes by anything less than 1 dummy")
				return
			}

			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.Lock()
			defer player.Unlock()

			availablePoints := GetLevelFromXP(player.XP) - player.UsedAttributePoints()

			if availablePoints <= 0 {
				go SendMessage(m.ChannelID, "No available attribute points")
				return
			}

			attribute := p.Args[0].Str()

			realAttribute := ""
			switch strings.ToLower(attribute) {
			case "strength", "str":
				realAttribute = "Strength"
				player.Strength += num
			case "agility", "ag", "agi":
				realAttribute = "Agility"
				player.Agility += num
			case "stamina", "sta", "stam":
				realAttribute = "Stamina"
				player.Stamina += num
			}

			msg := fmt.Sprintf("Increased %s by %d\n\nCurrent stats:\n%s", realAttribute, num, player.GetPrettyDiscordStats())

			go SendMessage(m.ChannelID, msg)
		},
	},
}

func SendHelp(channel string) {
	out := "**BattleBot help**\n\n"

	for _, cmd := range commands {
		out += " - " + cmd.String() + "\n"
	}

	out += "\n" + VERSION

	go SendMessage(channel, out)
}
