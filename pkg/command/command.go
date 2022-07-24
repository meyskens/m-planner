package command

import (
	"github.com/bwmarrin/discordgo"
)

// Command is a struct of a bot command
type Command struct {
	Name        string
	Description string
}

// Registry is the interface of a command registry
type Registry interface {
	RegisterInteractionCreate(command string, fn func(*discordgo.Session, *discordgo.InteractionCreate))
}

// Interface defines how a command should be structured
type Interface interface {
	Register(registry Registry)
	InstallSlashCommands(session *discordgo.Session) error
}
