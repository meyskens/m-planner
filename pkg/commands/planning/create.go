package planning

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/tj/go-naturaldate"
	"gorm.io/gorm/clause"

	"github.com/bwmarrin/discordgo"
)

func (p *PlanningCommands) registerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	desc := ""
	timeStr := ""
	annoying := false

	if len(i.ApplicationCommandData().Options) > 0 {
		if str, ok := i.ApplicationCommandData().Options[0].Value.(string); ok {
			desc = str
		}
		if str, ok := i.ApplicationCommandData().Options[1].Value.(string); ok {
			timeStr = str
		}
		if b, ok := i.ApplicationCommandData().Options[2].Value.(bool); ok {
			annoying = b
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

	t := parseTime(timeStr)

	dbPlan := db.Plan{
		User:        i.Member.User.ID,
		Description: desc,
		ChannelID:   i.ChannelID,
		Annoying:    annoying,
		Start:       t,
	}

	if tx := p.db.Save(&dbPlan); tx.Error != nil {
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
			Content: fmt.Sprintf("I will remind you about %q at %s", desc, t.Format(time.RFC850)),
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
							CustomID: fmt.Sprintf("change_planning--%d", dbPlan.ID),
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

func (p *PlanningCommands) changeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	idStr := ""
	if strings.HasPrefix(i.MessageComponentData().CustomID, "change_planning--") { // when a button
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

	dbWhere := db.Plan{
		User: i.Member.User.ID,
	}

	dbWhere.ID = uint(idInt)

	dbPlan := db.Plan{}

	if tx := p.db.Where(dbWhere).Find(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	annoying := "no"
	if dbPlan.Annoying {
		annoying = "yes"
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_change_planning--" + idStr,
			Title:    "Edit a reminder",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "Reminder",
							Label:     "What is the reminder?",
							Style:     discordgo.TextInputShort,
							Value:     dbPlan.Description,
							Required:  true,
							MinLength: 1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "Time",
							Label:     "What time should I remind you?",
							Style:     discordgo.TextInputShort,
							Value:     dbPlan.Start.Format("2006-01-02 15:04"),
							Required:  true,
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

func (p *PlanningCommands) modalReturnCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	if !strings.HasPrefix(data.CustomID, "modal_change_planning--") {
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

	dbWhere := db.Plan{
		User: i.Member.User.ID,
	}

	dbWhere.ID = uint(idInt)

	dbPlan := db.Plan{}

	if tx := p.db.Preload(clause.Associations).Where(dbWhere).Find(&dbPlan); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	dbPlan.Description = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	dbPlan.Start = parseTime(data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value)
	dbPlan.Annoying = strings.ToLower(data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value) == "yes"

	if tx := p.db.Save(&dbPlan); tx.Error != nil {
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
			Content: fmt.Sprintf("I will remind you about %q at %s", dbPlan.Description, dbPlan.Start.Format(time.RFC850)),
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
							CustomID: fmt.Sprintf("change_planning--%d", dbPlan.ID),
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

func parseTime(timeStr string) time.Time {
	// is it from the edit?
	t, err := time.Parse("2006-01-02 15:04", timeStr)
	if err == nil && !t.IsZero() {
		return t
	}
	// is it RFC850?
	t, err = time.Parse(time.RFC850, timeStr)
	if err == nil && !t.IsZero() {
		return t
	}
	brussesls := time.FixedZone("Europe/Brussels", 0)
	t, _ = naturaldate.Parse(timeStr, time.Now().In(brussesls), naturaldate.WithDirection(naturaldate.Future))

	return t.In(brussesls)
}
