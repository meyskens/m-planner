package ideas

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// IdeasCommands contains the /ideas command
type IdeasCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new IdeasCommands
func NewCommands(db *db.Connection) *IdeasCommands {
	return &IdeasCommands{
		db: db,
	}
}

// Register registers the handlers
func (i *IdeasCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("idea", i.registerCommand)
	registry.RegisterInteractionCreate("ideas", i.listCommand)
	registry.RegisterInteractionCreate("delete_idea", i.deleteCommand)
	registry.RegisterInteractionCreate("idea_delete_list", i.listDeleteCommand)
	i.registry = registry

}

// InstallSlashCommands registers the slash commands
func (i *IdeasCommands) InstallSlashCommands(session *discordgo.Session) error {
	apps := []discordgo.ApplicationCommand{
		{
			Name:        "idea",
			Description: "Adds an idea to the list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "What is the idea?",
					Required:    true,
				},
			},
		},
		{
			Name:        "ideas",
			Description: "Get list of ideas",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "print",
					Description: "print out list as well",
					Required:    false,
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
