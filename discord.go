package main

import "github.com/bwmarrin/discordgo"

type DiscordBot struct {
	token   string
	session *discordgo.Session
}

func NewDiscordBot(botToken string) *DiscordBot {
	bot := DiscordBot{
		token: botToken,
	}
	return &bot
}

func (bot *DiscordBot) LaunchSession() error {
	session, err := discordgo.New("Bot " + bot.token)
	if err != nil {
		return err
	}
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
