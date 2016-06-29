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

	Channel   string
	Initiator *BattlePlayer
	Defender  *BattlePlayer

	Log     []string
	CurTurn int
}

func NewBattle(attacker *Player, defender *Player, channel string) *Battle {
	return &Battle{
		Initiator: NewBattlePlayer(attacker),
		Defender:  NewBattlePlayer(defender),
		Initiated: time.Now(),
		Channel:   channel,
	}
}

func (b *Battle) Expire(lock bool) {
	if lock {
		b.Lock()
		defer b.Unlock()
	}

	go SendMessage(b.Channel, "<@"+b.Initiator.Player.Id+"> Your battle with"+b.Defender.Player.Id+" Has expired")
}

func (b *Battle) Battle() {
	b.Running = true

	b.Initiator.Player.Lock()
	defer b.Initiator.Player.Unlock()

	b.Defender.Player.Lock()
	defer b.Defender.Player.Unlock()

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

		attacker.NextTurn()
		defender.NextTurn()

		attacker.Attack()
		defender.Defend()

		dmg := attacker.Damage() * (rand.Float32() + 0.5) // The damage varies from 50% to 150%
		b.DealDamage(attacker, defender, dmg, "Basic Attack")

		if defender.Health <= 0 {
			winner = attacker
			loser = defender
			break
		}

		attackersTurn = !attackersTurn
	}

	xpRatio := float32(GetLevelFromXP(loser.Player.XP)) / float32(GetLevelFromXP(winner.Player.XP))
	xpGain := int(xpRatio * 5)

	b.Log = append(b.Log, fmt.Sprintf("**%s** Won against **%s** and earned %d XP! (%.2f vs %.2f)\n", winner.Player.Name, loser.Player.Name, xpGain, winner.Health, loser.Health))

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

	winner.Player.Wins++
	loser.Player.Losses++
}

func (b *Battle) DealDamage(attacker *BattlePlayer, defender *BattlePlayer, damage float32, source string) {
	if damage >= 0 { // Don't dodge heals
		// Check if defender dodged
		dodgeChance := defender.DodgeChance()
		if rand.Intn(100) < int(dodgeChance) {
			b.AppendLog(fmt.Sprintf("**%s** Dodged **%s**'s %s", defender.Player.Name, attacker.Player.Name, source))
			return
		}
	}

	originalHealth := defender.Health
	defender.Health -= damage

	action := "Attacked"
	dealtHealed := "dealt"

	if damage < 0 {
		action = "Healed"
		dealtHealed = "healed"
		damage = -damage
	}

	b.AppendLog(fmt.Sprintf("**%s** %s **%s** using **%s** and %s **%.2f** Damage! (**%.2f** -> **%.2f**)",
		attacker.Player.Name, action, defender.Player.Name, source, dealtHealed, damage, originalHealth, defender.Health))
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
