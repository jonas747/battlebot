package main

import (
	"fmt"
	"github.com/jonas747/discordgo"
	"log"
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
			&ArgumentDef{Name: "user", Description: "User to see stats for, leave empty for yourself", Type: ArgumentTypeUser},
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
			&ArgumentDef{Name: "user", Description: "User to battle against", Type: ArgumentTypeUser},
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
			&ArgumentDef{Name: "attribute", Description: "The attribute to upgrade (one of  strength/str, agility/agi, stamina/stam)", Type: ArgumentTypeString},
			&ArgumentDef{Name: "amount", Description: "The amount to upgrade it by (1 if not specified", Type: ArgumentTypeNumber},
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

			if availablePoints < num {
				go SendMessage(m.ChannelID, "No available attribute points")
				return
			}

			attributeString := p.Args[0].Str()

			var attribute AttributeType

			switch strings.ToLower(attributeString) {
			case "strength", "str":
				attribute = AttributeStrength
			case "agility", "ag", "agi":
				attribute = AttributeAgility
			case "stamina", "sta", "stam":
				attribute = AttributeStamina
			}

			player.Attributes.Modify(attribute, num)

			msg := fmt.Sprintf("Increased %s by %d\n\nCurrent stats:\n%s", StringAttributeType(attribute), num, player.GetPrettyDiscordStats())

			go SendMessage(m.ChannelID, msg)
		},
	},
	&CommandDef{
		Name:        "inventory",
		Description: "Shows your inventory and equipment",
		Aliases:     []string{"inv", "equipment"},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			out := "**Iventory**\n"

			if len(player.Inventory) < 1 {
				out += "*dust* (you have no items)"
				SendMessage(m.ChannelID, out)
				return
			}

			for k, v := range player.Inventory {
				out += fmt.Sprintf("[%d]", k)
				itemType := GetItemTypeById(v.Id)
				if itemType == nil {
					log.Println("Encountered unknown item id", v.Id, "User:", m.Author.ID)
					out += " - Unknown!?!? (contact the jonizz)\n"
					continue
				}
				out += fmt.Sprintf(" - %s (id: %d) - %s", itemType.Name, itemType.Id, itemType.Description)
				if v.EquipmentSlot != EquipmentSlotNone {
					out += fmt.Sprintf(" (Equipped %s)", StringEquipmentSlot(v.EquipmentSlot))
				}
				out += "\n"
			}

			SendMessage(m.ChannelID, out)
		},
	},
	&CommandDef{
		Name:         "item",
		Description:  "Shows info about an item or lists all items if itemid is not specified",
		Aliases:      []string{"i", "items"},
		RequiredArgs: 0,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "itemid", Description: "If itemid is specifed shows detailed info about an item, if not shows all items", Type: ArgumentTypeNumber},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			id := -1
			if len(p.Args) > 0 && p.Args[0] != nil {
				id = p.Args[0].Int()
			}

			if id > -1 {
				itemType := GetItemTypeById(id)
				if itemType == nil {
					go SendMessage(m.ChannelID, "Unknown item")
					return
				}

				out := fmt.Sprintf("#%d - **%s**\n%s\n", itemType.Id, itemType.Name, itemType.Description)
				if len(itemType.Slots) > 0 {
					out += " - Can be equipped as: "
					for k, slot := range itemType.Slots {
						if k != 0 {
							out += ", "
						}
						out += StringEquipmentSlot(slot)
					}
					out += "\n"
				}
				pasiveEffects := itemType.Item.GetStaticAttributes()
				if len(pasiveEffects) > 0 {
					out += "\nPassive attributes:\n"
					for _, effect := range pasiveEffects {
						out += fmt.Sprintf(" - %s: %d", StringAttributeType(effect.Type), effect.Val)
					}
				}
				go SendMessage(m.ChannelID, out)
			} else {
				out := "**Items** (see `i {itemid}` for more info about an item\n"
				for k, item := range itemTypes {
					eqStr := ""

					for i, slot := range item.Slots {
						if i != 0 {
							eqStr += ", "
						}
						eqStr += StringEquipmentSlot(slot)
					}

					out += fmt.Sprintf("[%d] - %s (%s) - %s\n", k, item.Name, eqStr, item.Description)
				}
				go SendMessage(m.ChannelID, out)
			}
		},
	},
	&CommandDef{
		Name:         "equip",
		Description:  "Equips an item from your inventory",
		Aliases:      []string{"eq"},
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "inventoryslot", Description: "The inventory slot to equip (see `inventory` to list your inventory)", Type: ArgumentTypeNumber},
			&ArgumentDef{Name: "equipmentslot", Description: "Optionally sepcify a specific slot you want it in (one of head, righthand, lefthand, torso, feet, leggings)", Type: ArgumentTypeString},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			invSlot := p.Args[0].Int()

			var equipmentSlot EquipmentSlot
			if len(p.Args) > 1 && p.Args[1] != nil {
				slotString := p.Args[1].Str()
				equipmentSlot = EquipmentSlotFromString(slotString)
			}

			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.RLock()
			if invSlot >= len(player.Inventory) || invSlot < 0 {
				go SendMessage(m.ChannelID, "That inventory slot dosen't exist, see the inventory command for more info")
				player.RUnlock()
				return
			}
			itemType := GetItemTypeById(player.Inventory[invSlot].Id)

			player.RUnlock()
			if equipmentSlot == EquipmentSlotNone {
				if itemType == nil {
					SendMessage(m.ChannelID, "Unknown item at slot")
					return
				}
				equipmentSlot = itemType.Slots[0]
			}

			player.Lock()
			err := player.EquipItem(invSlot, equipmentSlot)
			if err != nil {
				go SendMessage(m.ChannelID, "Failed: "+err.Error())
			} else {
				go SendMessage(m.ChannelID, fmt.Sprintf("Equipped %s in %s", itemType.Name, StringEquipmentSlot(equipmentSlot)))
			}
			player.Unlock()
		},
	},
	&CommandDef{
		Name:         "give",
		Description:  "Gives someone an item (admin only)",
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "id", Type: ArgumentTypeNumber},
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := m.Author
			if len(p.Args) > 1 && p.Args[1] != nil {
				user = p.Args[1].DiscordUser()
			}

			itemType := GetItemTypeById(p.Args[0].Int())

			if itemType == nil {
				go SendMessage(m.ChannelID, "Unknown item")
				return
			}

			player := playerManager.GetCreatePlayer(user.ID, user.Username)
			player.Lock()
			player.Inventory = append(player.Inventory, &PlayerItem{Id: itemType.Id})

			go SendMessage(m.ChannelID, fmt.Sprintf("Gave **%s** %s (#%d)", player.Name, itemType.Name, itemType.Id))
			player.Unlock()
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
