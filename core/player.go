package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var (
	Players = &PlayerManager{Players: make([]*Player, 0)}
)

type PlayerManager struct {
	sync.RWMutex
	Players []*Player
}

func (pm *PlayerManager) Run() {
	err := pm.Load()
	if err != nil {
		log.Println("Failed loading data, consider using backup")
	}

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			err := pm.Save()
			if err != nil {
				log.Printf("Error saving log: ", err)
			}
		}
	}
}

func (pm *PlayerManager) Load() error {
	file, err := ioutil.ReadFile("players.json")
	if err != nil {
		return err
	}
	var decoded []*Player
	err = json.Unmarshal(file, &decoded)
	if err != nil {
		return err
	}

	pm.Lock()
	pm.Players = decoded
	pm.Unlock()
	return nil
}

func (pm *PlayerManager) Save() error {
	// Rotate savedata if existing
	_, err := os.Stat("players.json")
	if err == nil {
		err = CopyFile("players.json", "players.json.1")
		if err != nil {
			return err
		}
	}

	pm.Lock()
	out, err := json.Marshal(pm.Players)
	pm.Unlock()
	if err != nil {
		return err
	}

	file, err := os.Create("players.json")
	if err != nil {
		return err
	}
	file.Write(out)
	return file.Close()
}

func (pm *PlayerManager) AddPlayer(player *Player, lock bool) {
	if lock {
		pm.Lock()
		defer pm.Unlock()
	}

	pm.Players = append(pm.Players, player)
}

func (pm *PlayerManager) GetCreatePlayer(id, name string) *Player {
	pm.Lock()
	defer pm.Unlock()

	for _, v := range pm.Players {
		if v.Id == id {
			return v
		}
	}

	player := &Player{
		Name: name,
		Id:   id,
	}
	pm.AddPlayer(player, false)
	return player
}

type Player struct {
	sync.RWMutex
	Name  string
	Id    string
	XP    int
	Money int

	Wins   int
	Losses int

	Attributes AttributeContainer
	Inventory  []*PlayerItem
}

func NewPlayer(user *discordgo.User) *Player {
	return &Player{
		Name: user.Username,
		Id:   user.ID,
		XP:   0,
	}
}

func GetLevelFromXP(xp int) int {
	return (xp / 10) + 1
}

func GetXPForLevel(level int) int {
	return (level - 1) * 10
}

func (p *Player) EquipItem(invSlot int, eqSlot EquipmentSlot) error {
	if invSlot >= len(p.Inventory) {
		return errors.New("That inventory slot does not exist, check with the inventory command")
	}

	for _, v := range p.Inventory {
		if v.EquipmentSlot == eqSlot {
			v.EquipmentSlot = EquipmentSlotNone
		}
	}

	p.Inventory[invSlot].EquipmentSlot = eqSlot
	return nil
}

func (p *Player) MaxHealth() int {
	return GetLevelFromXP(p.XP) + p.Attributes.Get(AttributeStamina) + 10
}

func (p *Player) UsedAttributePoints() int {
	return p.Attributes.Get(AttributeStrength) + p.Attributes.Get(AttributeStamina) + p.Attributes.Get(AttributeAgility)
}

func (p *Player) AvailableAttributePoints() int {
	return GetLevelFromXP(p.XP) - p.UsedAttributePoints()
}

func (p *Player) BaseDodgeChange() float32 {
	return GetDodgeChance(float32(GetLevelFromXP(p.XP) + p.Attributes.Get(AttributeAgility)))
}

func (p *Player) BaseDamage() float32 {
	return float32(GetLevelFromXP(p.XP)+p.Attributes.Get(AttributeStrength)) + 2
}

func (p *Player) BaseMissChance() float32 {
	return GetMissChance(float32(GetLevelFromXP(p.XP) + p.Attributes.Get(AttributeAgility)))
}

func (p *Player) GetPrettyDiscordStats() string {
	level := GetLevelFromXP(p.XP)
	next := GetXPForLevel(level+1) - GetXPForLevel(level)
	curXp := p.XP - GetXPForLevel(level)

	// General info
	general := fmt.Sprintf(" - Level: %d\n - Attribute points: %d\n - XP: %d (%d)\n - Money %d$\n - Wins: %d\n - Losses: %d",
		GetLevelFromXP(p.XP), p.AvailableAttributePoints(), curXp, next, p.Money, p.Wins, p.Losses)

	// Create a battleplayer since that manages item stats for us
	bp := NewBattlePlayer(p)
	bp.Init(nil, nil)

	attributes := fmt.Sprintf(" - Strength: %d (+%d) (increases damage)\n - Stamina: %d(+%d) (increases health)\n - Agility: %d (+%d) (increases dodge chance, decreases miss chance)",
		p.Attributes.Get(AttributeStrength), bp.Attributes.Get(AttributeStrength), p.Attributes.Get(AttributeStamina), bp.Attributes.Get(AttributeStamina), p.Attributes.Get(AttributeAgility), bp.Attributes.Get(AttributeAgility))

	stats := fmt.Sprintf(" - Health: %.2f\n - Damage %.2f\n - Dodge Chance: %.2f%%\n - Miss Chance: %.2f%%",
		bp.MaxHealth(), bp.Damage(), bp.DodgeChance(), bp.MissChance())

	return general + "\n\n" + attributes + "\n\n" + stats
}
