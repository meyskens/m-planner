package recycle

import (
	"fmt"

	"github.com/meyskens/m-planner/pkg/db"

	"github.com/bwmarrin/discordgo"

	recyclebelgium "github.com/meyskens/go-recycle-belgium"
)

func (r *RecycleCommands) registerRecycle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	street := ""
	house := ""
	postalCode := ""
	annoying := false

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			street = str
		}
		if str, ok := i.ApplicationCommandData().Options[1].Value.(string); ok {
			house = str
		}
		if str, ok := i.ApplicationCommandData().Options[2].Value.(string); ok {
			postalCode = str
		}
		if b, ok := i.ApplicationCommandData().Options[3].Value.(bool); ok {
			annoying = b
		}
	}

	if r.xSecret == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry, I can't add a recycle reminder without an x-secret set",
			},
		})
		return
	}

	api := recyclebelgium.NewAPI(r.xSecret)

	zipResp, err := api.GetZipCodes(postalCode)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry, I can't add a recycle reminder for this postal code: %s", err),
			},
		})
		return
	}

	if len(zipResp.Items) < 1 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry, I can't add a recycle reminder for this postal code, no results found",
			},
		})
		return
	}

	zipID := zipResp.Items[0].ID // assuming an exact match was given

	cityResp, err := api.GetStreets(zipID, street)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry, I can't add a recycle reminder for this street: %s", err),
			},
		})
		return
	}
	if len(cityResp.Items) < 1 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry, I can't add a recycle reminder for this street, no results found",
			},
		})
		return
	}
	streetID := cityResp.Items[0].ID

	dbPlan := db.RecycleReminder{
		User:         i.Member.User.ID,
		ChannelID:    i.ChannelID,
		StreetID:     streetID,
		PostalCodeID: zipID,
		HouseNumber:  house,
		Annoying:     annoying,
	}

	r.db.Delete(&db.RecycleReminder{
		User:      i.Member.User.ID,
		ChannelID: i.ChannelID,
	}) // delete any old ones

	if tx := r.db.Save(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "I will remind you when to put the trash out!",
		},
	})
}
