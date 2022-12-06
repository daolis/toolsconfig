package commands

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daolis/toolsconfig"
)

var rootArgs struct {
	saveName         string
	runFavouriteName string
}

var favCmd = &cobra.Command{
	Use:     "fav",
	Aliases: []string{"favourite"},
	Short:   "Favourites",
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(cmd.Help())
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// do not call PersistentPreRun from parent
	},
}

var favListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List favourites",
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(listFavourites(cmd.Root().Name()))
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// do not call PersistentPreRun from parent
	},
}

func newToolsConfig(options ...toolsconfig.ConfigOption) (toolsconfig.Configuration, error) {
	return toolsconfig.NewToolConfiguration(options...)
}

// AddToRootCommand adds all commands and flags to the given root command.
// If you want to use the Run and PersistentPostRun functions, you need to add them using the
//WithRunFunctions and WithPersistentPostRunFunctions options.
//These functions are called after the internal toolconfig functions.
// Commands:
// * fav (Favourites)
// * fav list (List favourites)
// Flags:
// * --save <name> (Save favourite, persistent flag to use it on every sub command)
// * --run <name> (Run favourite)
func AddToRootCommand(command *cobra.Command, opts ...commandOption) {
	if command.HasParent() {
		panic("AddToRootCommand can only be called with the root command!")
	}
	if command.Run != nil {
		panic("Run function already set on root command !")
	}
	options := &commandOptions{}
	for _, opt := range opts {
		opt(options)
	}

	persistentPostRun := func(cmd *cobra.Command, args []string) {
		if len(rootArgs.saveName) != 0 {
			var saveArgs []string
			var removeNextToo bool
			for _, arg := range os.Args[1:] {
				if arg == "--save" {
					removeNextToo = true
					continue
				}
				if removeNextToo {
					removeNextToo = false
					continue
				}
				saveArgs = append(saveArgs, arg)
			}
			cobra.CheckErr(saveFavourite(cmd.Root().Name(), rootArgs.saveName, saveArgs))
			log.WithField("name", rootArgs.saveName).Info("Saved command as favourite")
			return
		}
		for _, runFn := range options.persistentPostRunFunctions {
			runFn(cmd, args)
		}
	}

	rootRun := func(cmd *cobra.Command, args []string) {
		if len(rootArgs.runFavouriteName) != 0 {
			cfg, err := toolsconfig.NewToolConfiguration()
			cobra.CheckErr(err)
			favourite, err := cfg.GetFavourite(cmd.Root().Name(), rootArgs.runFavouriteName)
			cobra.CheckErr(err)
			cmd.SetArgs(favourite.Args)
			log.WithFields(log.Fields{"name": rootArgs.runFavouriteName, "args": strings.Join(favourite.Args, " ")}).Info("Running saved favourite")
			cobra.CheckErr(cmd.Execute())
			return
		}
		for _, runFn := range options.runFunctions {
			runFn(cmd, args)
		}
	}

	command.PersistentFlags().StringVar(&rootArgs.saveName, "save", "", "Save the command with the given name!")
	command.Flags().StringVar(&rootArgs.runFavouriteName, "run", "", "Run the saved favourite with the given name")
	_ = command.RegisterFlagCompletionFunc("run", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := newToolsConfig()
		if err != nil {
			panic(err)
		}
		favourites := cfg.GetFavourites(cmd.Root().Name())
		favNames := make([]string, len(favourites))
		for idx, favourite := range favourites {
			favNames[idx] = favourite.Name
		}
		return favNames, cobra.ShellCompDirectiveDefault
	})

	command.AddCommand(favCmd)
	command.PersistentPostRun = persistentPostRun
	command.Run = rootRun
}

func InitialRootCommand(use, short, long string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}
	AddToRootCommand(rootCmd)
	return rootCmd
}

type commandOption func(options *commandOptions)

type commandOptions struct {
	runFunctions               []func(cmd *cobra.Command, args []string)
	persistentPostRunFunctions []func(cmd *cobra.Command, args []string)
}

func WithRunFunctions(functions ...func(cmd *cobra.Command, args []string)) commandOption {
	return func(options *commandOptions) {
		options.runFunctions = functions
	}
}

func WithPersistentPostRunFunctions(functions ...func(cmd *cobra.Command, args []string)) commandOption {
	return func(options *commandOptions) {
		options.persistentPostRunFunctions = functions
	}
}

func init() {
	favCmd.AddCommand(favListCmd)
}
