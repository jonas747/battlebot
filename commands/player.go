package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/battlebot/core"
	"strings"
)

var PlayerCommands = []*core.CommandDef{
	&core.CommandDef{
		Name:        "stats",
		Aliases:     []string{"s"},
		Description: "Shows stats for a user",
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "User", Description: "User to see stats for, leave empty for yourself", Type: core.ArgumentTypeUser},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			var player *core.Player
			if len(p.Args) > 0 && p.Args[0] != nil {
				player = p.Args[0].GetCreatePlayer()
			} else {
				player = core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			}

			player.RLock()

			out := "**Stats**\n" + player.GetPrettyDiscordStats()

			player.RUnlock()
			go core.SendMessage(m.ChannelID, out)
		},
	},
	&core.CommandDef{
		Name:         "up",
		Description:  "Increases an attribute",
		RequiredArgs: 1,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "attribute", Description: "The attribute to upgrade (one of  strength/str, agility/agi, stamina/stam)", Type: core.ArgumentTypeString},
			&core.ArgumentDef{Name: "amount", Description: "The amount to upgrade it by (1 if not specified", Type: core.ArgumentTypeNumber},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			num := 1
			if len(p.Args) > 1 && p.Args[1] != nil {
				num = p.Args[1].Int()
			}
			if num < 1 {
				go core.SendMessage(m.ChannelID, "You can't increase attributes by anything less than 1 dummy")
				return
			}

			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.Lock()
			defer player.Unlock()

			availablePoints := core.GetLevelFromXP(player.XP) - player.UsedAttributePoints()

			if availablePoints < num {
				go core.SendMessage(m.ChannelID, "No available attribute points")
				return
			}

			attributeString := p.Args[0].Str()

			var attribute core.AttributeType

			switch strings.ToLower(attributeString) {
			case "strength", "str":
				attribute = core.AttributeStrength
			case "agility", "ag", "agi":
				attribute = core.AttributeAgility
			case "stamina", "sta", "stam":
				attribute = core.AttributeStamina
			}

			player.Attributes.Modify(attribute, num)

			msg := fmt.Sprintf("Increased %s by %d\n\nCurrent stats:\n%s", attribute.String(), num, player.GetPrettyDiscordStats())

			go core.SendMessage(m.ChannelID, msg)
		},
	},
	&core.CommandDef{
		Name:         "givemoney",
		Aliases:      []string{"givem", "gm"},
		Description:  "Give someone money",
		RequiredArgs: 2,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "Money", Description: "Money you want to give", Type: core.ArgumentTypeNumber},
			&core.ArgumentDef{Name: "Receiver", Description: "Person who's receiving the item", Type: core.ArgumentTypeUser},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			amount := p.Args[0].Int()

			sender := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			receiverUser := p.Args[1].DiscordUser()
			receiver := core.Players.GetCreatePlayer(receiverUser.ID, receiverUser.Username)

			sender.Lock()
			if sender.Money < amount {
				go core.SendMessage(m.ChannelID, "Not enough money to send")
				sender.Unlock()
				return
			}

			sender.Money -= amount
			sender.Unlock()

			receiver.Lock()
			receiver.Money += amount
			go core.SendMessage(m.ChannelID, fmt.Sprintf("**%s** Gave **%s** %s$ (%d$ -> %d$)", sender.Name, receiver.Name, amount, receiver.Money-amount, receiver.Money))
			receiver.Unlock()
		},
	},
}
