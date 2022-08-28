package groceries

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/meyskens/m-planner/pkg/db"

	"github.com/bwmarrin/discordgo"
)

func (id *GroceriesCommands) registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	item := ""

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			item = str
		}
	}

	if item == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry friend, i got no conent :(",
			},
		})
		return
	}

	dbGrocery := db.Grocery{
		ChannelID: i.ChannelID,
		Item:      item,
	}

	if tx := id.db.Save(&dbGrocery); tx.Error != nil {
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
			Content: fmt.Sprintf("I put %q on your list!", item),
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
							CustomID: fmt.Sprintf("delete_grocery--%d", dbGrocery.ID),
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

func (id *GroceriesCommands) deleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if !strings.HasPrefix(i.MessageComponentData().CustomID, "delete_grocery--") {
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
	if err != nil && idStr != "all" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Crap... incorrect ID %q", err),
			},
		})
	}

	dbGrocery := db.Grocery{
		ChannelID: i.ChannelID,
	}

	if idStr != "all" {
		dbGrocery.ID = uint(idInt)
	}

	if tx := id.db.Where(&dbGrocery).Delete(&dbGrocery); tx.Error != nil {
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
			Content:    "Deleted!",
			Components: []discordgo.MessageComponent{},
		},
	})
}
