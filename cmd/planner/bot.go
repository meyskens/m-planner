package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/meyskens/m-planner/pkg/command"
	"github.com/meyskens/m-planner/pkg/commands/daily"
	"github.com/meyskens/m-planner/pkg/commands/ideas"
	"github.com/meyskens/m-planner/pkg/commands/planning"
	"github.com/meyskens/m-planner/pkg/commands/recycle"
	"github.com/meyskens/m-planner/pkg/commands/token"
	"github.com/meyskens/m-planner/pkg/db"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewServeCmd())
}

type serveCmdOptions struct {
	Token    string
	dg       *discordgo.Session
	handlers []command.Interface
	db       *db.Connection

	postgresHost     string
	postgresPort     int
	postgresUsername string
	postgresDatabase string
	postgresPassword string
	enableSSL        bool

	recycleSecret string

	onInteractionCreateHandler map[string][]func(*discordgo.Session, *discordgo.InteractionCreate)
}

// NewServeCmd generates the `serve` command
func NewServeCmd() *cobra.Command {
	s := serveCmdOptions{}
	c := &cobra.Command{
		Use:     "bot",
		Short:   "Run the Discord Bot",
		Long:    `This connects to Discord and handle all events`,
		RunE:    s.RunE,
		PreRunE: s.Validate,
	}

	c.Flags().StringVar(&s.Token, "token", "", "Discord Bot Token")
	c.MarkFlagRequired("token")

	c.Flags().StringVar(&s.postgresHost, "postgres-host", "", "PostgreSQL hostname")
	c.Flags().IntVar(&s.postgresPort, "postgres-port", 5432, "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresUsername, "postgres-username", "", "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresPassword, "postgres-password", "", "PostgreSQL hostname")
	c.Flags().StringVar(&s.postgresDatabase, "postgres-database", "", "PostgreSQL hostname")
	c.Flags().BoolVar(&s.enableSSL, "postgres-enable-ssl", false, "PostgreSQL to use TLS")
	c.Flags().StringVar(&s.recycleSecret, "recycle-secret", "", "Recycle secret")
	c.MarkFlagRequired("jwt-secret")
	c.MarkFlagRequired("postgres-host")
	c.MarkFlagRequired("postgres-username")
	c.MarkFlagRequired("postgres-password")
	c.MarkFlagRequired("postgres-database")

	return c
}

func (s *serveCmdOptions) Validate(cmd *cobra.Command, args []string) error {
	if s.Token == "" {
		return errors.New("no token specified")
	}

	return nil
}

func (s *serveCmdOptions) RunE(cmd *cobra.Command, args []string) error {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	err = s.db.DoMigrate()
	if err != nil {
		return fmt.Errorf("error migrating the DB %w", err)
	}

	dg, err := discordgo.New("Bot " + s.Token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	dg.State.TrackVoice = true

	err = dg.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	defer dg.Close()
	s.dg = dg

	s.RegisterHandlers()

	s.dg.AddHandler(s.onInteractionCreate)

	for _, handler := range s.handlers {
		err := handler.InstallSlashCommands(s.dg)
		if err != nil {
			log.Println("error installing slash commandos", err)
		}
	}

	go func() {
		start := time.Now()
		for {
			dg.UpdateGameStatus(0, fmt.Sprintf("Been hugging a Blahaj for %d minutes", time.Since(start)/time.Minute))
			time.Sleep(time.Minute)
		}
	}()

	log.Println("M-Planner Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}

func (s *serveCmdOptions) RegisterHandlers() {
	s.handlers = []command.Interface{
		ideas.NewCommands(s.db),
		daily.NewCommands(s.db),
		planning.NewCommands(s.db),
		recycle.NewCommands(s.db, s.recycleSecret),
		token.NewCommands(s.db),
	}

	for _, handler := range s.handlers {
		handler.Register(s)
	}
}

func (s *serveCmdOptions) onInteractionCreate(sess *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		// what we're doing here is allowing the app to set a custom ID after -- to pass along hidden values like an ID
		name := strings.Split(i.ApplicationCommandData().Name, "--")[0]
		for _, handler := range s.onInteractionCreateHandler[name] {
			handler(sess, i)
		}
	}

	if i.Type == discordgo.InteractionMessageComponent {
		// what we're doing here is allowing the app to set a custom ID after -- to pass along hidden values like an ID
		name := strings.Split(i.MessageComponentData().CustomID, "--")[0]
		for _, handler := range s.onInteractionCreateHandler[name] {
			handler(sess, i)
		}
	}

	if i.Type == discordgo.InteractionModalSubmit {
		data := i.ModalSubmitData()
		// what we're doing here is allowing the app to set a custom ID after -- to pass along hidden values like an ID
		name := strings.Split(data.CustomID, "--")[0]
		for _, handler := range s.onInteractionCreateHandler[name] {
			handler(sess, i)
		}
	}
}

func (s *serveCmdOptions) RegisterInteractionCreate(command string, fn func(*discordgo.Session, *discordgo.InteractionCreate)) {
	if s.onInteractionCreateHandler == nil {
		s.onInteractionCreateHandler = map[string][]func(*discordgo.Session, *discordgo.InteractionCreate){}
	}

	if _, exists := s.onInteractionCreateHandler[command]; !exists {
		s.onInteractionCreateHandler[command] = []func(*discordgo.Session, *discordgo.InteractionCreate){}
	}

	s.onInteractionCreateHandler[command] = append(s.onInteractionCreateHandler[command], fn)
}
