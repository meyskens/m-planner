package groceries

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// GroceriesCommands contains the /groceries command
type GroceriesCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new IdeasCommands
func NewCommands(db *db.Connection) *GroceriesCommands {
	return &GroceriesCommands{
		db: db,
	}
}

// Register registers the handlers
func (g *GroceriesCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("grocery", g.registerCommand)
	registry.RegisterInteractionCreate("groceries", g.listCommand)
	registry.RegisterInteractionCreate("delete_grocery", g.deleteCommand)
	registry.RegisterInteractionCreate("grocery_delete_list", g.listDeleteCommand)
	g.registry = registry

}

// InstallSlashCommands registers the slash commands
func (g *GroceriesCommands) InstallSlashCommands(session *discordgo.Session) error {
	apps := []discordgo.ApplicationCommand{
		{
			Name:        "grocery",
			Description: "Adds an grocery to the list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "item",
					Description: "What do we need?",
					Required:    true,
				},
			},
		},
		{
			Name:        "groceries",
			Description: "Get list of groceries",
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
