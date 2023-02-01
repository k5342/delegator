package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
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
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "please wait...",
					},
				})
				// todo: dispatch command and execute here
				// todo: wait for completion
				time.Sleep(time.Second * 10)
				// todo: update result placeholder
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "completed!",
					},
				})
			})()
		}
	})
	err = session.Open()
	if err != nil {
		return err
	}
	logger.Sugar().Info("bot launched")
	bot.session = session
	return nil
}

func (bot *DiscordBot) TerminateSession() error {
	err := bot.session.Close()
	if err != nil {
		return err
	}
	return nil
}
