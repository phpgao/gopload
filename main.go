package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "gopload"
	app.Usage = "upload file via cli"
	app.Flags = appFlags()
	app.Action = appAction

	log.Fatal(app.Run(os.Args))
}

func appFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

func appAction(cCtx *cli.Context) error {
	conf := &Config{
		Bind:   cCtx.String("listen"),
		Debug:  cCtx.Bool("debug"),
		MaxInt: cCtx.Int("max"),
		Length: cCtx.Int("length"),
		Dir:    cCtx.String("dir"),
		Expire: cCtx.Int("expire"),
	}
	return NewServer(conf).Run()
}
