package main

import (
	"context"
	"fmt"
	"io"
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
			queuedAt := time.Now()
			embed := &discordgo.MessageEmbed{
				Title:       ":clock2: Queued",
				Description: fmt.Sprintf("```\n%s\n```", cmd.Exec),
				Color:       0xd7d7d7,
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			if err != nil {
				logger.Error("failed to send a deferred response", zap.Error(err), zap.Any("interaction", i))
				return
			}
			queueTimeElapsed := time.Since(queuedAt)
			// todo: separate following command execution code as a dispatcher
			// todo: a handler for timeout
			// todo: support to make signal types configurable
			var ctx context.Context
			var cancel context.CancelFunc
			if cmd.Timeout.Seconds > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cmd.Timeout.Seconds)*time.Second)
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			args := strings.Split(cmd.Exec, " ")
			cmdExecutor := exec.CommandContext(ctx, args[0], args[1:]...)
			stdoutPipe, _ := cmdExecutor.StdoutPipe()
			stderrPipe, _ := cmdExecutor.StderrPipe()
			err = cmdExecutor.Start()
			if err != nil {
				embed = &discordgo.MessageEmbed{
					Title:       ":boom: Launch Error",
					Description: fmt.Sprintf("An error while launching the process.\n```\n%s\n```", cmd.Exec),
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("queue = %0.02fs", queueTimeElapsed.Seconds()),
					},
					Color: 0xc92323,
				}
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{embed},
				})
				if err != nil {
					logger.Error("failed to send an launch failed message", zap.Error(err), zap.Any("interaction", i))
					return
				}
				return
			}
			// launched a process
			executedAt := time.Now()
			embed = &discordgo.MessageEmbed{
				Title:       ":rocket: Launched",
				Description: fmt.Sprintf("Waiting for completion...\n```\n%s\n```", cmd.Exec),
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("queue = %0.02fs, pid = %d",
						queueTimeElapsed.Seconds(), cmdExecutor.Process.Pid),
				},
				Color: 0x3dc7df,
			}
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				logger.Error("failed to send an launch completion message", zap.Error(err), zap.Any("interaction", i))
				return
			}
			// wait for completion
			logger.Debug("waiting for a completion...", zap.Any("interaction", i), zap.Any("timeout", cmd.Timeout))
			stdout, _ := io.ReadAll(stdoutPipe)
			stderr, _ := io.ReadAll(stderrPipe)
			_ = cmdExecutor.Wait()
			executionTimeElapsed := time.Since(executedAt)
			footer := &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("queue = %0.02fs, pid = %d, execution = %0.02fs, exit code = %d",
					queueTimeElapsed.Seconds(), cmdExecutor.Process.Pid,
					executionTimeElapsed.Seconds(), cmdExecutor.ProcessState.ExitCode()),
			}
			if cmdExecutor.ProcessState.Success() {
				// no error
				embed = &discordgo.MessageEmbed{
					Title:       ":white_check_mark: Execution Completed",
					Description: fmt.Sprintf("```\n%s\n---\n%s%s\n```", cmd.Exec, stdout, stderr),
					Footer:      footer,
					Color:       0x12dd00,
				}
			} else {
				var msg string
				if ctx.Err() == nil {
					msg = fmt.Sprintf("Error: %s", err)
				} else {
					msg = fmt.Sprintf("Error: %s", ctx.Err())
				}
				embed = &discordgo.MessageEmbed{
					Title:       ":warning: Completed with an error",
					Description: fmt.Sprintf("%s\n```\n%s\n---\n%s%s\n```", msg, cmd.Exec, stdout, stderr),
					Footer:      footer,
					Color:       0xddd200,
				}
			}
			logger.Debug("completed", zap.Any("interaction", i), zap.Error(err))
			_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				logger.Error("failed to send an execution completion message", zap.Error(err), zap.Any("interaction", i))
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
