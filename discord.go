package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
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

func (bot *DiscordBot) lookupCommandByName(commandName string) (bool, *Command) {
	for _, cmd := range bot.config.Commands {
		if commandName == cmd.Name {
			return true, &cmd
		}
	}
	return false, nil
}

func (bot *DiscordBot) LaunchSession() error {
	session, err := discordgo.New("Bot " + bot.config.DiscordBotToken)
	if err != nil {
		return err
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// check whether the bot has a corresponding command handler of a requested command
		commandName := i.ApplicationCommandData().Name
		logger.Debug("interaction created", zap.Any("interaction", i))
		found, cmd := bot.lookupCommandByName(commandName)
		logger.Debug("getCommandExec", zap.Bool("found", found), zap.String("cmdExec", cmd.Exec))
		if !found {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("hmm, command `/%s` is not found or outdated. please reload the bot or cleanup unnecessary slash commands.", cmd.Exec),
				},
			})
			if err != nil {
				logger.Error("failed to send an error response", zap.Error(err), zap.Any("interaction", i))
				return
			}
			return
		}
		// todo: limit maximum executions in parallel
		go (func() {
			// todo: enqueue request here
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Run: `%s`", cmd.Exec),
				},
			})
			if err != nil {
				logger.Error("failed to send a deferred response", zap.Error(err), zap.Any("interaction", i))
				return
			}
			// todo: separate following command execution code as a dispatcher
			// todo: a handler for timeout
			// todo: support to make signal types configurable
			logger.Debug("waiting for a completion...", zap.Any("interaction", i), zap.Any("timeout", cmd.Timeout))
			var ctx context.Context
			var cancel context.CancelFunc
			if cmd.Timeout.Seconds > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cmd.Timeout.Seconds)*time.Second)
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			args := strings.Split(cmd.Exec, " ")
			output, err := exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
			var msg string
			if err == nil {
				msg = fmt.Sprintf("```\n%s\n```", string(output))
			} else {
				msg = fmt.Sprintf("Error: %s", err)
			}
			logger.Debug("completed", zap.Any("interaction", i), zap.Error(err))
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			if err != nil {
				logger.Error("failed to send a result response", zap.Error(err), zap.Any("interaction", i))
				return
			}
		})()
	})
	err = session.Open()
	if err != nil {
		return err
	}
	logger.Sugar().Info("bot launched")
	bot.session = session

	// todo: warn duplicate commands
	// prepare commands as root commands
	var commands []*discordgo.ApplicationCommand
	for _, command := range bot.config.Commands {
		cmd := discordgo.ApplicationCommand{
			Name:        command.Name,
			Description: command.Description,
		}
		commands = append(commands, &cmd)
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
