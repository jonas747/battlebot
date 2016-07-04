package core

import (
	"errors"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
	"strings"
)

const (
	VERSION = "BattleBot 0.0.4 Alpha"
)

var (
	flagToken string
	flagDebug bool
	dgo       *discordgo.Session
	Commands  []*CommandDef = make([]*CommandDef, 0)
	ItemTypes []*ItemType   = make([]*ItemType, 0)
)

func init() {
	flag.StringVar(&flagToken, "t", "", "Token to use")
	flag.BoolVar(&flagDebug, "d", false, "Set to turn on debug info, such as pprof http server")

	if !flag.Parsed() {
		flag.Parse()
	}
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Run() {
	log.Println("Launching " + VERSION)

	session, err := discordgo.New(flagToken)
	PanicErr(err)

	session.AddHandler(MessageHandler)
	session.AddHandler(HandleReady)
	session.AddHandler(HandleServerJoin)
	dgo = session
	err = session.Open()
	PanicErr(err)

	log.Println("Launched!")
	go Battles.Run()
	go Players.Run()

	if flagDebug {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	select {}
}

// Registers commands to the command system
// Only safe to call before bot has started
func RegisterCommands(cmds ...*CommandDef) {
	if Commands == nil {
		Commands = make([]*CommandDef, 0, len(cmds))
	}
	Commands = append(Commands, cmds...)
}

// Registers items to the item system
// Only safe to call before bot has started
func RegisterItems(items ...*ItemType) {
	if ItemTypes == nil {
		ItemTypes = make([]*ItemType, 0, len(items))
	}
	ItemTypes = append(ItemTypes, items...)
}

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if s.State == nil || s.State.User == nil {
		return // Wait till we have state initialized
	}

	if len(m.Mentions) == 0 {
		return
	}

	if strings.Index(m.Content, s.State.User.ID) == 2 {
		err := HandleCommand(m.Content, m)
		if err != nil {
			SendMessage(m.ChannelID, "Error: "+err.Error()+" See `@bot help` for more info")
			log.Println("Error handling command:", err)
		}
	}

}

var (
	ErrCommandEmpty    = errors.New("No comand specified")
	ErrCommandNotFound = errors.New("Command not found :'(")
)

func HandleCommand(cmd string, m *discordgo.MessageCreate) error {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			SendMessage(m.ChannelID, "Panic when handling Command!! ```\n"+stack+"\n```")
			log.Println("Recovered from panic ", r, "\n", m.Content, "\n", stack)
		}
	}()

	// Remove our mention
	cmd = strings.Replace(cmd, "<@"+dgo.State.User.ID+">", "", 1)
	cmd = strings.Replace(cmd, "<@!"+dgo.State.User.ID+">", "", 1)
	cmd = strings.TrimSpace(cmd)

	fields := strings.Fields(cmd)
	if len(fields) < 1 {
		return ErrCommandEmpty
	}

	cmdName := strings.ToLower(fields[0])

	for _, v := range Commands {

		match := v.Name == cmdName
		if !match {
			for _, alias := range v.Aliases {
				if alias == cmdName {
					match = true
					break
				}
			}
		}

		if match {
			parsed, err := ParseCommand(cmd, m, v)
			if err != nil {
				return err
			}

			if v.RunFunc != nil {
				v.RunFunc(parsed, m)
			}
			return nil
		}
	}

	return ErrCommandNotFound
}

func HandleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Ready received! Connected to", len(s.State.Guilds), "Guilds")
}

func HandleServerJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	log.Println("Joined guild", g.Name, " Connected to", len(s.State.Guilds), "Guilds")
}
