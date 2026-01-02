package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/cli"
)

var exitFunc = os.Exit

func main() {
	exitFunc(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, out io.Writer, errOut io.Writer) int {
	command := cli.New()
	exitCode := -1
	parser, err := kong.New(command,
		kong.Name("spogo"),
		kong.Description("Spotify power CLI using web cookies."),
		kong.UsageOnError(),
		kong.Writers(out, errOut),
		kong.Vars(cli.VersionVars()),
		kong.Exit(func(code int) {
			exitCode = code
		}),
	)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	kctx, err := parser.Parse(args)
	if exitCode >= 0 {
		return exitCode
	}
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	settings, err := command.Globals.Settings()
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	ctx, err := app.NewContext(settings)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	if err := ctx.ValidateProfile(); err != nil {
		fmt.Fprintln(errOut, err)
		return 2
	}
	if err := kctx.Run(ctx); err != nil {
		ctx.Output.Errorf("%v", err)
		return app.ExitCode(err)
	}
	return 0
}
