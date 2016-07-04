package core

type Effect interface {
	Init(owner *BattlePlayer, opponent *BattlePlayer, battle *Battle)
	Apply()
	Remove()

	OnTurn()
	OnAttack()
	OnDefend()
}
