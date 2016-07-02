package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var battleManager = &BattleManager{
	Battles: make([]*Battle, 0),
}

type BattleManager struct {
	sync.RWMutex

	Battles []*Battle
}

func (bm *BattleManager) MaybeAddBattle(battle *Battle) bool {
	bm.Lock()
	defer bm.Unlock()

	for _, v := range bm.Battles {

		v.RLock()
		if v.ContainsPlayer(battle.Initiator.Player, false) || v.ContainsPlayer(battle.Defender.Player, false) {
			v.RUnlock()
			return false // Already battling
		}
		v.RUnlock()
	}

	bm.Battles = append(bm.Battles, battle)
	return true
}

func (bm *BattleManager) MaybeAcceptBattle(id string) bool {
	bm.Lock()
	defer bm.Unlock()

	for _, battle := range bm.Battles {
		battle.RLock()
		if battle.Defender.Player.Id == id {
			battle.RUnlock()

			go func() {
				battle.Lock()
				if battle.Running || battle.Finished {
					battle.Unlock()
					return
				}

				battle.Battle()
				battle.Unlock()
			}()

			return true // BAttle possibly accepted
		} else {
			battle.RUnlock()
		}
	}

	return false
}

func (bm *BattleManager) Run() {
	for {
		ticker := time.NewTicker(time.Second)
		select {
		case <-ticker.C:
			bm.Lock()
			bm.CheckBattles()
			bm.Unlock()
		}
	}
}

func (bm *BattleManager) CheckBattles() {
	newCopy := make([]*Battle, 0)
	for _, v := range bm.Battles {
		v.Lock()
		if time.Since(v.Initiated) < time.Minute && !v.Finished {
			newCopy = append(newCopy, v)
		} else {
			if !v.Finished {
				v.Expire(false)
			}
		}
		v.Unlock()
	}
	bm.Battles = newCopy
}

type Battle struct {
	sync.RWMutex

	Initiated time.Time

	Finished bool
	Running  bool

	Money int

	Channel   string
	Initiator *BattlePlayer
	Defender  *BattlePlayer

	IsMonster bool
	Log       []string
	CurTurn   int
}

func NewBattle(attacker *Player, defender *Player, money int, channel string) *Battle {
	return &Battle{
		Initiator: NewBattlePlayer(attacker),
		Defender:  NewBattlePlayer(defender),
		Initiated: time.Now(),
		Channel:   channel,
		Money:     money,
	}
}

func (b *Battle) Expire(lock bool) {
	if lock {
		b.Lock()
		defer b.Unlock()
	}

	go SendMessage(b.Channel, "<@"+b.Initiator.Player.Id+"> Your battle with"+b.Defender.Player.Id+" Has expired")
}

func (b *Battle) CheckMoney() bool {
	if !b.IsMonster {
		if b.Initiator.Player.Money < b.Money {
			return false
		}
	}

	if b.Defender.Player.Money < b.Money {
		return false
	}

	return true
}

func (b *Battle) Battle() {
	b.Running = true

	b.Initiator.Player.Lock()
	defer b.Initiator.Player.Unlock()

	b.Defender.Player.Lock()
	defer b.Defender.Player.Unlock()

	if !b.CheckMoney() {
		go SendMessage(b.Channel, "Not enough money to battle...")
		b.Finished = true
		b.Running = false
		return
	}

	var winner *BattlePlayer
	var loser *BattlePlayer

	b.Initiator.Init(b.Defender, b)
	b.Defender.Init(b.Initiator, b)

	attackersTurn := false
	for {
		b.CurTurn++

		attacker := b.Initiator
		defender := b.Defender
		if !attackersTurn {
			attacker = b.Defender
			defender = b.Initiator
		}

		b.Turn(attacker, defender)

		if defender.Health <= 0 {
			winner = attacker
			loser = defender
			break
		}

		attackersTurn = !attackersTurn

	}
	b.End(winner, loser)
}

