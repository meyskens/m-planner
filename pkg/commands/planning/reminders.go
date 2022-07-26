package planning

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/db"
)

func (p *PlanningCommands) markPlanningComplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !strings.HasPrefix(i.MessageComponentData().CustomID, "mark_planning_complete--") {
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

	dbPlan := db.Plan{}
	dbPlan.ID = uint(idInt)

	p.db.Delete(&dbPlan)

	if dbPlan.User == "161504618017325057" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Good ~~girl~~! You deserve a hug 🧸",
				Components: []discordgo.MessageComponent{},
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Good girl! You deserve a hug 🧸",
				Components: []discordgo.MessageComponent{},
			},
		})
	}

}

func (id *PlanningCommands) snoozePlanning(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !strings.HasPrefix(i.MessageComponentData().CustomID, "snooze_planning--") {
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

	if len(i.MessageComponentData().Values) < 1 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry buddy, did not get a value",
			},
		})
	}
	value := i.MessageComponentData().Values[0]
	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Crap... incorrect value %q", err),
			},
		})
		return
	}

	dbPlan := db.Plan{}
	dbPlan.ID = uint(idInt)

	if tx := id.db.Find(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
		return
	}

	dbPlan.SnoozedTill = dbPlan.SnoozedTill.Add(time.Duration(valueInt) * time.Minute)

	if tx := id.db.Save(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("Sorry to bother you, I will be quiet for %d minutes", valueInt),
			Components: []discordgo.MessageComponent{},
		},
	})

	if err != nil {
		log.Panicln(err)
	}
}
