package items

import (
	"github.com/jonas747/battlebot/core"
)

func init() {
	core.RegisterItems(ItemTypes...)
}

var ItemTypes = []*core.ItemType{
	&core.ItemType{
		Id:          0,
		Name:        "Poor mans boots",
		Description: "Simple boots that increase your stamina by 1",
		Slots:       []core.EquipmentSlot{core.EquipmentSlotFeet},
		Cost:        1,
		Item: &core.SimpleItem{
			Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeStamina, Amount: 1}},
		},
	},
	&core.ItemType{
		Id:          1,
		Name:        "Watermelon",
		Description: "Increases your stamina by 5",
		Cost:        10,
		Slots:       []core.EquipmentSlot{core.EquipmentSlotHead},
		Item: &core.SimpleItem{
			Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeStamina, Amount: 5}},
		},
	},
	&core.ItemType{
		Id:          2,
		Name:        "Knife",
		Description: "Increases your strength by 5",
		Cost:        10,
		Slots:       []core.EquipmentSlot{core.EquipmentSlotRightHand, core.EquipmentSlotLeftHand},
		Item: &core.SimpleItem{
			Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeStrength, Amount: 5}},
		},
	},
	&core.ItemType{
		Id:          3,
		Name:        "Speeeed Bootz",
		Description: "Increases your agility by 5",
		Cost:        10,
		Slots:       []core.EquipmentSlot{core.EquipmentSlotFeet},
		Item: &core.SimpleItem{
			Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeAgility, Amount: 5}},
		},
	},
	&core.ItemType{
		Id:          4,
		Name:        "Holy Torso",
		Description: "Every turn you heal 2 damage",
		Slots:       []core.EquipmentSlot{core.EquipmentSlotTorso},
		Cost:        10,
		Item: &core.ItemEffectEmitter{
			SimpleItem: core.SimpleItem{
				Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeStamina, Amount: 2}},
			},
			Triggers: []*core.EffectTrigger{
				&core.EffectTrigger{
					Target:  core.TargetSelf,
					Trigger: core.EffectTriggerTurn,
					Apply: func(sender *core.BattlePlayer, receiver *core.BattlePlayer, battle *core.Battle) {
						if battle.CurTurn%2 == 0 {
							battle.DealDamage(sender, receiver, -2, "Holy Torso")
						}
					},
				},
			},
		},
	},
	&core.ItemType{
		Id:          5,
		Name:        "Basic Wand",
		Description: "On attacks there's a 20% chance the wand will shoot flowers and heal your opponent for 10 damage",
		Slots:       []core.EquipmentSlot{core.EquipmentSlotLeftHand, core.EquipmentSlotRightHand},
		Cost:        10,
		Item: &core.ItemEffectEmitter{
			SimpleItem: core.SimpleItem{
				Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeStrength, Amount: 5}},
			},
			Triggers: []*core.EffectTrigger{
				&core.EffectTrigger{
					Chance:  0.2,
					Target:  core.TargetOpponent,
					Trigger: core.EffectTriggerAttack,
					Apply: func(sender *core.BattlePlayer, receiver *core.BattlePlayer, battle *core.Battle) {
						battle.DealDamage(sender, receiver, -10, "Flowers")
					},
				},
			},
		},
	},
	&core.ItemType{
		Id:          6,
		Name:        "Paper airplane",
		Description: "20%% Chance you stun the enemy for 4 turns on attack",
		Slots:       []core.EquipmentSlot{core.EquipmentSlotRightHand, core.EquipmentSlotLeftHand},
		Cost:        20,
		Item: &core.ItemEffectEmitter{
			Triggers: []*core.EffectTrigger{
				&core.EffectTrigger{
					Chance:  0.2,
					Target:  core.TargetOpponent,
					Trigger: core.EffectTriggerAttack,
					Apply: func(sender *core.BattlePlayer, receiver *core.BattlePlayer, battle *core.Battle) {
						battle.Stun(sender, receiver, 4, "Paper airplane")
					},
				},
			},
		},
	},
	&core.ItemType{
		Id:          7,
		Name:        "War Paint",
		Description: "Increases your foucs, your miss chance is decreased by 10%",
		Slots:       []core.EquipmentSlot{core.EquipmentSlotRightHand, core.EquipmentSlotLeftHand, core.EquipmentSlotHead, core.EquipmentSlotFeet, core.EquipmentSlotLeggings, core.EquipmentSlotTorso},
		Cost:        10,
		Item: &core.SimpleItem{
			Attributes: []core.ItemAttribute{core.ItemAttribute{Type: core.ItemAttributeMissChance, Amount: -10}},
		},
	},
}
