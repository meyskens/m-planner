package daily

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/db"
	"gorm.io/gorm/clause"
)

func (d *DailyCommands) markEventComplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !strings.HasPrefix(i.MessageComponentData().CustomID, "mark_event_complete--") {
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

	dbEvent := db.DailyReminderEvent{}

	d.db.Preload(clause.Associations).Where("id = ?", idInt).Find(&dbEvent)

	if tx := d.db.Delete(&dbEvent); tx.Error != nil {
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
			Content:    "You dit it! I am so proud of you 🥹",
			Components: []discordgo.MessageComponent{},
		},
	})

	go func() {
		for _, msg := range dbEvent.SentMessages {
			if msg.MessageID != i.Message.ID {
				s.ChannelMessageDelete(i.ChannelID, msg.MessageID)
			}
			d.db.Unscoped().Delete(&msg)
			time.Sleep(time.Millisecond * 400)
		}
	}()
}

func (id *DailyCommands) snoozeEvent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !strings.HasPrefix(i.MessageComponentData().CustomID, "snooze_event--") {
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

	dbEvent := db.DailyReminderEvent{}
	dbEvent.ID = uint(idInt)

	if tx := id.db.Find(&dbEvent); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
		return
	}

	dbEvent.SnoozedTill = dbEvent.SnoozedTill.Add(time.Duration(valueInt) * time.Minute)

	if tx := id.db.Save(&dbEvent); tx.Error != nil {
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
