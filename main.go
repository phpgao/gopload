package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/phpgao/gopload/internal/config"
	"github.com/phpgao/gopload/internal/service"
)

func main() {
	app := &cli.App{
		Name:  "gopload",
		Usage: "upload file via cli",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen",
				Usage:   "0.0.0.0:8088 or :8088",
				Value:   ":8088",
				EnvVars: []string{"GOPLOAD_SERVER_ADDR"},
				Aliases: []string{"b"},
			},
			&cli.StringFlag{
				Name:    "dir",
				Usage:   "dir for storing file",
				Value:   "",
				EnvVars: []string{"GOPLOAD_SERVER_DIR"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "debug mode",
				EnvVars: []string{"GOPLOAD_DEBUG"},
			},
			&cli.IntFlag{
				Name:    "length",
				Usage:   "dir-length",
				Value:   7,
				EnvVars: []string{"GOPLOAD_DRI_LENGTH"},
				Aliases: []string{"l"},
			},
			&cli.IntFlag{
				Name:    "max",
				Usage:   "max file size in MB",
				Value:   100,
				EnvVars: []string{"GOPLOAD_MAX_SIZE"},
				Aliases: []string{"m"},
			},
			&cli.IntFlag{
				Name:    "expire",
				Usage:   "delete file after expire in days",
				Value:   3,
				EnvVars: []string{"GOPLOAD_EXPIRE"},
				Aliases: []string{"e"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			conf := &config.Config{
				Bind:   cCtx.String("listen"),
				Debug:  cCtx.Bool("debug"),
				MaxInt: cCtx.Int("max"),
				Length: cCtx.Int("length"),
				Dir:    cCtx.String("dir"),
				Expire: cCtx.Int("expire"),
			}
			return service.NewServer(conf).Run()
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
