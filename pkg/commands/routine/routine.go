package routine

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// RoutineCommands contains the /routine command
type RoutineCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new DailyCommands
func NewCommands(db *db.Connection) *RoutineCommands {
	return &RoutineCommands{
		db: db,
	}
}

// Register registers the handlers
func (d *RoutineCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("routine", d.registerCommand)
	registry.RegisterInteractionCreate("routines", d.listCommand)
	registry.RegisterInteractionCreate("routine_delete_list", d.listDeleteCommand)
	registry.RegisterInteractionCreate("change_routine", d.changeCommand)
	registry.RegisterInteractionCreate("routine_edit_list", d.changeCommand)
	registry.RegisterInteractionCreate("modal_change_routine", d.modalReturnCommand)

	d.registry = registry

}

// InstallSlashCommands registers the slash commands
func (d *RoutineCommands) InstallSlashCommands(session *discordgo.Session) error {
	go d.run(session) // run reminers in background

	apps := []discordgo.ApplicationCommand{
		{
			Name:        "routine",
			Description: "Adds a routine reminder",
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
			Name:        "routines",
			Description: "Get list of daily routines",
		},
	}

	for _, app := range apps {
		if err := slash.InstallSlashCommand(session, "", app); err != nil {
			return err
		}
	}

	return nil
}