func (b *Battle) End(winner, loser *BattlePlayer) {
	xpRatio := float32(GetLevelFromXP(loser.Player.XP)) / float32(GetLevelFromXP(winner.Player.XP))
	xpGain := int(xpRatio * 5)

	b.Log = append(b.Log, fmt.Sprintf("**%s** Won against **%s** and earned %d$ and %d XP! (**%.2f** vs **%.2f**)\n", winner.Player.Name, loser.Player.Name, b.Money, xpGain, winner.Health, loser.Health))

	curLevel := GetLevelFromXP(winner.Player.XP)
	winner.Player.XP += xpGain
	newLevel := GetLevelFromXP(winner.Player.XP)
	if curLevel != newLevel {
		b.Log = append(b.Log, fmt.Sprintf("**%s** Reached Level **%d**!", winner.Player.Name, newLevel))
	}

	out := "**Battle Log**:\n"
	for _, msg := range b.Log {
		out += msg + "\n"
	}

	go SendMessage(b.Channel, out)
	b.Finished = true
	b.Running = false

	winner.Player.Money += b.Money
	if !b.IsMonster {
		loser.Player.Money -= b.Money
	}

	winner.Player.Wins++
	loser.Player.Losses++
}

func (b *Battle) Turn(attacker, defender *BattlePlayer) {
	attacker.NextTurn()
	defender.NextTurn()

	if attacker.StunDuration > 0 {
		b.AppendLog(fmt.Sprintf("**%s** :zzz: (%d turn(s) left)", attacker.Player.Name, attacker.StunDuration))
		return
	}

	attacker.Attack()
	defender.Defend()

	b.DealDamage(attacker, defender, attacker.Damage(), "Basic Attack")
}

func (b *Battle) DealDamage(attacker *BattlePlayer, defender *BattlePlayer, damage float32, source string) {
	//origDamage := damage
	modifier := rand.Float32() + 0.5
	damage = damage * modifier // The damage varies from 50% to 150%

	if damage >= 0 { // Don't dodge/miss heals
		// Check if attacker missed
		missChance := attacker.MissChance()
		if rand.Intn(100) < int(missChance) {
			b.AppendLog(fmt.Sprintf("**%s** Missed **%s** with %s", attacker.Player.Name, defender.Player.Name, source))
			return
		}

		// Check if defender dodged
		dodgeChance := defender.DodgeChance()
		if rand.Intn(100) < int(dodgeChance) {
			b.AppendLog(fmt.Sprintf("**%s** Dodged **%s**'s %s", defender.Player.Name, attacker.Player.Name, source))
			return
		}

	}

	originalHealth := defender.Health
	defender.Health -= damage

	action := ":crossed_swords:"
	dealtHealed := "dealt"

	if damage < 0 {
		action = "💓"
		dealtHealed = "healed"
		damage = -damage
	}

	b.AppendLog(fmt.Sprintf("**%s** %s **%s** using **%s** and %s **%.1f** (%.2f:game_die:) (**%.1f** -> **%.1f:hearts:**)",
		attacker.Player.Name, action, defender.Player.Name, source, dealtHealed, damage, modifier, originalHealth, defender.Health))
}

func (b *Battle) Stun(attacker, defender *BattlePlayer, duration int, source string) {
	defender.StunDuration += duration
	b.AppendLog(fmt.Sprintf("**%s** Stunned **%s** for %d turn(s) using %s", attacker.Player.Name, defender.Player.Name, duration, source))
}

func (b *Battle) AppendLog(msg string) {
	b.Log = append(b.Log, fmt.Sprintf("[%d]: %s", b.CurTurn, msg))
}

func (b *Battle) ContainsPlayer(player *Player, lock bool) bool {
	if lock {
		b.RLock()
		defer b.RUnlock()
	}

	contains := b.Initiator.Player.Id == player.Id || b.Defender.Player.Id == player.Id
	return contains
}
