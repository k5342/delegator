package main

import (
	"os"
)

type Config struct {
	SlashCommandPrefix string    `yaml:"slash_command_prefix"`
	DiscordBotToken    string    `yaml:"discord_bot_token"`
	Commands           []Command `yaml:"commands"`
}

type Command struct {
	BaseCommand `yaml:"base_command,inline"`
	Subcommands []BaseCommand `yaml:"subcommands,omitempty"`
}

type TimeoutConfig struct {
	Signal  string `yaml:"signal,omitempty"`
	Seconds int    `yaml:"seconds"`
}

type BaseCommand struct {
	Name        string        `yaml:"name"`
	Exec        string        `yaml:"exec"`
	Description string        `yaml:"description,omitempty"`
	Timeout     TimeoutConfig `yaml:"timeout,omitempty"`
}

func generateDefaultConfig() string {
	return `
# prefix is a root command name to call the bot
prefix: delegator

# you can list commands here
commands:
	- name: date
		exec: /usr/bin/date # a full path to the command
		description: Returns a result of date command
		subcommands:
			- name: unixtime
				description: Returns a result of date command in Unixtime
				exec: /usr/bin/date +%s # arguments are allowed and it is separated by the space.
	- name: timeout
		exec: /usr/bin/sleep 10
		description: an example for execution time timeout
		timeout:
		  Signal: SIGTERM
		  Seconds: 1 # default is infinity (= 0). this limits to 1s`
}
func getConfigPath() string {
	return "./config.yaml"
}

func createDefaultConfig() error {
	return os.WriteFile(getConfigPath(), []byte(generateDefaultConfig()), 0600)
}
