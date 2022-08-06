package recycle

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	recyclebelgium "github.com/meyskens/go-recycle-belgium"
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/multiplay/go-cticker"
)

func (r *RecycleCommands) run(dg *discordgo.Session) {
	t := cticker.New(time.Hour, time.Second)
	for range t.C {
		go r.doRoutine(dg)
	}
}

func (r *RecycleCommands) doRoutine(dg *discordgo.Session) {
	// get Europe/Brussels time
	loc, _ := time.LoadLocation("Europe/Brussels")
	r.createRemindEvents(time.Now().In(loc), dg)
}

func (r *RecycleCommands) createRemindEvents(now time.Time, dg *discordgo.Session) {

	recycles := []db.RecycleReminder{}
	if err := r.db.Find(&recycles).Error; err != nil {
		log.Printf("[recycle] error getting recycle plans: %s", err)
		return
	}

	for _, plan := range recycles {
		if time.Now().After(plan.LastRun.Add(24 * time.Hour).Truncate(24 * time.Hour)) {
			api := recyclebelgium.NewAPI(r.xSecret)

			collections, _ := api.GetCollections(plan.PostalCodeID, plan.StreetID, plan.HouseNumber, time.Now().Add(24*time.Hour), time.Now().Add(24*time.Hour), 100)
			if len(collections.Items) < 1 {
				continue
			}

			collectionsPerDay := make(map[string]string)
			for _, collection := range collections.Items {
				time := collection.Timestamp.Format("2006-01-02")
				if _, ok := collectionsPerDay[time]; !ok {
					collectionsPerDay[time] = ""
				}
				collectionsPerDay[time] += fmt.Sprintf(", %s", collection.Fraction.Name.NL)
			}

			for date, collections := range collectionsPerDay {
				dbPlan := db.Plan{
					User:        plan.User,
					ChannelID:   plan.ChannelID,
					Annoying:    plan.Annoying,
					Description: fmt.Sprintf("%s will be collected at %s", strings.TrimLeft(collections, ", "), date),
					Start:       time.Now().Truncate(24 * time.Hour).Add(16 * time.Hour),
				}
				if err := r.db.Create(&dbPlan).Error; err != nil {
					log.Printf("[recycle] error creating plan: %s", err)
					continue
				}
			}
		}
	}
}
