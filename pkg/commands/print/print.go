package print

import (
	"fmt"

	"github.com/meyskens/m-planner/pkg/db"
	printlib "github.com/meyskens/m-planner/pkg/print"
	"github.com/meyskens/m-planner/pkg/util/slash"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
)

// PrintCommands contains the /ideas command
type PrintCommands struct {
	registry command.Registry
	db       *db.Connection
}

// NewCommands gives a new PrintCommands
func NewCommands(db *db.Connection) *PrintCommands {
	return &PrintCommands{
		db: db,
	}
}

// Register registers the handlers
func (p *PrintCommands) Register(registry command.Registry) {
	registry.RegisterInteractionCreate("print", p.printCommand)
	p.registry = registry
}

// InstallSlashCommands registers the slash commands
func (p *PrintCommands) InstallSlashCommands(session *discordgo.Session) error {
	apps := []discordgo.ApplicationCommand{
		{
			Name:        "print",
			Description: "Print something out",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "What should I print?",
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

func (p *PrintCommands) printCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	text := ""

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			text = str
		}
	}

	if text == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry friend, i got no conent :(",
			},
		})
		return
	}

	pd, err := printlib.PrintText(i.Member.User.ID, text)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a printing error %q :(", err),
			},
		})
		return
	}
	for _, pd := range pd {
		if err := p.db.Create(&pd).Error; err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", err),
				},
			})
			return
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "I will print this!",
		},
	})
}
