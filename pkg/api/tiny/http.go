package tiny

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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

}

type TinyTask struct {
	Name      string `json:"name"`
	Time      string `json:"time"`
	IsInAlert bool   `json:"isInAlert"`
}

func (h *HTTPHandler) getTasks(c echo.Context) error {
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
		if plan.Start.After(time.Now()) && plan.Start.Before(time.Now().Truncate(24*time.Hour).Add(time.Hour*24)) {
			tasks = append(tasks, TinyTask{
				Name:      plan.Description,
				Time:      plan.Start.Format("15:04"),
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

			if reminder.Hour >= time.Now().Hour() || len(reminders) > 0 {
				if reminder.Hour == time.Now().Hour() && reminder.Minute < time.Now().Minute() && len(reminders) == 0 {
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

	return c.JSON(http.StatusOK, tasks)
}

func getCurrentDay() time.Weekday {
	now := time.Now()
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
