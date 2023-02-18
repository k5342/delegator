package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DiscordBotToken string    `yaml:"discord_bot_token"`
	Commands        []Command `yaml:"commands"`
}

type Command struct {
	BaseCommand `yaml:"base_command,inline"`
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
# issue your discord bot token from Discord Developer Portal
discord_bot_token: FILL_IT_HERE

# you can list commands here
commands:
- name: date
  exec: /usr/bin/date # a full path to the command
  description: Returns a result of date command
- name: timeout
  exec: /usr/bin/sleep 10
  description: an example for execution time timeout
  timeout:
    signal: SIGTERM
    seconds: 1 # default is infinity (= 0). this limits to 1s`
}
func getConfigPath() string {
	return "./config.yaml"
}

func createDefaultConfig() error {
	return os.WriteFile(getConfigPath(), []byte(generateDefaultConfig()), 0600)
}

func loadConfigFromFile() (*Config, error) {
	bytes, err := os.ReadFile(getConfigPath())
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
