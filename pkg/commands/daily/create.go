package daily

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/meyskens/m-planner/pkg/db"
	"gorm.io/gorm/clause"

	"github.com/bwmarrin/discordgo"
)

func (d *DailyCommands) registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	desc := ""

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			desc = str
		}
	}

	if desc == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry friend, i got no conent :(",
			},
		})
		return
	}

	dbIdea := db.Daily{
		User:        i.Member.User.ID,
		Description: desc,
		ChannelID:   i.ChannelID,
		Reminders: []db.DailyReminder{ // for now we keep 2 days for week and weekend
			{
				Weekday: time.Monday,
				Hour:    0,
				Minute:  0,
			},
			{
				Weekday: time.Saturday,
				Hour:    0,
				Minute:  0,
			},
		},
	}

	if tx := d.db.Save(&dbIdea); tx.Error != nil {
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
			Content: fmt.Sprintf("I will remind you about %q every day now!", desc),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Configure",
							Style:    discordgo.SuccessButton,
							Disabled: false,
							Emoji: discordgo.ComponentEmoji{
								Name: "⚙️",
							},
							CustomID: fmt.Sprintf("change_daily--%d", dbIdea.ID),
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

func (id *DailyCommands) changeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	idStr := ""
	if strings.HasPrefix(i.MessageComponentData().CustomID, "change_daily--") { // when a button
		idStr = strings.Split(i.MessageComponentData().CustomID, "--")[1]
	} else if len(i.MessageComponentData().Values) > 0 && i.MessageComponentData().Values[0] != "" { // when in list
		idStr = i.MessageComponentData().Values[0]
	}

	if idStr == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "Crap... that got sent to the wrong handler",
			},
		})
	}

	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Crap... incorrect ID %q", err),
			},
		})
	}

	dbWhere := db.Daily{
		User: i.Member.User.ID,
	}

	dbWhere.ID = uint(idInt)

	dbDaily := db.Daily{}

	if tx := id.db.Preload(clause.Associations).Where(dbWhere).Find(&dbDaily); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}
	var wHour, wMin, weHour, weMin int
	for _, rem := range dbDaily.Reminders {
		if rem.Weekday == time.Monday {
			wHour = rem.Hour
			wMin = rem.Minute
		}

		if rem.Weekday == time.Saturday {
			weHour = rem.Hour
			weMin = rem.Minute
		}
	}

	annoying := "no"
	if dbDaily.Annoying {
		annoying = "yes"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_change_daily--" + idStr,
			Title:    "Edit a daily reminder",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "Reminder",
							Label:     "What is the reminder?",
							Style:     discordgo.TextInputShort,
							Value:     dbDaily.Description,
							Required:  true,
							MinLength: 1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "weekdays",
							Label:     "What time should I remind you on weekdays?",
							Style:     discordgo.TextInputShort,
							Value:     fmt.Sprintf("%02d:%02d", wHour, wMin),
							Required:  true,
							MaxLength: 5,
							MinLength: 1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "weekends",
							Label:     "What time should I remind you on weekends?",
							Value:     fmt.Sprintf("%02d:%02d", weHour, weMin),
							Style:     discordgo.TextInputShort,
							Required:  true,
							MaxLength: 5,
							MinLength: 1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							Label:     "Should I be annoying? (yes/no)",
							CustomID:  "annoying",
							MaxLength: 3,
							MinLength: 2,
							Required:  true,
							Style:     discordgo.TextInputShort,
							Value:     annoying,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
	}
}

func (d *DailyCommands) modalReturnCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	if !strings.HasPrefix(data.CustomID, "modal_change_daily--") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry i did not get any ID",
			},
		})
	}

	idStr := strings.Split(data.CustomID, "--")[1]
	idInt, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Crap... incorrect ID %q", err),
			},
		})
	}

	dbWhere := db.Daily{
		User: i.Member.User.ID,
	}

	dbWhere.ID = uint(idInt)

	dbDaily := db.Daily{}

	if tx := d.db.Preload(clause.Associations).Where(dbWhere).Find(&dbDaily); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	dbDaily.Description = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	weekTime := strings.Split(data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value, ":")
	weekendTime := strings.Split(data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value, ":")
	dbDaily.Annoying = strings.ToLower(data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value) == "yes"

	for j := range dbDaily.Reminders {
		if dbDaily.Reminders[j].Weekday == time.Monday {
			dbDaily.Reminders[j].Hour, _ = strconv.Atoi(weekTime[0])
			dbDaily.Reminders[j].Minute, _ = strconv.Atoi(weekTime[1])
		}

		if dbDaily.Reminders[j].Weekday == time.Saturday {
			dbDaily.Reminders[j].Hour, _ = strconv.Atoi(weekendTime[0])
			dbDaily.Reminders[j].Minute, _ = strconv.Atoi(weekendTime[1])
		}

		if tx := d.db.Save(&dbDaily.Reminders[j]); tx.Error != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
				},
			})
		}
	}

	if tx := d.db.Save(&dbDaily); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("I will remind you about %q every day now!", dbDaily.Description),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Configure",
							Style:    discordgo.SuccessButton,
							Disabled: false,
							Emoji: discordgo.ComponentEmoji{
								Name: "⚙️",
							},
							CustomID: fmt.Sprintf("change_daily--%d", dbDaily.ID),
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
