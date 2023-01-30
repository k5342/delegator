package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "delegator",
	}
	rootCmd.AddCommand(&cobra.Command{
		Use: "run",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
			os.Exit(1)
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
