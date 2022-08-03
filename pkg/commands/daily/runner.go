package daily

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/multiplay/go-cticker"
	"gorm.io/gorm/clause"
)

func (d *DailyCommands) run(dg *discordgo.Session) {
	t := cticker.New(time.Minute, time.Second)
	for range t.C {
		go d.doRoutine(dg)
	}
}

func (d *DailyCommands) doRoutine(dg *discordgo.Session) {
	// get Europe/Brussels time
	loc, _ := time.LoadLocation("Europe/Brussels")

	now := time.Now().In(loc)
	currentDay := now.Weekday()
	// correct to use monday or saturday
	switch currentDay {
	case time.Saturday:
		fallthrough
	case time.Sunday:
		currentDay = time.Saturday
	default:
		currentDay = time.Monday
	}

	d.createEvents(now, currentDay)
	d.remindEvents(dg)

}

func (d *DailyCommands) createEvents(now time.Time, currentDay time.Weekday) {

	dailys := []db.Daily{}
	if err := d.db.Preload(clause.Associations).Find(&dailys).Error; err != nil {
		log.Printf("[daily] error getting dailys: %s", err)
		return
	}

	for _, daily := range dailys {
		for _, reminder := range daily.Reminders {
			if currentDay == reminder.Weekday && now.Hour() == reminder.Hour && now.Minute() == reminder.Minute {
				if tx := d.db.Save(&db.DailyReminderEvent{
					DailyID: daily.ID,
					Start:   now,
				}); tx.Error != nil {
					log.Printf("error saving daily reminder event: %s", tx.Error)
				}
			}
		}
	}

}

func (d *DailyCommands) remindEvents(dg *discordgo.Session) {

	events := []db.DailyReminderEvent{}
	if err := d.db.Preload(clause.Associations).Find(&events).Error; err != nil {
		log.Printf("[daily] error getting dailys: %s", err)
		return
	}

	for _, event := range events {
		if time.Now().After(event.Start) {
			snoozer := discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    fmt.Sprintf("snooze_event--%d", event.ID),
						MaxValues:   1,
						Placeholder: "Snooze this for...",
						Options: []discordgo.SelectMenuOption{
							{
								Label: "Snooze for 20 minutes",
								Value: "20",
							},
							{
								Label: "Snooze for 30 minutes",
								Value: "30",
							},
							{
								Label: "Snooze for 45 minutes",
								Value: "45",
							},
							{
								Label: "Snooze for 60 minutes",
								Value: "60",
							},
							{
								Label: "Snooze for 2 hours",
								Value: "120",
							},
						},
					},
				},
			}

			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "I did it!",
							Style: discordgo.SuccessButton,
							Emoji: discordgo.ComponentEmoji{
								Name: "💖",
							},
							CustomID: fmt.Sprintf("mark_event_complete--%d", event.ID),
						},
					},
				},
			}
			if event.Daily.Annoying {
				components = append(components, snoozer)
			}
			msg, err := dg.ChannelMessageSendComplex(event.Daily.ChannelID, &discordgo.MessageSend{
				Content:    fmt.Sprintf("<@%s> don't forget to %s", event.Daily.User, event.Daily.Description),
				Components: components,
			})

			if err != nil {
				fmt.Println(err)
			} else {
				if tx := d.db.Save(&db.SentMessage{
					DailyReminderEventID: event.ID,
					MessageID:            msg.ID,
				}); tx.Error != nil {
					log.Printf("error saving sent message ID: %s", tx.Error)
				}
			}

			if !event.Daily.Annoying {
				if tx := d.db.Delete(&event); tx.Error != nil {
					log.Printf("error deleting daily reminder event: %s", tx.Error)
				}
			} else {
				event.SnoozedTill = time.Now().Truncate(time.Minute).Add(10 * time.Minute)
				if tx := d.db.Save(&event); tx.Error != nil {
					log.Printf("error deleting daily reminder event: %s", tx.Error)
				}

				if time.Now().Add(-8 * time.Hour).After(event.Start) {
					dg.ChannelMessageSend(event.Daily.ChannelID, fmt.Sprintf("<@%s> i have been trying to get you to %q for 8 hours now... FUCK OFF", event.Daily.User, event.Daily.Description))
				} else if time.Now().Add(-4 * time.Hour).After(event.Start) {
					dg.ChannelMessageSend(event.Daily.ChannelID, fmt.Sprintf("<@%s> i have been trying to get you to %q for 4 hours now... please...", event.Daily.User, event.Daily.Description))
				}
			}
		}
	}

}
