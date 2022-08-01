package planning

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/embed"

	"github.com/bwmarrin/discordgo"
)

func (p *PlanningCommands) listCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Interaction.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "planning--") {
		startStr := strings.Split(i.MessageComponentData().CustomID, "planning--")[1]
		start, _ := strconv.Atoi(startStr)
		p.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, start)
	}
	p.listCommandInternal(s, i, discordgo.InteractionResponseChannelMessageWithSource, 0)
}

func (p *PlanningCommands) listCommandInternal(s *discordgo.Session, i *discordgo.InteractionCreate, typeResponse discordgo.InteractionResponseType, start int) {
	plans := []db.Plan{}
	if tx := p.db.Where(&db.Plan{
		User: i.Member.User.ID,
	}).Find(&plans); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	if len(plans) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no planned items. Weird... enjoy a empty day I guess...",
			},
		})
		return
	}

	loc, _ := time.LoadLocation("Europe/Brussels")

	embeds := []*discordgo.MessageEmbed{}
	deleters := []discordgo.SelectMenuOption{}
	editors := []discordgo.SelectMenuOption{}
	for _, daily := range plans {
		e := embed.NewEmbed()
		e.AddField("Description", daily.Description)
		e.AddField("When", daily.Start.In(loc).Format(time.RFC850))
		if daily.Annoying {
			e.AddField("Should I annoy you?", "YEEES")
		} else {
			e.AddField("Should I annoy you?", "No")
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
				CustomID: fmt.Sprintf("planning--%d", start-10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏮️",
				},
			})
		}
		if len(embeds) > start+10 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Next page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("planning--%d", start+10),
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
					Placeholder: "Select a plan to delete",
					MaxValues:   1,
					Options:     deleters,
					CustomID:    "planning_delete_list",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select a plan to edit",
					MaxValues:   1,
					Options:     editors,
					CustomID:    "planning_edit_list",
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
			Content:    "Here are your reminders:",
			Components: comps,
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (p *PlanningCommands) listDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	dbPlan := db.Plan{
		User: i.Member.User.ID,
	}

	dbPlan.ID = uint(idInt)

	if tx := p.db.Delete(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	p.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, 0)
}
