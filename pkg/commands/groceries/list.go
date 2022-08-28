package groceries

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/embed"
	printlib "github.com/meyskens/m-planner/pkg/print"

	"github.com/bwmarrin/discordgo"
)

func (id *GroceriesCommands) listCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Interaction.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "groceries--") {
		startStr := strings.Split(i.MessageComponentData().CustomID, "groceries--")[1]
		start, _ := strconv.Atoi(startStr)
		id.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, start)
	}
	id.listCommandInternal(s, i, discordgo.InteractionResponseChannelMessageWithSource, 0)
}

func (id *GroceriesCommands) listCommandInternal(s *discordgo.Session, i *discordgo.InteractionCreate, typeResponse discordgo.InteractionResponseType, start int) {
	print := false

	if i.Type == discordgo.InteractionApplicationCommand && len(i.ApplicationCommandData().Options) > 0 {
		if v, ok := i.ApplicationCommandData().Options[0].Value.(bool); ok {
			print = v
		}
	}

	groceries := []db.Grocery{}
	if tx := id.db.Where(&db.Grocery{
		ChannelID: i.ChannelID,
	}).Find(&groceries); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	if len(groceries) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no items you need right now! You can stay home if you want :)",
			},
		})
		return
	}

	if print {
		pd, err := printlib.PrintGroceriesList(i.Member.User.ID, groceries)
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
			if err := id.db.Create(&pd).Error; err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", err),
					},
				})
				return
			}
		}
	}

	embeds := []*discordgo.MessageEmbed{}
	deleters := []discordgo.SelectMenuOption{}
	for _, idea := range groceries {
		e := embed.NewEmbed()
		e.AddField("Item", idea.Item)
		embeds = append(embeds, e.MessageEmbed)

		deleters = append(deleters, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("Delete %q", idea.Item),
			Value: fmt.Sprintf("%d", idea.ID),
		})
	}

	buttons := []discordgo.MessageComponent{}
	if len(embeds) > 10 {
		if start > 0 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Previous page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("ideas--%d", start-10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏮️",
				},
			})
		}
		if len(embeds) > start+10 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Next page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("ideas--%d", start+10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏭️",
				},
			})
		}

		embeds = embeds[start:]
		deleters = deleters[start:]
		if len(embeds) > 10 {
			embeds = embeds[:10]
			deleters = deleters[:10]
		}
	}

	comps := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select an item to delete",
					MaxValues:   1,
					Options:     deleters,
					CustomID:    "grocery_delete_list",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Delete all",
					Style:    discordgo.DangerButton,
					Disabled: false,
					Emoji: discordgo.ComponentEmoji{
						Name: "🗑️",
					},
					CustomID: "delete_grocery--all",
				},
			},
		},
	}
	if len(buttons) > 0 {
		comps = append(comps, discordgo.ActionsRow{Components: buttons})
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: typeResponse,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Content:    "Here is your shopping list",
			Components: comps,
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (id *GroceriesCommands) listDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if len(i.MessageComponentData().Values) < 1 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry buddy, did not get a value",
			},
		})
	}
	idStr := i.MessageComponentData().Values[0]
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
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

	dbGrocery.ID = uint(idInt)

	if tx := id.db.Delete(&dbGrocery); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	id.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, 0)
}
