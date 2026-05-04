// Command server is the application entrypoint. It dispatches to the
// real work in internal/app, which knows how to serve, migrate, or emit
// the OpenAPI spec for the typed-client generator.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yackey-labs/yauth-go-vue-template/server/internal/app"
)

const usage = `usage: server [command] [flags]

Commands:
  serve       Run the HTTP server (default).
  migrate     Run schema migrations and exit. Idempotent.
  gen-spec    Emit the application OpenAPI 3.1 document and exit. Use
              this to (re)generate the typed TS client (` + "`task gen`" + `).
  help        Show this message.

Common flags:
  -c <path>   yauth.yaml path (default "yauth.yaml" relative to cwd).

gen-spec flags:
  -o <path>   Output file. Defaults to stdout when omitted or "-".
`

func main() {
	if err := dispatch(); err != nil {
		log.Fatal(err)
	}
}

func dispatch() error {
	cmd := "serve"
	args := os.Args[1:]
	// Allow flags before any subcommand: `server -c yauth.yaml` is the
	// same as `server serve -c yauth.yaml`. We only consume os.Args[1]
	// as a subcommand when it doesn't look like a flag.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}

	ctx := context.Background()
	switch cmd {
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		cfg := fs.String("c", app.DefaultConfigPath, "path to yauth.yaml")
		_ = fs.Parse(args)
		return app.Serve(ctx, *cfg)
	case "migrate":
		fs := flag.NewFlagSet("migrate", flag.ExitOnError)
		cfg := fs.String("c", app.DefaultConfigPath, "path to yauth.yaml")
		_ = fs.Parse(args)
		return app.Migrate(ctx, *cfg)
	case "gen-spec":
		fs := flag.NewFlagSet("gen-spec", flag.ExitOnError)
		out := fs.String("o", "", "output path (default: stdout)")
		_ = fs.Parse(args)
		return app.GenSpec(*out)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(2)
		return nil
	}
}
