package main

import (
	"errors"
	"flag"
	"github.com/jonas747/discordgo"
	"log"
	"strings"
)

const (
	VERSION = "BattleBot 0.0.1 Alpha"
)

var (
	flagToken string
	dgo       *discordgo.Session
	commands  []*CommandDef
)

func init() {
	flag.StringVar(&flagToken, "t", "", "Token to use")
	flag.Parse()

	commands = CommonCommands
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Println("Launching battlebot v" + VERSION)

	session, err := discordgo.New(flagToken)
	PanicErr(err)

	session.AddHandler(MessageHandler)
	session.AddHandler(HandleReady)
	dgo = session
	err = session.Open()
	PanicErr(err)

	log.Println("Launched!")
	select {}
}

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if s.State == nil || s.State.User == nil {
		return // Wait till we have state initialized
	}

	if len(m.Mentions) > 0 && m.Mentions[0].ID == s.State.User.ID {
		err := HandleCommand(m.Content, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error()+"See `@bot help` for more info")
		}
	}
}

var (
	ErrCommandEmpty    = errors.New("No comand specified")
	ErrCommandNotFound = errors.New("Command not found :'(")
)

func HandleCommand(cmd string, m *discordgo.MessageCreate) error {
	// Remove our mention
	cmd = strings.Replace(cmd, "<@"+dgo.State.User.ID+">", "", 1)
	cmd = strings.TrimSpace(cmd)

	fields := strings.Fields(cmd)
	if len(fields) < 1 {
		return ErrCommandEmpty
	}

	for _, v := range commands {
		if v.Name == strings.ToLower(fields[0]) {
			parsed, err := ParseCommand(cmd, v)
			if err != nil {
				return err
			}

			if v.RunFunc != nil {
				v.RunFunc(parsed, m)
			}
		}
	}

	return ErrCommandNotFound
}

func HandleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Ready received!")
}
