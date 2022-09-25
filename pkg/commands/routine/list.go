package routine

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/embed"

	"github.com/bwmarrin/discordgo"
)

func (r *RoutineCommands) listCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Interaction.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "routines--") {
		startStr := strings.Split(i.MessageComponentData().CustomID, "routines--")[1]
		start, _ := strconv.Atoi(startStr)
		r.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, start)
	}
	r.listCommandInternal(s, i, discordgo.InteractionResponseChannelMessageWithSource, 0)
}

func (r *RoutineCommands) listCommandInternal(s *discordgo.Session, i *discordgo.InteractionCreate, typeResponse discordgo.InteractionResponseType, start int) {
	routines := []db.Routine{}
	if tx := r.db.Where(&db.Routine{
		User: i.Member.User.ID,
	}).Find(&routines); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	if len(routines) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Having no routine is sad... Poor you...",
			},
		})
		return
	}

	embeds := []*discordgo.MessageEmbed{}
	deleters := []discordgo.SelectMenuOption{}
	editors := []discordgo.SelectMenuOption{}
	for _, daily := range routines {
		e := embed.NewEmbed()
		e.AddField("Description", daily.Description)
		if daily.Message {
			e.AddField("Should I Discord you?", "Yes")
		} else {
			e.AddField("Should I Discord you?", "No")
		}
		if daily.Print {
			e.AddField("Should I print it?", "Yes")
		} else {
			e.AddField("Should I print it?", "No")
		}
		embeds = append(embeds, e.MessageEmbed)

		deleters = append(deleters, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("Delete %q", daily.Description),
			Value: fmt.Sprintf("%d", daily.ID),
		})

		editors = append(editors, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("Edit %q", daily.Description),
			Value: fmt.Sprintf("%d", daily.ID),
		})
	}

	buttons := []discordgo.MessageComponent{}
	if len(embeds) > 10 {
		if start > 0 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Previous page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("routines--%d", start-10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏮️",
				},
			})
		}
		if len(embeds) > start+10 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Next page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("routines--%d", start+10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏭️",
				},
			})
		}

		embeds = embeds[start:]
		deleters = deleters[start:]
		editors = editors[start:]
		if len(embeds) > 10 {
			embeds = embeds[:10]
			deleters = deleters[:10]
			editors = editors[:10]
		}
	}

	comps := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select a routine to delete",
					MaxValues:   1,
					Options:     deleters,
					CustomID:    "routine_delete_list",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select a routine to edit",
					MaxValues:   1,
					Options:     editors,
					CustomID:    "routine_edit_list",
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
			Content:    "Here are your daily routine reminders:",
			Components: comps,
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (r *RoutineCommands) listDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	dbRoutine := db.Routine{
		User: i.Member.User.ID,
	}

	dbRoutine.ID = uint(idInt)

	if tx := r.db.Delete(&dbRoutine); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	r.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, 0)
}
