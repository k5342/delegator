package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func main() {
	logger, _ = zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	go func() {
		logger.Sugar().Info(http.ListenAndServe("localhost:6060", nil))
	}()

	rootCmd := &cobra.Command{
		Use: "delegator",
	}
	rootCmd.AddCommand(&cobra.Command{
		Use: "run",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfigFromFile()
			if err != nil {
				logger.Fatal("failed to load the config file", zap.Error(err))
			}
			bot := NewDiscordBot(config.DiscordBotToken)
			err = bot.LaunchSession()
			if err != nil {
				logger.Fatal("failed to launch a bot session", zap.Error(err))
			}
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
			<-sc
			err = bot.TerminateSession()
			if err != nil {
				logger.Warn("failed to terminate the bot session", zap.Error(err))
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := os.Stat(getConfigPath())
			if err == nil {
				fmt.Println("config file is already exist")
				os.Exit(1)
			}
			err = createDefaultConfig()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("successfully generated!")
		},
	})
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
