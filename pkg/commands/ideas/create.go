package ideas

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/meyskens/m-planner/pkg/db"

	"github.com/bwmarrin/discordgo"
)

func (id *IdeasCommands) registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	idea := ""

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			idea = str
		}
	}

	if idea == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry friend, i got no conent :(",
			},
		})
		return
	}

	dbIdea := db.Idea{
		User:        i.Member.User.ID,
		Description: idea,
	}

	if tx := id.db.Save(&dbIdea); tx.Error != nil {
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
			Content: fmt.Sprintf("I will remind you about %q", idea),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Delete",
							Style:    discordgo.DangerButton,
							Disabled: false,
							Emoji: discordgo.ComponentEmoji{
								Name: "🗑️",
							},
							CustomID: fmt.Sprintf("delete_idea--%d", dbIdea.ID),
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (id *IdeasCommands) deleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if !strings.HasPrefix(i.MessageComponentData().CustomID, "delete_idea--") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "Crap... that got sent to the wrong handler",
			},
		})
		return
	}

	idStr := strings.Split(i.MessageComponentData().CustomID, "--")[1]
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Crap... incorrect ID %q", err),
			},
		})
	}

	dbIdea := db.Idea{
		User: i.Member.User.ID,
	}

	dbIdea.ID = uint(idInt)

	if tx := id.db.Delete(&dbIdea); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Deleted!",
		},
	})
}
