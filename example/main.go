package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/daolis/toolsconfig"
	"github.com/daolis/toolsconfig/commands"
)

var rootCmd = &cobra.Command{
	Use:   "example",
	Short: "My example tool",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "my first command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Execute test command")

		configuration, err := toolsconfig.NewToolConfiguration(
			toolsconfig.WithConfigDir("."),
			toolsconfig.WithConfigFile("unittestconfig.yaml"),
			toolsconfig.UpdateConfig(false),
			toolsconfig.RequiredServer("testserver.io"),
			toolsconfig.RequiredSubscription("testSubscription01"))
		if err != nil {
			// required values missing
			var configErr *toolsconfig.ConfigError
			if errors.As(err, &configErr) {
				missingCredentials := configErr.Missing
				fmt.Println("Values missing:", missingCredentials)
			}
			return
		}
		repoCredentials, err := configuration.GetServerCredentials("testserver.io")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Credentials for testserver.io")
		fmt.Println("Username:", repoCredentials.Username)
		fmt.Println("Password:", repoCredentials.Password)
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(testCmd)

	commands.AddToRootCommand(rootCmd, commands.WithRunFunctions(
		func(cmd *cobra.Command, args []string) {
			fmt.Println("Execute root command")
			fmt.Println("To see the credentials run with parameter 'test'")
		},
	))
}
