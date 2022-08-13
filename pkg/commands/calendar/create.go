package calendar

import (
	"fmt"
	"log"

	"github.com/meyskens/m-planner/pkg/db"

	"github.com/bwmarrin/discordgo"
)

func (c *CalendarCommands) registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := ""
	link := ""

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			name = str
		}
		if str, ok := i.ApplicationCommandData().Options[1].Value.(string); ok {
			link = str
		}
	}

	if link == "" || name == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry friend, i got no conent :(",
			},
		})
		return
	}

	dbCalendar := db.Calendar{
		User: i.Member.User.ID,
		Name: name,
		Link: link,
	}

	if tx := c.db.Save(&dbCalendar); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("I will now watch your %q calendar", name),
		},
	})

	if err != nil {
		log.Println(err)
	}
}
