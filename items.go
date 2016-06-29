package main

var itemTypes = []*ItemType{
	&ItemType{
		Id:          0,
		Name:        "Poor mans boots",
		Description: "Simple boots that increase your stamina by 1",
		Slots:       []EquipmentSlot{EquipmentSlotFeet},
		Cost:        1,
		Item: &SimpleItem{
			Attributes: []Attribute{Attribute{Type: AttributeStamina, Val: 1}},
		},
	},
	&ItemType{
		Id:          1,
		Name:        "Watermelon",
		Description: "Increases your stamina by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotHead},
		Item: &SimpleItem{
			Attributes: []Attribute{Attribute{Type: AttributeStamina, Val: 5}},
		},
	},
	&ItemType{
		Id:          2,
		Name:        "Knife",
		Description: "Increases your strength by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotRightHand, EquipmentSlotLeftHand},
		Item: &SimpleItem{
			Attributes: []Attribute{Attribute{Type: AttributeStrength, Val: 5}},
		},
	},
	&ItemType{
		Id:          3,
		Name:        "Speeeed Bootz",
		Description: "Increases your agility by 5",
		Cost:        10,
		Slots:       []EquipmentSlot{EquipmentSlotFeet},
		Item: &SimpleItem{
			Attributes: []Attribute{Attribute{Type: AttributeAgility, Val: 5}},
		},
	},
	&ItemType{
		Id:          4,
		Name:        "Holy Torso",
		Description: "Every turn you heal 2 damage",
		Slots:       []EquipmentSlot{EquipmentSlotTorso},
		Cost:        10,
		Item: &ItemChanceEmitEffect{
			SimpleItem: SimpleItem{
				Attributes: []Attribute{Attribute{Type: AttributeStamina, Val: 2}},
			},
			Triggers: []*EffectTriggerChance{
				&EffectTriggerChance{
					Chance:  0.2,
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
		Slots:       []EquipmentSlot{EquipmentSlotTorso},
		Cost:        10,
		Item: &ItemChanceEmitEffect{
			SimpleItem: SimpleItem{
				Attributes: []Attribute{Attribute{Type: AttributeStrength, Val: 5}},
			},
			Triggers: []*EffectTriggerChance{
				&EffectTriggerChance{
					Chance: 0.2,
					Target: TargetOpponent,
					Apply: func(sender *BattlePlayer, receiver *BattlePlayer, battle *Battle) {
						battle.DealDamage(sender, receiver, -10, "Flowers")
					},
				},
			},
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
