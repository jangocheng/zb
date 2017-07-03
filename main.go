package main // import "jrubin.io/zb"

import (
	"os"
	"strings"

	"github.com/urfave/cli"

	"jrubin.io/slog"
	"jrubin.io/slog/handlers/text"
	"jrubin.io/zb/cmd"
	"jrubin.io/zb/cmd/build"
	"jrubin.io/zb/cmd/clean"
	"jrubin.io/zb/cmd/commands"
	"jrubin.io/zb/cmd/complete"
	"jrubin.io/zb/cmd/install"
	"jrubin.io/zb/cmd/lint"
	"jrubin.io/zb/cmd/list"
	"jrubin.io/zb/cmd/test"
	"jrubin.io/zb/cmd/version"
	"jrubin.io/zb/lib/zbcontext"
)

// TODO(jrubin) logo
// TODO(jrubin) fix all lint issues
// TODO(jrubin) test all the things
// TODO(jrubin) detect import cycles
// TODO(jrubin) godoc documentation
// TODO(jrubin) vendor? (wrap goimports)

var (
	// populated by zb build ldflags
	gitCommit, buildDate string

	level  = slog.InfoLevel
	app    = cli.NewApp()
	logger = slog.Logger{}

	ctx = zbcontext.Context{
		GitCommit: &gitCommit,
		BuildDate: &buildDate,
		Logger:    &logger,
	}
)

var subcommands = []cmd.Constructor{
	build.Cmd,
	clean.Cmd,
	commands.Cmd,
	complete.Cmd,
	install.Cmd,
	lint.Cmd,
	list.Cmd,
	test.Cmd,
	version.Cmd,
}

func init() {
	app.Name = "zb"
	app.HideVersion = true
	app.Version = "0.2.1"
	app.Usage = "an opinionated repo based tool for working with go"
	app.EnableBashCompletion = true
	app.BashComplete = cmd.BashComplete
	app.Before = func(*cli.Context) error {
		setup()
		return nil
	}
	app.Authors = []cli.Author{
		{Name: "Joshua Rubin", Email: "joshua@rubixconsulting.com"},
	}
	app.Flags = []cli.Flag{
		cli.GenericFlag{
			Name:   "log-level, l",
			EnvVar: "LOG_LEVEL",
			Usage:  "set log level (DEBUG, INFO, WARN, ERROR)",
			Value:  &level,
		},
		cli.BoolFlag{
			Name:        "no-warn-todo-fixme, n",
			EnvVar:      strings.ToUpper("no_warn_todo_fixme"),
			Usage:       "do not warn when finding " + strings.ToUpper("warn") + " or " + strings.ToUpper("fixme") + " in .go files",
			Destination: &ctx.NoWarnTodoFixme,
		},
		cli.StringFlag{
			Name:        "cache",
			Destination: &ctx.CacheDir,
			EnvVar:      "CACHE",
			Value:       cmd.DefaultCacheDir(app.Name),
			Usage:       "commands that cache results use this as their base directory",
		},
		cli.BoolFlag{
			Name:        "package, p",
			Destination: &ctx.Package,
			Usage:       "run tests only for the listed packages, not all packages in the projects",
		},
		cli.BoolFlag{
			Name:        "no-generate, g",
			EnvVar:      "NO_GENERATE",
			Destination: &ctx.NoGenerate,
			Usage:       "calculating dependencies for go generate can sometimes be slow, enable this to speed things up",
		},
	}

	app.Metadata = map[string]interface{}{}

	for _, sc := range subcommands {
		c := sc.New(app)
		c.Before = wrapFn(c.Before)
		c.Action = wrapFn(c.Action)
		c.After = wrapFn(c.After)
		if c.BashComplete == nil {
			c.BashComplete = cmd.BashCommandComplete(c)
		}
		app.Commands = append(app.Commands, c)
	}
}

func main() {
	_ = app.Run(os.Args) // nosec
}

func setup() {
	app.Metadata["Context"] = ctx
	logger.RegisterHandler(level, &text.Handler{
		Writer:           os.Stderr,
		DisableTimestamp: true,
	})
}
