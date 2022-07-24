package daily

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// DailyCommands contains the /ideas command
type DailyCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new DailyCommands
func NewCommands(db *db.Connection) *DailyCommands {
	return &DailyCommands{
		db: db,
	}
}

// Register registers the handlers
func (d *DailyCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("daily", d.registerCommand)
	registry.RegisterInteractionCreate("dailys", d.listCommand)
	registry.RegisterInteractionCreate("daily_delete_list", d.listDeleteCommand)
	registry.RegisterInteractionCreate("change_daily", d.changeCommand)
	registry.RegisterInteractionCreate("daily_edit_list", d.changeCommand)
	registry.RegisterInteractionCreate("modal_change_daily", d.modalReturnCommand)
	registry.RegisterInteractionCreate("mark_event_complete", d.markEventComplete)
	registry.RegisterInteractionCreate("snooze_event", d.snoozeEvent)

	d.registry = registry

}

// InstallSlashCommands registers the slash commands
func (d *DailyCommands) InstallSlashCommands(session *discordgo.Session) error {
	go d.run(session) // run reminers in background

	apps := []discordgo.ApplicationCommand{
		{
			Name:        "daily",
			Description: "Adds a daily reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "What should i remind you about?",
					Required:    true,
				},
			},
		},
		{
			Name:        "dailys",
			Description: "Get list of daily reminders",
		},
	}

	for _, app := range apps {
		if err := slash.InstallSlashCommand(session, "", app); err != nil {
			return err
		}
	}

	return nil
}
