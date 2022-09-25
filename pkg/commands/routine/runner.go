package routine

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/meyskens/m-planner/pkg/facts"
	printlib "github.com/meyskens/m-planner/pkg/print"
	"github.com/multiplay/go-cticker"
	"gorm.io/gorm/clause"
)

func (r *RoutineCommands) run(dg *discordgo.Session) {
	t := cticker.New(time.Minute, time.Second)
	for range t.C {
		go r.doRoutine(dg)
	}
}

func (r *RoutineCommands) doRoutine(dg *discordgo.Session) {
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

	r.remindEvents(dg, now, currentDay)

}

func (r *RoutineCommands) remindEvents(dg *discordgo.Session, now time.Time, currentDay time.Weekday) {
	routines := []db.Routine{}
	if err := r.db.Preload(clause.Associations).Find(&routines).Error; err != nil {
		log.Printf("[daily] error getting dailys: %s", err)
		return
	}

	currentRoutines := []db.Routine{}

	for _, routine := range routines {
		for _, reminder := range routine.Reminders {
			if currentDay == reminder.Weekday && now.Hour() == reminder.Hour && now.Minute() == reminder.Minute {
				currentRoutines = append(currentRoutines, routine)
			}
		}
	}

	for _, routine := range currentRoutines {

		if routine.Message {
			dg.ChannelMessageSendComplex(routine.ChannelID, &discordgo.MessageSend{
				Content: fmt.Sprintf("<@%s> %s", routine.User, routine.Description),
			})
		}

		if routine.Print {
			fact := ""
			if routine.FunFact {
				fact, _ = facts.FetchFunFact()
			}
			pd, err := printlib.PrintRoutine(routine.User, routine.Description, fact)
			if err != nil {
				log.Printf("error printing reminder: %s", err)
			}
			for _, pd := range pd {
				if err := r.db.Create(&pd).Error; err != nil {
					log.Printf("error saving print data: %s", err)
				}
			}
		}
	}

}
