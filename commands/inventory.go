package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/battlebot/core"
	"log"
	"strconv"
)

var InventoryCommands = []*core.CommandDef{
	&core.CommandDef{
		Name:        "inventory",
		Description: "Shows your inventory and equipment",
		Aliases:     []string{"inv", "equipment"},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			out := "**Iventory**\n"

			if len(player.Inventory) < 1 {
				out += "*dust* (you have no items)"
				go core.SendMessage(m.ChannelID, out)
				return
			}

			for k, v := range player.Inventory {
				out += fmt.Sprintf("[%d]", k)
				itemType := core.GetItemTypeById(v.Id)
				if itemType == nil {
					log.Println("Encountered unknown item id", v.Id, "User:", m.Author.ID)
					out += " - Unknown!?!? (contact the jonizz)\n"
					continue
				}
				out += fmt.Sprintf(" - %s (id: %d) - %s", itemType.Name, itemType.Id, itemType.Description)
				if v.EquipmentSlot != core.EquipmentSlotNone {
					out += fmt.Sprintf(" (Equipped %s)", v.EquipmentSlot.String())
				}
				out += "\n"
			}

			go core.SendMessage(m.ChannelID, out)
		},
	},
	&core.CommandDef{
		Name:         "item",
		Description:  "Shows info about an item or lists all items if itemid is not specified",
		Aliases:      []string{"i", "items"},
		RequiredArgs: 0,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "itemid", Description: "If itemid is specifed shows detailed info about an item, if not shows all items", Type: core.ArgumentTypeNumber},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			id := -1
			if len(p.Args) > 0 && p.Args[0] != nil {
				id = p.Args[0].Int()
			}

			if id > -1 {
				itemType := core.GetItemTypeById(id)
				if itemType == nil {
					go core.SendMessage(m.ChannelID, "Unknown item")
					return
				}

				out := fmt.Sprintf("#%d - **%s** - $%d\n%s\n", itemType.Id, itemType.Name, itemType.Cost, itemType.Description)
				if len(itemType.Slots) > 0 {
					out += " - Can be equipped as: "
					for k, slot := range itemType.Slots {
						if k != 0 {
							out += ", "
						}
						out += slot.String()
					}
					out += "\n"
				}
				pasiveEffects := itemType.Item.GetStaticAttributes()
				if len(pasiveEffects) > 0 {
					out += "\nPassive attributes:\n"
					for _, effect := range pasiveEffects {
						out += fmt.Sprintf(" - %s: %.2f", effect.Type.String(), effect.Amount)
					}
				}
				go core.SendMessage(m.ChannelID, out)
			} else {
				out := "**Items** (see `i {itemid}` for more info about an item)\n"
				for k, item := range core.ItemTypes {
					eqStr := ""

					for i, slot := range item.Slots {
						if i != 0 {
							eqStr += ", "
						}
						eqStr += slot.String()
					}

					out += fmt.Sprintf("[%d] - %s (%s) - %d$ - %s\n", k, item.Name, eqStr, item.Cost, item.Description)
				}
				go core.SendMessage(m.ChannelID, out)
			}
		},
	},
	&core.CommandDef{
		Name:         "equip",
		Description:  "Equips an item from your inventory",
		Aliases:      []string{"eq"},
		RequiredArgs: 1,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "inventoryslot", Description: "The inventory slot to equip (see `inventory` to list your inventory)", Type: core.ArgumentTypeNumber},
			&core.ArgumentDef{Name: "core.equipmentslot", Description: "Optionally sepcify a specific slot you want it in (one of head, righthand, lefthand, torso, feet, leggings)", Type: core.ArgumentTypeString},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			invSlot := p.Args[0].Int()

			var equipmentSlot core.EquipmentSlot
			if len(p.Args) > 1 && p.Args[1] != nil {
				slotString := p.Args[1].Str()
				equipmentSlot = core.EquipmentSlotFromString(slotString)
			}

			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.RLock()
			if invSlot >= len(player.Inventory) || invSlot < 0 {
				go core.SendMessage(m.ChannelID, "That inventory slot dosen't exist, see the inventory command for more info")
				player.RUnlock()
				return
			}
			itemType := core.GetItemTypeById(player.Inventory[invSlot].Id)

			player.RUnlock()
			if equipmentSlot == core.EquipmentSlotNone {
				if itemType == nil {
					core.SendMessage(m.ChannelID, "Unknown item at slot")
					return
				}
				equipmentSlot = itemType.Slots[0]
			}

			player.Lock()
			err := player.EquipItem(invSlot, equipmentSlot)
			if err != nil {
				go core.SendMessage(m.ChannelID, "Failed: "+err.Error())
			} else {
				go core.SendMessage(m.ChannelID, fmt.Sprintf("Equipped %s in %s", itemType.Name, equipmentSlot.String()))
			}
			player.Unlock()
		},
	},
	&core.CommandDef{
		Name:         "unequip",
		Description:  "Unequips an item from your inventory",
		Aliases:      []string{"ueq", "ue", "deq", "dequip"},
		RequiredArgs: 1,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "Inventory or Equipment Slot", Description: "Either a inventory slot (number)  or Equipment slot (one of head, righthand, lefthand, torso, feet, leggings)", Type: core.ArgumentTypeString},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			val := p.Args[0].Str()

			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)

			var itemType *core.ItemType

			// Check if its a number
			num, err := strconv.ParseInt(val, 10, 32)

			player.Lock()
			if err == nil {
				// An inventory slot
				invSlot := int(num)
				if invSlot >= len(player.Inventory) || invSlot < 0 {
					go core.SendMessage(m.ChannelID, "That inventory slot dosen't exist, see the inventory command for more info")
					player.Unlock()
					return
				}
				itemType = core.GetItemTypeById(player.Inventory[invSlot].Id)

				player.Inventory[invSlot].EquipmentSlot = core.EquipmentSlotNone
			} else {
				// An equipment slot
				equipmentSlot := core.EquipmentSlotFromString(val)
				if equipmentSlot == core.EquipmentSlotNone {
					go core.SendMessage(m.ChannelID, "Unknown equipment slot")
					player.Unlock()
					return
				}
				for _, v := range player.Inventory {
					if v.EquipmentSlot == equipmentSlot {
						v.EquipmentSlot = core.EquipmentSlotNone
						itemType = core.GetItemTypeById(v.Id)
						break
					}
				}
			}
			player.Unlock()

			if itemType == nil {
				go core.SendMessage(m.ChannelID, "Didn't strip...")
			} else {
				go core.SendMessage(m.ChannelID, fmt.Sprintf("Unequipped %s", itemType.Name))
			}
		},
	},
	&core.CommandDef{
		Name:         "create",
		Description:  "Creates an item for someone (admin only)",
		RequiredArgs: 1,
		HideFromHelp: true,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "id", Type: core.ArgumentTypeNumber},
			&core.ArgumentDef{Name: "user", Type: core.ArgumentTypeUser},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			if m.Author.ID != "105487308693757952" {
				go core.SendMessage(m.ChannelID, "You're not an admin >:O")
				return
			}

			user := m.Author
			if len(p.Args) > 1 && p.Args[1] != nil {
				user = p.Args[1].DiscordUser()
			}

			itemType := core.GetItemTypeById(p.Args[0].Int())

			if itemType == nil {
				go core.SendMessage(m.ChannelID, "Unknown item")
				return
			}

			player := core.Players.GetCreatePlayer(user.ID, user.Username)
			player.Lock()
			player.Inventory = append(player.Inventory, &core.PlayerItem{Id: itemType.Id})

			go core.SendMessage(m.ChannelID, fmt.Sprintf("Gave **%s** %s (#%d)", player.Name, itemType.Name, itemType.Id))
			player.Unlock()
		},
	},
	&core.CommandDef{
		Name:         "give",
		Description:  "Give someone an item from your inventory",
		RequiredArgs: 2,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "Inventory slot", Description: "Inventoryslot of the item you're giving away", Type: core.ArgumentTypeNumber},
			&core.ArgumentDef{Name: "Receiver", Description: "Person who's receiving the item", Type: core.ArgumentTypeUser},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			sender := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			receiverUser := p.Args[1].DiscordUser()
			receiver := core.Players.GetCreatePlayer(receiverUser.ID, receiverUser.Username)

			if sender.Id == receiver.Id {
				go core.SendMessage(m.ChannelID, "Can't give yourself an item...")
				return
			}

			slotIndex := p.Args[0].Int()

			sender.Lock()

			if slotIndex < 0 || slotIndex >= len(sender.Inventory) {
				go core.SendMessage(m.ChannelID, "There's no item in that slot")
				sender.Unlock()
				return
			}

			item := sender.Inventory[slotIndex]
			sender.Inventory = append(sender.Inventory[:slotIndex], sender.Inventory[slotIndex+1:]...)
			sender.Unlock()

			item.EquipmentSlot = core.EquipmentSlotNone
			itemType := core.GetItemTypeById(item.Id)

			receiver.Lock()
			receiver.Inventory = append(receiver.Inventory, item)
			receiver.Unlock()

			go core.SendMessage(m.ChannelID, fmt.Sprintf("**%s** Gave **%s** %s(#%d)", sender.Name, receiver.Name, itemType.Name, itemType.Id))
		},
	},
	&core.CommandDef{
		Name:         "buy",
		Description:  "Buys an item",
		RequiredArgs: 1,
		Arguments: []*core.ArgumentDef{
			&core.ArgumentDef{Name: "id", Description: "Item you want to buy (see `items/i` for item id's)", Type: core.ArgumentTypeNumber},
		},
		RunFunc: func(p *core.ParsedCommand, m *discordgo.MessageCreate) {
			itemType := core.GetItemTypeById(p.Args[0].Int())

			if itemType == nil {
				go core.SendMessage(m.ChannelID, "Unknown item")
				return
			}

			player := core.Players.GetCreatePlayer(m.Author.ID, m.Author.Username)
			player.Lock()

			if player.Money >= itemType.Cost {
				player.Inventory = append(player.Inventory, &core.PlayerItem{Id: itemType.Id})
				originalMoney := player.Money
				player.Money -= itemType.Cost
				go core.SendMessage(m.ChannelID, fmt.Sprintf("**%s** Purchased: %s (#%d) for %d$ (%d$ -> %d$)", player.Name, itemType.Name, itemType.Id, itemType.Cost, originalMoney, player.Money))
			} else {
				go core.SendMessage(m.ChannelID, "Can't afford that item")
			}

			player.Unlock()
		},
	},
}
