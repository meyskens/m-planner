package planning

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/db"
	printlib "github.com/meyskens/m-planner/pkg/print"
	"github.com/multiplay/go-cticker"
)

func (p *PlanningCommands) run(dg *discordgo.Session) {
	t := cticker.New(time.Minute, time.Second)
	for range t.C {
		go p.doRoutine(dg)
	}
}

func (p *PlanningCommands) doRoutine(dg *discordgo.Session) {
	// get Europe/Brussels time
	loc, _ := time.LoadLocation("Europe/Brussels")
	p.remindEvents(time.Now().In(loc), dg)
}

func (p *PlanningCommands) remindEvents(now time.Time, dg *discordgo.Session) {

	plans := []db.Plan{}
	if err := p.db.Find(&plans).Error; err != nil {
		log.Printf("[planning] error getting plans: %s", err)
		return
	}

	for _, plan := range plans {
		if !plan.Start.IsZero() && time.Now().After(plan.SnoozedTill) && time.Now().After(plan.Start) {
			snoozer := discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    fmt.Sprintf("snooze_planning--%d", plan.ID),
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
							CustomID: fmt.Sprintf("mark_planning_complete--%d", plan.ID),
						},
					},
				},
			}
			if plan.Annoying {
				components = append(components, snoozer)
			}
			_, err := dg.ChannelMessageSendComplex(plan.ChannelID, &discordgo.MessageSend{
				Content:    fmt.Sprintf("<@%s> don't forget to %s", plan.User, plan.Description),
				Components: components,
			})

			if err != nil {
				fmt.Println(err)
			}

			if plan.Print && plan.SnoozedTill.IsZero() {
				pd, err := printlib.PrintReminder(plan.User, fmt.Sprintf("Don't forget to %s", plan.Description))
				if err != nil {
					log.Printf("error printing reminder: %s", err)
				}
				for _, pd := range pd {
					if err := d.db.Create(&pd).Error; err != nil {
						log.Printf("error saving print data: %s", err)
					}
				}
			}

			if !plan.Annoying {
				if tx := p.db.Delete(&plan); tx.Error != nil {
					log.Printf("error deleting plan: %s", tx.Error)
				}
			} else {
				plan.SnoozedTill = time.Now().Truncate(time.Minute).Add(10 * time.Minute)
				if tx := p.db.Save(&plan); tx.Error != nil {
					log.Printf("error deleting daily reminder event: %s", tx.Error)
				}

				if time.Now().Add(-8 * time.Hour).After(plan.Start) {
					dg.ChannelMessageSend(plan.ChannelID, fmt.Sprintf("<@%s> i have been trying to get you to %q for 8 hours now... FUCK OFF", plan.User, plan.Description))
				}
			}
		}
	}
}
