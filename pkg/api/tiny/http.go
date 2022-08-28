package tiny

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/meyskens/m-planner/pkg/commands/calendar"
	"github.com/meyskens/m-planner/pkg/db"
	"gorm.io/gorm/clause"
)

// The Tiny API gives small data back, meant to be used in microcontrollers

type HTTPHandler struct {
	db *db.Connection
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) Register(e *echo.Echo, dbConn *db.Connection) {
	h.db = dbConn

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !strings.HasPrefix(c.Path(), "/tiny/") {
				return next(c)
			}

			token := c.Request().Header.Get("Authorization")

			if token == "" {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "did not get a valid token"})
			}

			dbToken := db.ApiToken{}
			if err := dbConn.Where("token = ?", token).First(&dbToken).Error; err != nil {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "did not get a valid token"})
			}

			c.Set("user", dbToken.User)

			if dbToken.User == "" {

				return c.JSON(http.StatusBadRequest, echo.Map{"error": "did not get a valid token"})
			}

			return next(c)
		}
	})

	e.GET("/tiny/tasks", h.getTasks)
	e.GET("/tiny/calendar", h.getCalendar)
	e.GET("/tiny/ideas/:channel", h.getIdeas)
	e.GET("/tiny/printjobs", h.getPrintjobs)
}

type TinyTask struct {
	Name      string `json:"name"`
	Time      string `json:"time"`
	IsInAlert bool   `json:"isInAlert"`
}

func (h *HTTPHandler) getTasks(c echo.Context) error {
	loc, _ := time.LoadLocation("Europe/Brussels")

	user := c.Get("user").(string)
	plans := []db.Plan{}
	if err := h.db.Where(&db.Plan{
		User: user,
	}).Find(&plans).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not get plans"})
	}
	dailys := []db.Daily{}
	if err := h.db.Preload(clause.Associations).Where(&db.Daily{
		User: user,
	}).Find(&dailys).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not get dailys"})
	}

	tasks := []TinyTask{}
	for _, plan := range plans {
		// if starts today and is in the future add it
		if plan.Start.After(time.Now().Truncate(24*time.Hour)) && plan.Start.Before(time.Now().Truncate(24*time.Hour).Add(time.Hour*24)) {
			tasks = append(tasks, TinyTask{
				Name:      plan.Description,
				Time:      plan.Start.In(loc).Format("15:04"),
				IsInAlert: !plan.SnoozedTill.IsZero(),
			})
		}
	}

	for _, daily := range dailys {
		// if starts today and is in the future add it
		for _, reminder := range daily.Reminders {
			reminders := []db.DailyReminderEvent{}
			h.db.Where(&db.DailyReminderEvent{
				DailyID: daily.ID,
			}).Find(&reminders)

			if reminder.Weekday != getCurrentDay() {
				continue
			}

			if reminder.Hour >= time.Now().In(loc).Hour() || len(reminders) > 0 {
				if reminder.Hour == time.Now().In(loc).Hour() && reminder.Minute < time.Now().In(loc).Minute() && len(reminders) == 0 {
					continue
				}

				tasks = append(tasks, TinyTask{
					Name:      daily.Description,
					Time:      fmt.Sprintf("%02d:%02d", reminder.Hour, reminder.Minute),
					IsInAlert: len(reminders) > 0,
				})
			}
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Time < tasks[j].Time
	})

	// limit to limit parameter
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}

	return c.JSON(http.StatusOK, tasks)
}

func getCurrentDay() time.Weekday {
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

	return currentDay
}

type TinyCalendarEvent struct {
	Name     string `json:"name"`
	Start    string `json:"start"`
	Location string `json:"location"`
}

func (h *HTTPHandler) getCalendar(c echo.Context) error {
	cals := []db.Calendar{}
	if err := h.db.Preload(clause.Associations).Where(&db.Calendar{
		User: c.Get("user").(string),
	}).Find(&cals).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not get calendars"})
	}

	e := []TinyCalendarEvent{}

	for _, cal := range cals {
		events, err := calendar.ParseSchedule(cal.Link)
		if err != nil {
			continue
		}
		for _, event := range events {
			e = append(e, TinyCalendarEvent{
				Name:     event.Name,
				Start:    event.Start.Format("15:04"),
				Location: event.Location,
			})
		}
	}

	sort.Slice(e, func(i, j int) bool {
		return e[i].Start < e[j].Start
	})

	// limit to limit parameter
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit > 0 && len(e) > limit {
		e = e[:limit]
	}

	return c.JSON(http.StatusOK, e)
}

type TinyIdea struct {
	Name string `json:"name"`
}

func (h *HTTPHandler) getIdeas(c echo.Context) error {
	ideas := []db.Idea{}
	if err := h.db.Preload(clause.Associations).Where(&db.Idea{
		User:      c.Get("user").(string),
		ChannelID: c.Param("channel"),
	}).Find(&ideas).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not get ideas"})
	}

	i := []TinyIdea{}

	for _, idea := range ideas {
		i = append(i, TinyIdea{
			Name: idea.Description,
		})
	}

	// limit to limit parameter
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit > 0 && len(i) > limit {
		i = i[:limit]
	}

	return c.JSON(http.StatusOK, i)
}

func (h *HTTPHandler) getPrintjobs(c echo.Context) error {
	pj := []db.PrintJob{}
	if err := h.db.Preload(clause.Associations).Where(&db.PrintJob{
		User: c.Get("user").(string),
	}).Find(&pj).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not get printjobs"})
	}

	for _, job := range pj {
		h.db.Delete(&job) // printed so we delete them
	}

	return c.JSON(http.StatusOK, echo.Map{"data": pj})
}
