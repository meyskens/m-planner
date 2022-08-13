package token

import (
	"fmt"
	"math/rand"

	"github.com/meyskens/m-planner/pkg/db"

	"github.com/bwmarrin/discordgo"
)

func (t *TokenCommands) registerToken(s *discordgo.Session, i *discordgo.InteractionCreate) {

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	token := make([]rune, 32)
	for i := range token {
		token[i] = letters[rand.Intn(len(letters))]
	}

	dbToken := db.ApiToken{
		User:  i.Member.User.ID,
		Token: string(token),
	}

	if tx := t.db.Save(&dbToken); tx.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Sorry friend, i got a database error %q :(", tx.Error),
			},
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,

		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Your token is %s", string(token)),
			Flags:   64, // ephemeral
		},
	})
}
