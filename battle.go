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

	battleLog := "**Battle log**\n"

	attackersTurn := false
	for {
		attacker := b.Initiator
		defender := b.Defender
		if !attackersTurn {
			attacker = b.Defender
			defender = b.Initiator
		}

		// Check if defender dodged
		dodgeChance := defender.DodgeChance()
		if rand.Intn(100) < int(dodgeChance) {
			battleLog += fmt.Sprintf("**%s** Dodged **%s**\n", defender.Player.Name, attacker.Player.Name)
			attackersTurn = !attackersTurn
			continue
		}

		dmg := attacker.Damage() * (rand.Float32() + 0.5) // The damage varies from 50% to 150%
		originalHealth := defender.Health
		defender.Health -= float32(dmg)

		battleLog += fmt.Sprintf("**%s** Attacked **%s** with **%f** Damage! (**%f** -> **%f**)\n", attacker.Player.Name, defender.Player.Name, dmg, originalHealth, defender.Health)

		if defender.Health <= 0 {
			winner = attacker
			loser = defender
			break
		}

		attackersTurn = !attackersTurn
	}

	xpGain := int(float32(GetLevelFromXP(loser.Player.XP)/GetLevelFromXP(winner.Player.XP)) * 5)
	battleLog += fmt.Sprintf("**%s** Won against **%s** and earned %d XP! (%f vs %f)\n", winner.Player.Name, loser.Player.Name, xpGain, winner.Health, loser.Health)

	curLevel := GetLevelFromXP(winner.Player.XP)
	winner.Player.XP += xpGain
	newLevel := GetLevelFromXP(winner.Player.XP)
	if curLevel != newLevel {
		battleLog += fmt.Sprintf("**%s** Reached Level **%d**!", winner.Player.Name, newLevel)
	}

	go SendMessage(b.Channel, battleLog)
	b.Finished = true
	b.Running = false

	winner.Player.Wins++
	loser.Player.Losses++
}

func (b *Battle) ContainsPlayer(player *Player, lock bool) bool {
	if lock {
		b.RLock()
		defer b.RUnlock()
	}

	contains := b.Initiator.Player.Id == player.Id || b.Defender.Player.Id == player.Id
	return contains
}
