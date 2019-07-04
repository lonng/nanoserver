package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/lonng/nanoserver/internal/game"
	"github.com/lonng/nanoserver/internal/web"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	// base application info
	app.Name = "mahjong server"
	app.Author = "MaJong"
	app.Version = "0.0.1"
	app.Copyright = "majong team reserved"
	app.Usage = "majiang server"

	// flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "./configs/config.toml",
			Usage: "load configuration from `FILE`",
		},
		cli.BoolFlag{
			Name:  "cpuprofile",
			Usage: "enable cpu profile",
		},
	}

	app.Action = serve
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	viper.SetConfigType("toml")
	viper.SetConfigFile(c.String("config"))
	viper.ReadInConfig()

	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	if viper.GetBool("core.debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.Bool("cpuprofile") {
		filename := fmt.Sprintf("cpuprofile-%d.pprof", time.Now().Unix())
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() { defer wg.Done(); game.Startup() }() // 开启游戏服
	go func() { defer wg.Done(); web.Startup() }()  // 开启web服务器

	wg.Wait()
	return nil
}
