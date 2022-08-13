package calendar

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/embed"

	ical "github.com/arran4/golang-ical"
	"github.com/bwmarrin/discordgo"
)

type CalendarEvent struct {
	Name     string
	Start    time.Time
	End      time.Time
	Location string
}

func ParseSchedule(icalURL string) ([]CalendarEvent, error) {
	resp, err := http.Get(icalURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cal, err := ical.ParseCalendar(resp.Body)
	if err != nil {
		return nil, err
	}

	out := make([]CalendarEvent, 0)
	for _, icalEvent := range cal.Events() {
		fixICalTime(icalEvent)
		start, err := icalEvent.GetStartAt()
		if err != nil {
			log.Println(err)
			continue
		}
		end, err := icalEvent.GetEndAt()
		if err != nil {
			log.Println(err)
			continue
		}

		if end.After(time.Now()) && start.Before(time.Now().Add(time.Hour*24)) { // only get for 24 hours
			out = append(out, CalendarEvent{
				Name:     icalEvent.GetProperty(ical.ComponentPropertySummary).Value,
				Start:    start,
				End:      end,
				Location: icalEvent.GetProperty(ical.ComponentPropertyLocation).Value,
			})
		}
	}

	return out, nil
}

func (c *CalendarCommands) listCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Interaction.Type == discordgo.InteractionMessageComponent && strings.HasPrefix(i.MessageComponentData().CustomID, "calendar--") {
		startStr := strings.Split(i.MessageComponentData().CustomID, "calendar--")[1]
		start, _ := strconv.Atoi(startStr)
		c.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, start)
	}
	c.listCommandInternal(s, i, discordgo.InteractionResponseChannelMessageWithSource, 0)
}

func (c *CalendarCommands) listCommandInternal(s *discordgo.Session, i *discordgo.InteractionCreate, typeResponse discordgo.InteractionResponseType, start int) {
	cals := []db.Calendar{}
	if tx := c.db.Where(&db.Calendar{
		User: i.Member.User.ID,
	}).Find(&cals); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	if len(cals) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no calendars linked!",
			},
		})
		return
	}

	embeds := []*discordgo.MessageEmbed{}
	deleters := []discordgo.SelectMenuOption{}
	for _, cal := range cals {

		deleters = append(deleters, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("Delete %q", cal.Name),
			Value: fmt.Sprintf("%d", cal.ID),
		})

		events, err := ParseSchedule(cal.Link)
		if err != nil {
			e := embed.NewEmbed()
			e.SetTitle("Error")
			e.SetDescription(fmt.Sprintf("Error parsing calendar %q: %q", cal.Name, err))
			embeds = append(embeds, e.MessageEmbed)
			continue
		}

		for _, event := range events {
			e := embed.NewEmbed()
			e.AddField("Name", event.Name)
			e.AddField("Start", event.Start.Format("2006-01-02 15:04"))
			e.AddField("End", event.End.Format("2006-01-02 15:04"))
			e.AddField("Location", event.Location)
			e.AddField("Source", cal.Name)
			embeds = append(embeds, e.MessageEmbed)
		}

	}

	buttons := []discordgo.MessageComponent{}
	if len(embeds) > 10 {
		if start > 0 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Previous page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("calendar--%d", start-10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏮️",
				},
			})
		}
		if len(embeds) > start+10 {
			buttons = append(buttons, discordgo.Button{
				Label:    "Next page",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("calendar--%d", start+10),
				Emoji: discordgo.ComponentEmoji{
					Name: "⏭️",
				},
			})
		}

		embeds = embeds[start:]
		if len(embeds) > 10 {
			embeds = embeds[:10]
		}
	}

	comps := []discordgo.MessageComponent{}
	if len(deleters) > 0 {
		comps = append(comps, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					Placeholder: "Select a calendar to unlink",
					MaxValues:   1,
					Options:     deleters,
					CustomID:    "calendar_delete_list",
				},
			},
		})
	}
	if len(buttons) > 0 {
		comps = append(comps, discordgo.ActionsRow{Components: buttons})
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: typeResponse,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Content:    "Here are is your calendar for the next 24 hours",
			Components: comps,
		},
	})

	if err != nil {
		log.Println(err)
	}
}

func (c *CalendarCommands) listDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	dbCal := db.Calendar{
		User: i.Member.User.ID,
	}

	dbCal.ID = uint(idInt)

	if tx := c.db.Delete(&dbCal); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	c.listCommandInternal(s, i, discordgo.InteractionResponseUpdateMessage, 0)
}

func fixICalTime(icalEvent *ical.VEvent) {
	icalStart := icalEvent.GetProperty(ical.ComponentPropertyDtStart)
	if !strings.Contains(icalStart.Value, "Z") {
		icalStart.Value = icalStart.Value + "Z"
	}

	icalEvent.SetProperty(ical.ComponentPropertyDtStart, icalStart.Value)

	icalEnd := icalEvent.GetProperty(ical.ComponentPropertyDtEnd)
	if !strings.Contains(icalEnd.Value, "Z") {
		icalEnd.Value = icalEnd.Value + "Z"
	}
	icalEvent.SetProperty(ical.ComponentPropertyDtEnd, icalEnd.Value)
}
