package core

import (
	"strings"
)

func GetItemTypeById(id int) *ItemType {
	for _, v := range ItemTypes {
		if v.Id == id {
			return v
		}
	}
	return nil
}

// General item definition
type ItemType struct {
	Id          int
	Slots       []EquipmentSlot
	Name        string
	Description string

	Cost    int
	CanBuy  bool
	CanSell bool

	// The actual item
	// Items are not deep copied but simple copied so do not do
	// modifications to values pointed to by preset pointers (e.g preset slices and pointers)
	Item Item
}

type Item interface {
	Effect

	// Passive attribute bonuses
	GetStaticAttributes() []ItemAttribute

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

func (slot EquipmentSlot) String() string {
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
// Item is equipped if EquipmentSlot is not none
type PlayerItem struct {
	Id            int
	EquipmentSlot EquipmentSlot
}

type ItemAttributeType int

const (
	ItemAttributeStrength ItemAttributeType = iota
	ItemAttributeAgility
	ItemAttributeStamina
	ItemAttributeDodgeChance
	ItemAttributeMissChance
)

func (i ItemAttributeType) String() string {
	switch i {
	case ItemAttributeStrength:
		return "Strength"
	case ItemAttributeAgility:
		return "Agility"
	case ItemAttributeStamina:
		return "Stamina"
	case ItemAttributeDodgeChance:
		return "DodgeChance"
	case ItemAttributeMissChance:
		return "MissChance"
	}
	return "Unknown"
}

type ItemAttribute struct {
	Type   ItemAttributeType
	Amount float32
}
