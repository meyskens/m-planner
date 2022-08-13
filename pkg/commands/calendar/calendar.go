package calendar

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// CalendarCommands contains the /calendar command
type CalendarCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new CalendarCommands
func NewCommands(db *db.Connection) *CalendarCommands {
	return &CalendarCommands{
		db: db,
	}
}

// Register registers the handlers
func (c *CalendarCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("calendar-link", c.registerCommand)
	registry.RegisterInteractionCreate("calendar", c.listCommand)
	registry.RegisterInteractionCreate("calendar_delete_list", c.listDeleteCommand)
	c.registry = registry

}

// InstallSlashCommands registers the slash commands
func (c *CalendarCommands) InstallSlashCommands(session *discordgo.Session) error {
	apps := []discordgo.ApplicationCommand{
		{
			Name:        "calendar-link",
			Description: "Adds an iCal to the calendar module",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "What is this calendar for?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Where is this calendar?",
					Required:    true,
				},
			},
		},
		{
			Name:        "calendar",
			Description: "Get your calendar",
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
