package planning

import (
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// PlanningCommands contains the /ideas command
type PlanningCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new PlanningCommands
func NewCommands(db *db.Connection) *PlanningCommands {
	return &PlanningCommands{
		db: db,
	}
}

// Register registers the handlers
func (p *PlanningCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("plan", p.registerCommand)
	registry.RegisterInteractionCreate("planning", p.listCommand)
	registry.RegisterInteractionCreate("planning_delete_list", p.listDeleteCommand)
	registry.RegisterInteractionCreate("change_planning", p.changeCommand)
	registry.RegisterInteractionCreate("planning_edit_list", p.changeCommand)
	registry.RegisterInteractionCreate("modal_change_planning", p.modalReturnCommand)
	registry.RegisterInteractionCreate("mark_planning_complete", p.markPlanningComplete)
	registry.RegisterInteractionCreate("snooze_planning", p.snoozePlanning)

	p.registry = registry

}

// InstallSlashCommands registers the slash commands
func (p *PlanningCommands) InstallSlashCommands(session *discordgo.Session) error {
	go p.run(session) // run reminers in background

	apps := []discordgo.ApplicationCommand{
		{
			Name:        "plan",
			Description: "Adds a reminder",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "What should i remind you about?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "when",
					Description: "When should I remind you",
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
		{
			Name:        "planning",
			Description: "Get list of reminders",
		},
	}

	for _, app := range apps {
		if err := slash.InstallSlashCommand(session, "", app); err != nil {
			return err
		}
	}

	return nil
}
