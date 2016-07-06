package items

import (
	"github.com/jonas747/battlebot/core"
	"math/rand"
)

// Simple item that implements the Item interface
// Provides only static attributes
type SimpleItem struct {
	Attributes []core.ItemAttribute
	Player     *core.BattlePlayer
	Opponent   *core.BattlePlayer
	Battle     *core.Battle
}

func (s *SimpleItem) Init(wearer *core.BattlePlayer, opponent *core.BattlePlayer, battle *core.Battle) {
	s.Player = wearer
	s.Opponent = opponent
	s.Battle = battle
}

func (s *SimpleItem) GetStaticAttributes() []core.ItemAttribute {
	return s.Attributes
}

func (s *SimpleItem) Apply() {
	for _, attrib := range s.Attributes {
		s.Player.ApplyItemAttribute(attrib.Type, attrib.Amount)
	}
}

func (s *SimpleItem) Remove() {
	for _, attrib := range s.Attributes {
		// Undo the change
		s.Player.ApplyItemAttribute(attrib.Type, -attrib.Amount)
	}
}

func (s *SimpleItem) OnTurn()   {}
func (s *SimpleItem) OnAttack() {}
func (s *SimpleItem) OnDefend() {}

func (item *SimpleItem) GetCopy() core.Item {
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
	Apply    func(sender *core.BattlePlayer, receiver *core.BattlePlayer, battle *core.Battle)
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
func (item *ItemEffectEmitter) GetCopy() core.Item {
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
