package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/battlebot/core"
)

var BattleCommands = []*core.CommandDef{
	&core.CommandDef{
		Name:         "battle",
		Description:  "Requests a battle with another player",
		Aliases:      []string{"b"},
		RequiredArgs: 1,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "user", Description: "User to battle against", Type: core.ArgumentTypeUser},
			&core.ArgumentDef{Name: "money", Description: "Money to battle over, both of you put in this amountand winner gets all", Type: core.ArgumentTypeNumber},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			user := p.Args[0].DiscordUser()
			if m.Author.ID == user.ID {
				go core.SendMessage(m.ChannelID, "Can't fight yourself you idiot")
				return
			}

			money := 1
			if len(p.Args) > 1 && p.Args[1] != nil {
				money = p.Args[1].Int()
			}

			attacker := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			defender := core.Players.GetCreatePlayer(user.ID, user.Username)

			noMoneyMsg := ""
			attacker.RLock()
			if attacker.Money < money {
				noMoneyMsg = "You"
			}
			attacker.RUnlock()
			defender.RLock()
			if noMoneyMsg == "" && defender.Money < money {
				noMoneyMsg = defender.Name
			}
			defender.RUnlock()

			if noMoneyMsg != "" {
				go core.SendMessage(m.ChannelID, noMoneyMsg+" do not have enough money to battle :'( Battle some monsters first?")
				return
			}

			battle := core.NewBattle(attacker, defender, money, m.ChannelID)
			if core.Battles.MaybeAddBattle(battle) {
				go core.SendMessage(m.ChannelID, fmt.Sprintf("<@%s> Has requested a battle with <@%s> for %d$, you got 60 seconds.\nRepond with `@BattleBot accept`", m.Author.ID, user.ID, money))
			} else {
				go core.SendMessage(m.ChannelID, "Did not request battle")
			}
		},
	},
	&core.CommandDef{
		Name:        "battlemonster",
		Aliases:     []string{"bm"},
		Description: "Battle a random monster at your level",
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)

			monster := core.GetMonster(core.GetLevelFromXP(player.XP))

			battle := core.NewBattle(player, monster.Player, monster.Money, m.ChannelID)
			battle.IsMonster = true

			battle.Battle()
		},
	},
	&core.CommandDef{
		Name:        "accept",
		Description: "Accepts the pending battle",
		Aliases:     []string{"a"},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			if !core.Battles.MaybeAcceptBattle(m.Author.ID) {
				go core.SendMessage(m.ChannelID, "You have no pending battles")
			}
		},
	},
}
