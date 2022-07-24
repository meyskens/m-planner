package ideas

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/embed"

	"github.com/bwmarrin/discordgo"
)

func (id *IdeasCommands) listCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Interaction.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "ideas--") {
		startStr := strings.Split(i.MessageComponentData().CustomID, "ideas--")[1]
		start, _ := strconv.Atoi(startStr)
		id.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, start)
	}
	id.listCommandInternal(s, i, discordgo.InteractionResponseChannelMessageWithSource, 0)
}

func (id *IdeasCommands) listCommandInternal(s *discordgo.Session, i *discordgo.InteractionCreate, typeResponse discordgo.InteractionResponseType, start int) {
	ideas := []db.Idea{}
	if tx := id.db.Where(&db.Idea{
		User: i.Member.User.ID,
	}).Find(&ideas); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	if len(ideas) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no ideas left, go outside and think of one!",
			},
		})
		return
	}

	embeds := []*discordgo.MessageEmbed{}
	deleters := []discordgo.SelectMenuOption{}
	for _, idea := range ideas {
		e := embed.NewEmbed()
		e.AddField("Description", idea.Description)
		embeds = append(embeds, e.MessageEmbed)

		deleters = append(deleters, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("Delete %q", idea.Description),
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
					Placeholder: "Select an idea to delete",
					MaxValues:   1,
					Options:     deleters,
					CustomID:    "idea_delete_list",
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
			Content:    "Here are your ideas",
			Components: comps,
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (id *IdeasCommands) listDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	id.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, 0)
}
