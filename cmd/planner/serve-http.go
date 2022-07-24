package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewServeHTTPCmd())
}

type serveHTTPCmdOptions struct {
	BindAddr string
	Port     int
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

	return c
}

func (s *serveHTTPCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (s *serveHTTPCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())

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
