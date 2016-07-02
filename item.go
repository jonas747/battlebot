package main

import (
	"math/rand"
	"strings"
)

// General item definition
type ItemType struct {
	Id          int
	Slots       []EquipmentSlot
	Name        string
	Description string

	Cost int

	// The actual item
	// Items are not deep copied but simple copied so do not do
	// modifications to preset slices or maps or pointers
	Item Item
}

func (it *ItemType) GetItemCopy() Item {
	return it.Item
}

type Item interface {
	Effect

	// Passive attribute bonuses
	GetStaticAttributes() []Attribute

	// Returns a copy to be used and is goroutine safe
	GetCopy() Item
}

type EquipmentSlot int

const (
	EquipmentSlotNone EquipmentSlot = iota
	EquipmentSlotHead
	EquipmentSlotRightHand
	EquipmentSlotLeftHand
	EquipmentSlotFeet
	EquipmentSlotTorso
	EquipmentSlotLeggings
)

func StringEquipmentSlot(slot EquipmentSlot) string {
	switch slot {

	case EquipmentSlotNone:
		return "None"
	case EquipmentSlotHead:
		return "Head"
	case EquipmentSlotRightHand:
		return "RightHand"
	case EquipmentSlotLeftHand:
		return "LeftHand"
	case EquipmentSlotFeet:
		return "Feet"
	case EquipmentSlotTorso:
		return "Torso"
	case EquipmentSlotLeggings:
		return "Leggings"

	}

	return "Unknown"
}

func EquipmentSlotFromString(slot string) EquipmentSlot {
	switch strings.ToLower(slot) {

	case "none":
		return EquipmentSlotNone
	case "head":
		return EquipmentSlotHead
	case "righthand":
		return EquipmentSlotRightHand
	case "lefthand":
		return EquipmentSlotLeftHand
	case "feet":
		return EquipmentSlotFeet
	case "torso":
		return EquipmentSlotTorso
	case "leggings":
		return EquipmentSlotLeggings

	}

	return EquipmentSlotNone
}

// An item in a players inventory
type PlayerItem struct {
	Id            int
	EquipmentSlot EquipmentSlot
}

// An item in an equipment slot
type EquippedItem struct {
	slot EquipmentSlot
	id   int
}

/////////////////////////
// Item implementations
/////////////////////////

// Simple item that implements the Item interface
// Provides only static attributes
type SimpleItem struct {
	Attributes []Attribute
	Player     *BattlePlayer
	Opponent   *BattlePlayer
	Battle     *Battle
}

func (s *SimpleItem) Init(wearer *BattlePlayer, opponent *BattlePlayer, battle *Battle) {
	s.Player = wearer
	s.Opponent = opponent
	s.Battle = battle
}

func (s *SimpleItem) GetStaticAttributes() []Attribute {
	return s.Attributes
}

func (s *SimpleItem) Apply() {
	for _, attrib := range s.Attributes {
		s.Player.Attributes.Modify(attrib.Type, attrib.Val)
	}
}

func (s *SimpleItem) Remove() {
	for _, attrib := range s.Attributes {
		// Undo the change
		s.Player.Attributes.Modify(attrib.Type, -attrib.Val)
	}
}

func (s *SimpleItem) OnTurn()   {}
func (s *SimpleItem) OnAttack() {}
func (s *SimpleItem) OnDefend() {}

func (item *SimpleItem) GetCopy() Item {
	temp := *item
	return &temp
}

type Target int

const (
	TargetSelf Target = iota
	TargetOpponent
)

type EffectTriggerType int

const (
	EffectTriggerTurn EffectTriggerType = iota
	EffectTriggerAttack
	EffectTriggerDefend
)

type EffectTrigger struct {
	Chance   float32
	Target   Target
	Trigger  EffectTriggerType
	Apply    func(sender *BattlePlayer, receiver *BattlePlayer, battle *Battle)
	OnlyOnce bool
}

func (e *EffectTrigger) MaybeTrigger(parent *ItemEffectEmitter) bool {
	if e.Apply == nil {
		return false
	}
	if e.Chance != 0 {
		rng := rand.Float32()
		if rng > e.Chance {
			return false
		}
	}

	if e.Target == TargetSelf {
		e.Apply(parent.Player, parent.Player, parent.Battle)
	} else {
		e.Apply(parent.Player, parent.Opponent, parent.Battle)
	}
	return true
}

// Similar to simpleitem but also gives a n% chance of an effect to be given to the opponent or defender on an event
type ItemEffectEmitter struct {
	SimpleItem

	Triggers []*EffectTrigger

	triggeredEffects []int
}

func (item *ItemEffectEmitter) OnTurn()   { item.handleTrigger(EffectTriggerTurn) }
func (item *ItemEffectEmitter) OnAttack() { item.handleTrigger(EffectTriggerAttack) }
func (item *ItemEffectEmitter) OnDefend() { item.handleTrigger(EffectTriggerDefend) }
func (item *ItemEffectEmitter) GetCopy() Item {
	temp := *item
	return &temp
}

func (item *ItemEffectEmitter) handleTrigger(triggerType EffectTriggerType) {

OUTER:
	for i, triggerListener := range item.Triggers {
		if triggerListener.Trigger != triggerType {
			continue
		}

		if triggerListener.OnlyOnce {
			for _, triggered := range item.triggeredEffects {
				if triggered == i {
					continue OUTER
				}
			}
		}

		if triggerListener.MaybeTrigger(item) && triggerListener.OnlyOnce {
			item.triggeredEffects = append(item.triggeredEffects, i)
		}
	}
}
