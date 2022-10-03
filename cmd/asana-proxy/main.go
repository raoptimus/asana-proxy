package main

import (
	"fmt"
	"github.com/raoptimus/asana-proxy/asana"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var (
	Version   string
	GitCommit string
)

func main() {
	app := cli.NewApp()
	app.Name = "asana-proxy"
	app.Usage = "Asana API Proxy and changes response to other format"
	app.Version = fmt.Sprintf("v%s.rev[%s]", Version, GitCommit)
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			EnvVars: []string{"DEBUG"},
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "Enable debug mode. Default false",
		},
		&cli.StringFlag{
			Name:    "debug-addr",
			EnvVars: []string{"DEBUG_ADDR"},
			Aliases: []string{"da"},
			Value:   ":6060",
			Usage:   "HTTP addr for listening of pprof. Uses if debug mode is enabled",
		},
		&cli.StringFlag{
			Name:    "asana-url",
			EnvVars: []string{"ASANA_URL"},
			Aliases: []string{"url"},
			Value:   "https://app.asana.com/api/1.0",
			Usage:   "Asana base URL",
		},
		&cli.StringFlag{
			Name:    "server-addr",
			EnvVars: []string{"SERVER_ADDR"},
			Aliases: []string{"saddr"},
			Value:   ":8089",
			Usage:   "Server address listening",
		},
		&cli.StringFlag{
			Name:    "log-level",
			EnvVars: []string{"LOG_LEVEL"},
			Aliases: []string{"ll"},
			Value:   "info",
			Usage:   "Level of logging. One of value: panic, fatal, error, warning, info, debug, trace",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		level, err := log.ParseLevel(ctx.String("log-level"))
		if err != nil {
			return err
		}
		log.SetLevel(level)

		if ctx.Bool("debug") {
			log.SetLevel(log.DebugLevel)

			go func() {
				log.Infof("Debug listening. See http://%s/debug/pprof/", ctx.String("debug-addr"))
				fmt.Println(http.ListenAndServe(ctx.String("debug-addr"), nil))
			}()
		}

		log.Infof("Level of logging: %s", log.GetLevel().String())

		svc := asana.NewProxy(asana.Options{
			URL:        ctx.String("asana-url"),
			ServerAddr: ctx.String("server-addr"),
		})

		return svc.Run()
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
