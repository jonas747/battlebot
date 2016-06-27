package main

import (
	"github.com/jonas747/discordgo"
	"sync"
)

var (
	playerManager = &PlayerManager{Players: make([]*Player, 0)}
)

type PlayerManager struct {
	sync.RWMutex
	Players []*Player
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
	Name string
	Id   string
	XP   int
}

func NewPlayer(user *discordgo.User) *Player {
	return &Player{
		Name: user.Username,
		Id:   user.ID,
		XP:   0,
	}
}

func GetLevelFromXP(xp int) int {
	if xp == 0 {
		return 1
	}
	return (xp / 10) + 1
}

func GetXPForLevel(level int) int {
	return (level * 10) + 1
}
