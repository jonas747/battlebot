package main

import (
	"log"
)

type BattlePlayer struct {
	Health float32
	Player *Player

	// Extra attributes from items and buffs/debuffs
	Attributes    AttributeContainer
	EquippedItems []Item
}

func NewBattlePlayer(player *Player) *BattlePlayer {
	return &BattlePlayer{
		Player: player,
	}
}

func (p *BattlePlayer) Init(opponent *BattlePlayer, battle *Battle) { // initializes and applies the base health
	for _, v := range p.Player.Inventory {
		if v.EquipmentSlot != EquipmentSlotNone {

			itemType := GetItemTypeById(v.Id)
			if itemType == nil {
				log.Println("Unknown item ", v.Id, "on", p.Player.Name)
				continue
			}

			p.EquippedItems = append(p.EquippedItems, itemType.Item.GetCopy())
		}
	}

	for _, v := range p.EquippedItems {
		v.Init(p, opponent, battle)
		v.Apply()
	}
	p.Health = float32(p.Player.MaxHealth() + p.Attributes.Get(AttributeStamina))
}

func (p *BattlePlayer) NextTurn() {
	for _, v := range p.EquippedItems {
		v.OnTurn()
	}
}

func (p *BattlePlayer) Attack() {
	for _, v := range p.EquippedItems {
		v.OnAttack()
	}
}

func (p *BattlePlayer) GetCombinedAttribute(a AttributeType) int {
	return p.Player.Attributes.Get(a) + p.Attributes.Get(a)
}

func (p *BattlePlayer) DodgeChance() float32 {
	return GetDodgeChance(float32(p.GetCombinedAttribute(AttributeAgility)))
}

func (p *BattlePlayer) Damage() float32 {
	return p.Player.BaseDamage() + float32(p.Attributes.Get(AttributeStrength))
}

func GetDodgeChance(agility float32) float32 {
	return ((agility / (agility + 100)) * 80) + 20
}
