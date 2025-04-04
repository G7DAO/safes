package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/G7DAO/safes/bindings/Safe"
	"github.com/G7DAO/safes/bindings/SafeL2"
	"github.com/G7DAO/safes/bindings/SafeProxy"
	"github.com/G7DAO/safes/bindings/SafeProxyFactory"
)

var SAFES_VERSION = "0.0.1"

func main() {
	if err := CreateRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func CreateRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "game7",
		Short: "game7: CLI to the Game7 protocol",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	rootCmd.SetOut(os.Stdout)
	rootCmd.AddCommand(
		CreateCompletionCommand(rootCmd),
		CreateVersionCommand(),
		createSafeCommand("singleton", Safe.CreateSafeCommand),
		createSafeCommand("singleton-l2", SafeL2.CreateSafeL2Command),
		createSafeCommand("proxy", SafeProxy.CreateSafeProxyCommand),
		createSafeCommand("factory", SafeProxyFactory.CreateSafeProxyFactoryCommand),
		CreateDelegateCmd(),
	)

	return rootCmd
}

func CreateCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts for game7",
		Long: `Generate shell completion scripts for game7. 
Source the script to get completions in your shell session. For example, for bash:
  $ . <(game7 completion bash)`,
	}

	addSubCommands(completionCmd, rootCmd)
	return completionCmd
}

func addSubCommands(cmd *cobra.Command, rootCmd *cobra.Command) {
	subCommands := []struct {
		use, short string
		run        func(cmd *cobra.Command, args []string)
	}{
		{"bash", "bash completions for game7", func(cmd *cobra.Command, args []string) { rootCmd.GenBashCompletion(cmd.OutOrStdout()) }},
		{"zsh", "zsh completions for game7", func(cmd *cobra.Command, args []string) { rootCmd.GenZshCompletion(cmd.OutOrStdout()) }},
		{"fish", "fish completions for game7", func(cmd *cobra.Command, args []string) { rootCmd.GenFishCompletion(cmd.OutOrStdout(), true) }},
		{"powershell", "powershell completions for game7", func(cmd *cobra.Command, args []string) { rootCmd.GenPowerShellCompletion(cmd.OutOrStdout()) }},
	}

	for _, sub := range subCommands {
		cmd.AddCommand(&cobra.Command{Use: sub.use, Short: sub.short, Run: sub.run})
	}
}

func CreateVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of game7 that you are currently using",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Println(SAFES_VERSION) },
	}
}

func createSafeCommand(name string, factory func() *cobra.Command) *cobra.Command {
	cmd := factory()
	cmd.Use = name
	return cmd
}
