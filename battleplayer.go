package main

import (
	"log"
)

// Battleplayer represents the player in a battle
type BattlePlayer struct {
	Health float32
	Player *Player

	// Extra attributes from items and buffs/debuffs
	Attributes    AttributeContainer
	EquippedItems []Item
	Effects       []Effect

	// Duration this player is stunned for
	StunDuration int

	ModifiedMissChance  float32
	ModifiedDodgeChance float32
	ModifiedDamage      float32
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
	p.Health = p.MaxHealth()
}

func (p *BattlePlayer) MaxHealth() float32 {
	return float32(p.Player.MaxHealth() + p.Attributes.Get(AttributeStamina))
}

func (p *BattlePlayer) NextTurn() {
	for _, v := range p.EquippedItems {
		v.OnTurn()
	}

	for _, v := range p.Effects {
		v.OnTurn()
	}
}

func (p *BattlePlayer) Attack() {
	for _, v := range p.EquippedItems {
		v.OnAttack()
	}

	for _, v := range p.Effects {
		v.OnAttack()
	}
}

func (p *BattlePlayer) Defend() {
	for _, v := range p.EquippedItems {
		v.OnDefend()
	}

	for _, v := range p.Effects {
		v.OnDefend()
	}
}

func (p *BattlePlayer) GetCombinedAttribute(a AttributeType) int {
	return p.Player.Attributes.Get(a) + p.Attributes.Get(a)
}

func (p *BattlePlayer) DodgeChance() float32 {
	return GetDodgeChance(float32(p.GetCombinedAttribute(AttributeAgility))) + p.ModifiedDodgeChance
}

func (p *BattlePlayer) Damage() float32 {
	return p.Player.BaseDamage() + float32(p.Attributes.Get(AttributeStrength)) + float32(p.ModifiedDamage)
}

func (p *BattlePlayer) MissChance() float32 {
	return GetMissChance(float32(p.GetCombinedAttribute(AttributeAgility)) + p.ModifiedMissChance)
}

func GetDodgeChance(agility float32) float32 {
	return ((agility / (agility + 100)) * 80) + 20
}

func GetMissChance(agility float32) float32 {
	return 50 - ((agility / (agility + 100)) * 50)
}
