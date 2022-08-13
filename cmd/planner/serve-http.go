package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewServeHTTPCmd())
}

type serveHTTPCmdOptions struct {
	BindAddr string
	Port     int

	db *db.Connection

	postgresHost     string
	postgresPort     int
	postgresUsername string
	postgresDatabase string
	postgresPassword string
	enableSSL        bool
}

// NewServeHTTPCmd generates the `serve-http` command
func NewServeHTTPCmd() *cobra.Command {
	s := serveHTTPCmdOptions{}
	c := &cobra.Command{
		Use:     "serve",
		Short:   "Run the HTTP server",
		Long:    `This runs the HTTP server for the REST api`,
		RunE:    s.RunE,
		PreRunE: s.Validate,
	}

	c.Flags().StringVarP(&s.BindAddr, "bind-address", "b", "0.0.0.0", "address to bind port to")
	c.Flags().IntVarP(&s.Port, "port", "p", 8080, "Port to listen on")

	c.Flags().StringVar(&s.postgresHost, "postgres-host", "", "PostgreSQL hostname")
	c.Flags().IntVar(&s.postgresPort, "postgres-port", 5432, "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresUsername, "postgres-username", "", "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresPassword, "postgres-password", "", "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresDatabase, "postgres-database", "", "PostgreSQL hostname")
	c.Flags().BoolVar(&s.enableSSL, "postgres-enable-ssl", false, "PostgreSQL to use TLS")

	c.MarkFlagRequired("postgres-host")
	c.MarkFlagRequired("postgres-username")
	c.MarkFlagRequired("postgres-password")
	c.MarkFlagRequired("postgres-database")

	return c
}

func (s *serveHTTPCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (s *serveHTTPCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())

	var err error
	s.db, err = db.NewConnection(db.ConnectionDetails{
		Host:      s.postgresHost,
		Port:      s.postgresPort,
		User:      s.postgresUsername,
		Password:  s.postgresPassword,
		Database:  s.postgresDatabase,
		EnableSSL: s.enableSSL,
	})
	if err != nil {
		return fmt.Errorf("error connecting to DB %w", err)
	}

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		ExposeHeaders: []string{echo.HeaderContentType, echo.HeaderAccept, "Num-Total-Entries"},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "M-Planner API endpoint")
	})

	go func() {
		e.Start(fmt.Sprintf("%s:%d", s.BindAddr, s.Port))
		cancel() // server ended, stop the world
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			return nil
		}
	}
}
