package main

import (
	"errors"
	"fmt"
	"github.com/jonas747/discordgo"
	"strconv"
	"strings"
)

type CommandDef struct {
	Name        string
	Description string
	Arguments   []*ArgumentDef
	RunFunc     func(cmd *ParsedCommand, m *discordgo.MessageCreate)
}

func (c *CommandDef) String() string {
	out := fmt.Sprintf("%s: %s.", c.Name, c.Description, c.Arguments)
	if len(c.Arguments) > 0 {
		out += fmt.Sprintf("(%v)", c.Arguments)
	}
	return out
}

type ArgumentType int

const (
	ArgumentTypeString ArgumentType = iota
	ArgumentTypeNumber
	ArgumentTypeUser
)

type ArgumentDef struct {
	Name string
	Type ArgumentType
}

func (a *ArgumentDef) String() string {
	typeStr := ""

	switch a.Type {
	case ArgumentTypeString:
		typeStr = "String"
	case ArgumentTypeNumber:
		typeStr = "Number"
	case ArgumentTypeUser:
		typeStr = "@User"
	}

	return a.Name + "(" + typeStr + ")"
}

type ParsedArgument struct {
	Raw    string
	Parsed interface{}
}

func (p *ParsedArgument) Int() int {
	val, _ := p.Parsed.(float64)
	return int(val)
}

func (p *ParsedArgument) Str() string {
	val, _ := p.Parsed.(string)
	return val
}

func (p *ParsedArgument) Float() float64 {
	val, _ := p.Parsed.(float64)
	return val
}

type ParsedCommand struct {
	Name string
	Cmd  *CommandDef
	Args []*ParsedArgument
}

var (
	ErrIncorrectNumArgs = errors.New("Icorrect number of arguments")
)

func ParseCommand(raw string, target *CommandDef) (*ParsedCommand, error) {
	// No arguments passed
	if len(target.Arguments) < 1 {
		return &ParsedCommand{
			Name: target.Name,
			Cmd:  target,
		}, nil
	}

	fields := strings.Fields(raw)

	if len(fields) != len(target.Arguments) {
		return nil, ErrIncorrectNumArgs
	}

	// Parse the arguments
	parsedArgs := make([]*ParsedArgument, len(target.Arguments))
	for k, field := range fields {
		var err error
		var val interface{}

		switch target.Arguments[k].Type {
		case ArgumentTypeNumber:
			val, err = strconv.ParseFloat(field, 64)
		case ArgumentTypeString:
			val = field
		case ArgumentTypeUser:
			// TODO
		}

		if err != nil {
			return nil, err
		}

		parsedArgs[k] = &ParsedArgument{
			Raw:    field,
			Parsed: val,
		}
	}

	return &ParsedCommand{
		Name: target.Name,
		Cmd:  target,
		Args: parsedArgs,
	}, nil
}
