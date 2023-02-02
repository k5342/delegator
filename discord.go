package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordBot struct {
	config       *Config
	session      *discordgo.Session
	sessionStore *SessionStore
}

func NewDiscordBot(config *Config) *DiscordBot {
	bot := DiscordBot{
		config:       config,
		sessionStore: NewSessionStore(),
	}
	return &bot
}

func (bot *DiscordBot) LaunchSession() error {
	session, err := discordgo.New("Bot " + bot.config.DiscordBotToken)
	if err != nil {
		return err
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commandName := i.ApplicationCommandData().Name
		// todo: limit maximum executions in parallel
		if commandName == bot.config.SlashCommandPrefix {
			go (func() {
				// todo: enqueue request here
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "please wait...",
					},
				})
				if err != nil {
					logger.Error("failed to send a deferred response", zap.Error(err), zap.Any("interaction", i))
					return
				}
				// todo: dispatch command and execute here
				// todo: wait for completion
				logger.Debug("waiting for a completion...", zap.Any("interaction", i))
				time.Sleep(time.Second * 10)
				logger.Debug("completed", zap.Any("interaction", i))
				// todo: update result placeholder
				msg := "completed!"
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				if err != nil {
					logger.Error("failed to send a result response", zap.Error(err), zap.Any("interaction", i))
					return
				}
			})()
		}
	})
	err = session.Open()
	if err != nil {
		return err
	}
	logger.Sugar().Info("bot launched")
	bot.session = session
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        bot.config.SlashCommandPrefix,
			Description: "run some specific commands",
		},
	}
	for _, val := range commands {
		logger.Sugar().Debug("creating a command", zap.String("command_name", val.Name))
		_, err := bot.session.ApplicationCommandCreate(bot.session.State.User.ID, "", val)
		if err == nil {
			logger.Sugar().Info("created a command", zap.String("command_name", val.Name))
		} else {
			logger.Sugar().Error("cannot create command", zap.String("command_name", val.Name), zap.Error(err))
		}
	}
	return nil
}

func (bot *DiscordBot) TerminateSession() error {
	err := bot.session.Close()
	if err != nil {
		return err
	}
	return nil
}
