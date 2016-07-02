package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
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
			&ArgumentDef{Name: "money", Description: "Money to battle over, both of you put in this amountand winner gets all", Type: ArgumentTypeNumber},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			user := p.Args[0].DiscordUser()
			if m.Author.ID == user.ID {
				go SendMessage(m.ChannelID, "Can't fight yourself you idiot")
				return
			}

			money := 1
			if len(p.Args) > 1 && p.Args[1] != nil {
				money = p.Args[1].Int()
			}

			attacker := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			defender := playerManager.GetCreatePlayer(user.ID, user.Username)

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
				go SendMessage(m.ChannelID, noMoneyMsg+" Does not have enough money to battle :'( Battle some monsters first?")
				return
			}

			battle := NewBattle(attacker, defender, money, m.ChannelID)
			if battleManager.MaybeAddBattle(battle) {
				go SendMessage(m.ChannelID, fmt.Sprintf("<@%s> Has requested a battle with <@%s> for %d$, you got 60 seconds.\nRepond with `@BattleBot accept`", m.Author.ID, user.ID, money))
			} else {
				go SendMessage(m.ChannelID, "Did not request battle")
			}
		},
	},
	&CommandDef{
		Name:        "battlemonster",
		Aliases:     []string{"bm"},
		Description: "Battle a random monster at your level",
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)

			monster := GetMonster(GetLevelFromXP(player.XP))

			battle := NewBattle(player, monster.Player, monster.Money, m.ChannelID)
			battle.IsMonster = true

			battle.Battle()
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

				out := fmt.Sprintf("#%d - **%s** - $%d\n%s\n", itemType.Id, itemType.Name, itemType.Cost, itemType.Description)
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
				out := "**Items** (see `i {itemid}` for more info about an item)\n"
				for k, item := range itemTypes {
					eqStr := ""

					for i, slot := range item.Slots {
						if i != 0 {
							eqStr += ", "
						}
						eqStr += StringEquipmentSlot(slot)
					}

					out += fmt.Sprintf("[%d] - %s (%s) - %d$ - %s\n", k, item.Name, eqStr, item.Cost, item.Description)
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
		Name:         "unequip",
		Description:  "Unequips an item from your inventory",
		Aliases:      []string{"ueq", "ue", "deq", "dequip"},
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "Inventory or Equipment Slot", Description: "Either a inventory slot (number)  or Equipment slot (one of head, righthand, lefthand, torso, feet, leggings)", Type: ArgumentTypeString},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			val := p.Args[0].Str()

			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)

			var itemType *ItemType

			// Check if its a number
			num, err := strconv.ParseInt(val, 10, 32)

			player.Lock()
			if err != nil {
				// An inventory slot
				invSlot := int(num)
				if invSlot >= len(player.Inventory) || invSlot < 0 {
					go SendMessage(m.ChannelID, "That inventory slot dosen't exist, see the inventory command for more info")
					player.Unlock()
					return
				}
				itemType = GetItemTypeById(player.Inventory[invSlot].Id)

				player.Inventory[invSlot].EquipmentSlot = EquipmentSlotNone
			} else {
				// An equipment slot
				equipmentSlot := EquipmentSlotFromString(val)
				if equipmentSlot == EquipmentSlotNone {
					go SendMessage(m.ChannelID, "Unknown equipment slot")
					player.Unlock()
					return
				}
				for _, v := range player.Inventory {
					if v.EquipmentSlot == equipmentSlot {
						v.EquipmentSlot = EquipmentSlotNone
						itemType = GetItemTypeById(v.Id)
						break
					}
				}
			}
			player.Unlock()

			if itemType == nil {
				go SendMessage(m.ChannelID, "Didn't strip...")
			} else {
				go SendMessage(m.ChannelID, fmt.Sprintf("Unequipped %s", itemType.Name))
			}
		},
	},
	&CommandDef{
		Name:         "create",
		Description:  "Creates an item for someone (admin only)",
		RequiredArgs: 1,
		HideFromHelp: true,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "id", Type: ArgumentTypeNumber},
			&ArgumentDef{Name: "user", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			if m.Author.ID != "105487308693757952" {
				go SendMessage(m.ChannelID, "You're not an admin >:O")
				return
			}

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
	&CommandDef{
		Name:         "buy",
		Description:  "Buys an item",
		RequiredArgs: 1,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "id", Description: "Item you want to buy (see `items/i` for item id's)", Type: ArgumentTypeNumber},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			itemType := GetItemTypeById(p.Args[0].Int())

			if itemType == nil {
				go SendMessage(m.ChannelID, "Unknown item")
				return
			}

			player := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.Lock()

			if player.Money >= itemType.Cost {
				player.Inventory = append(player.Inventory, &PlayerItem{Id: itemType.Id})
				originalMoney := player.Money
				player.Money -= itemType.Cost
				go SendMessage(m.ChannelID, fmt.Sprintf("**%s** Purchased: %s (#%d) for %d$ (%d$ -> %d$)", player.Name, itemType.Name, itemType.Id, itemType.Cost, originalMoney, player.Money))
			} else {
				go SendMessage(m.ChannelID, "Can't afford that item")
			}

			player.Unlock()
		},
	},
	&CommandDef{
		Name:         "give",
		Description:  "Give someone an item from your inventory",
		RequiredArgs: 2,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "Inventory slot", Description: "Inventoryslot of the item you're giving away", Type: ArgumentTypeNumber},
			&ArgumentDef{Name: "Receiver", Description: "Person who's receiving the item", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			sender := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			receiverUser := p.Args[1].DiscordUser()
			receiver := playerManager.GetCreatePlayer(receiverUser.ID, receiverUser.Username)

			if sender.Id == receiver.Id {
				go SendMessage(m.ChannelID, "Can't give yourself an item...")
				return
			}

			slotIndex := p.Args[0].Int()

			sender.Lock()

			if slotIndex < 0 || slotIndex >= len(sender.Inventory) {
				go SendMessage(m.ChannelID, "There's no item in that slot")
				sender.Unlock()
				return
			}

			item := sender.Inventory[slotIndex]
			sender.Inventory = append(sender.Inventory[:slotIndex], sender.Inventory[slotIndex+1:]...)
			sender.Unlock()

			item.EquipmentSlot = EquipmentSlotNone
			itemType := GetItemTypeById(item.Id)

			receiver.Lock()
			receiver.Inventory = append(receiver.Inventory, item)
			receiver.Unlock()

			go SendMessage(m.ChannelID, fmt.Sprintf("**%s** Gave **%s** %s(#%d)", sender.Name, receiver.Name, itemType.Name, itemType.Id))
		},
	},
	&CommandDef{
		Name:         "givemoney",
		Aliases:      []string{"givem", "gm"},
		Description:  "Give someone money",
		RequiredArgs: 2,
		Arguments: []*ArgumentDef{
			&ArgumentDef{Name: "Money", Description: "Money you want to give", Type: ArgumentTypeNumber},
			&ArgumentDef{Name: "Receiver", Description: "Person who's receiving the item", Type: ArgumentTypeUser},
		},
		RunFunc: func(p *ParsedCommand, m *discordgo.MessageCreate) {
			amount := p.Args[0].Int()

			sender := playerManager.GetCreatePlayer(m.Author.ID, m.Author.Username)
			receiverUser := p.Args[1].DiscordUser()
			receiver := playerManager.GetCreatePlayer(receiverUser.ID, receiverUser.Username)

			sender.Lock()
			if sender.Money < amount {
				go SendMessage(m.ChannelID, "Not enough money to send")
				sender.Unlock()
				return
			}

			sender.Money -= amount
			sender.Unlock()

			receiver.Lock()
			receiver.Money += amount
			go SendMessage(m.ChannelID, fmt.Sprintf("**%s** Gave **%s** %s$ (%d$ -> %d$)", sender.Name, receiver.Name, amount, receiver.Money-amount, receiver.Money))
			receiver.Unlock()
		},
	},
}

func SendHelp(channel string) {
	out := "**BattleBot help**\n\n"

	for _, cmd := range commands {
		if cmd.HideFromHelp {
			continue
		}
		out += " - " + cmd.String() + "\n"
	}

	out += "\n" + VERSION

	go SendMessage(channel, out)
}
