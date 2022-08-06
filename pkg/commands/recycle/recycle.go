package recycle

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// RecycleCommands contains the /recycle command
type RecycleCommands struct {
	registry command.Registry
	db       *db.Connection
	xSecret  string
}

// NewCommands gives a new RecycleCommands
func NewCommands(db *db.Connection, xSecret string) *RecycleCommands {
	return &RecycleCommands{
		db:      db,
		xSecret: xSecret,
	}
}

// Register registers the handlers
func (p *RecycleCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("recycle", p.registerRecycle)

	p.registry = registry

}

// InstallSlashCommands registers the slash commands
func (p *RecycleCommands) InstallSlashCommands(session *discordgo.Session) error {
	go p.run(session) // run reminers in background

	apps := []discordgo.ApplicationCommand{
		{
			Name:        "recycle",
			Description: "Adds a recycling reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "street",
					Description: "Which street do you live on?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "housenumber",
					Description: "What is your house number?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "postalcode",
					Description: "What is your postal code?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "annoying",
					Description: "Should I annoy you till you did it?",
					Required:    true,
				},
			},
		},
	}

	for _, app := range apps {
		if err := slash.InstallSlashCommand(session, "", app); err != nil {
			return err
		}
	}

	return nil
}
