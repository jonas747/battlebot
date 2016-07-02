package main

var itemTypes = []*ItemType{
	&ItemType{
		Id:          0,
		Name:        "Poor mans boots",
		Description: "Simple boots that increase your stamina by 1",
		Slots:       []EquipmentSlot{EquipmentSlotFeet},
		Cost:        1,
		Item: &SimpleItem{
			Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeStamina, Amount: 1}},
		},
	},
	&ItemType{
		Id:          1,
		Name:        "Watermelon",
		Description: "Increases your stamina by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotHead},
		Item: &SimpleItem{
			Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeStamina, Amount: 5}},
		},
	},
	&ItemType{
		Id:          2,
		Name:        "Knife",
		Description: "Increases your strength by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotRightHand, EquipmentSlotLeftHand},
		Item: &SimpleItem{
			Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeStrength, Amount: 5}},
		},
	},
	&ItemType{
		Id:          3,
		Name:        "Speeeed Bootz",
		Description: "Increases your agility by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotFeet},
		Item: &SimpleItem{
			Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeAgility, Amount: 5}},
		},
	},
	&ItemType{
		Id:          4,
		Name:        "Holy Torso",
		Description: "Every turn you heal 2 damage",
		Slots:       []EquipmentSlot{EquipmentSlotTorso},
		Cost:        10,
		Item: &ItemEffectEmitter{
			SimpleItem: SimpleItem{
				Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeStamina, Amount: 2}},
			},
			Triggers: []*EffectTrigger{
				&EffectTrigger{
					Target:  TargetSelf,
					Trigger: EffectTriggerTurn,
					Apply: func(sender *BattlePlayer, receiver *BattlePlayer, battle *Battle) {
						if battle.CurTurn%2 == 0 {
							battle.DealDamage(sender, receiver, -2, "Holy Torso")
						}
					},
				},
			},
		},
	},
	&ItemType{
		Id:          5,
		Name:        "Basic Wand",
		Description: "On attacks there's a 20% chance the wand will shoot flowers and heal your opponent for 10 damage",
		Slots:       []EquipmentSlot{EquipmentSlotLeftHand, EquipmentSlotRightHand},
		Cost:        10,
		Item: &ItemEffectEmitter{
			SimpleItem: SimpleItem{
				Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeStrength, Amount: 5}},
			},
			Triggers: []*EffectTrigger{
				&EffectTrigger{
					Chance:  0.2,
					Target:  TargetOpponent,
					Trigger: EffectTriggerAttack,
					Apply: func(sender *BattlePlayer, receiver *BattlePlayer, battle *Battle) {
						battle.DealDamage(sender, receiver, -10, "Flowers")
					},
				},
			},
		},
	},
	&ItemType{
		Id:          6,
		Name:        "Paper airplane",
		Description: "20%% Chance you stun the enemy for 4 turns on attack",
		Slots:       []EquipmentSlot{EquipmentSlotRightHand, EquipmentSlotLeftHand},
		Cost:        20,
		Item: &ItemEffectEmitter{
			Triggers: []*EffectTrigger{
				&EffectTrigger{
					Chance:  0.2,
					Target:  TargetOpponent,
					Trigger: EffectTriggerAttack,
					Apply: func(sender *BattlePlayer, receiver *BattlePlayer, battle *Battle) {
						battle.Stun(sender, receiver, 4, "Paper airplane")
					},
				},
			},
		},
	},
	&ItemType{
		Id:          7,
		Name:        "War Paint",
		Description: "Increases your foucs, your miss chance is decreased by 10%",
		Slots:       []EquipmentSlot{EquipmentSlotRightHand, EquipmentSlotLeftHand, EquipmentSlotHead, EquipmentSlotFeet, EquipmentSlotLeggings, EquipmentSlotTorso},
		Cost:        10,
		Item: &SimpleItem{
			Attributes: []ItemAttribute{ItemAttribute{Type: ItemAttributeMissChance, Amount: -10}},
		},
	},
}

func GetItemTypeById(id int) *ItemType {
	for _, v := range itemTypes {
		if v.Id == id {
			return v
		}
	}
	return nil
}
