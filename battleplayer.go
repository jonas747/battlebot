package main

type BattlePlayer struct {
	Health float32
	Player *Player
}

func NewBattlePlayer(player *Player) *BattlePlayer {
	return &BattlePlayer{
		Player: player,
		Health: float32(player.MaxHealth()),
	}
}
func (p *BattlePlayer) DodgeChance() float32 {
	return p.Player.BaseDodgeChange()
}

func (p *BattlePlayer) Damage() float32 {
	return p.Player.BaseDamage()
}
