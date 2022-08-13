package token

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// TokenCommands contains the /token command
type TokenCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new TokenCommands
func NewCommands(db *db.Connection) *TokenCommands {
	return &TokenCommands{
		db: db,
	}
}

// Register registers the handlers
func (p *TokenCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("token", p.registerToken)

	p.registry = registry

}

// InstallSlashCommands registers the slash commands
func (p *TokenCommands) InstallSlashCommands(session *discordgo.Session) error {

	apps := []discordgo.ApplicationCommand{
		{
			Name:        "token",
			Description: "Created an API token",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	}

	for _, app := range apps {
		if err := slash.InstallSlashCommand(session, "", app); err != nil {
			return err
		}
	}

	return nil
}
