package main

import (
	"embed"
	"fmt"
	"github.com/XANi/toolbox/project-templates/go-gin-embedded/store/file"
	"github.com/XANi/toolbox/project-templates/go-gin-embedded/web"
	"github.com/efigence/go-mon"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
)

var version string
var log *zap.SugaredLogger
var debug = true

// /* embeds with all files, just dir/ ignores files starting with _ or .
//go:embed static templates
var webContent embed.FS
var stopper = make(map[string]func(), 0)

func init() {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	// naive systemd detection. Drop timestamp if running under it
	if os.Getenv("INVOCATION_ID") != "" || os.Getenv("JOURNAL_STREAM") != "" {
		consoleEncoderConfig.TimeKey = ""
	}
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleStderr := zapcore.Lock(os.Stderr)
	_ = consoleStderr
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return (lvl < zapcore.ErrorLevel) != (lvl == zapcore.DebugLevel && !debug)
	})
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, os.Stderr, lowPriority),
		zapcore.NewCore(consoleEncoder, os.Stderr, highPriority),
	)
	logger := zap.New(core)
	if debug {
		logger = logger.WithOptions(
			zap.Development(),
			zap.AddCaller(),
			zap.AddStacktrace(highPriority),
		)
	} else {
		logger = logger.WithOptions(
			zap.AddCaller(),
		)
	}
	log = logger.Sugar()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	go func() {
		for sig := range s {
			log.Warnf("got signal %s, exiting", sig)
			for k, f := range stopper {
				log.Infof("stopping %s", k)
				f()
			}
			os.Exit(0)
		}
	}()
}

func main() {
	defer log.Sync()
	// register internal stats
	mon.RegisterGcStats()
	app := cli.NewApp()
	app.Name = "ush"
	app.Description = "small http file sharing app"
	app.Version = version
	app.HideHelp = true
	log.Errorf("Starting %s version: %s", app.Name, version)
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "help, h", Usage: "show help"},
		cli.BoolFlag{Name: "debug, d", Usage: "enable debug logs"},
		cli.StringFlag{
			Name:   "listen-addr",
			Value:  "127.0.0.1:3001",
			Usage:  "Listen addr",
			EnvVar: "LISTEN_ADDR",
		},
		cli.StringFlag{
			Name:  "data-dir",
			Value: "!tmp",
			Usage: "data dir. specify !tmp to create a temporary one in system tmp location. It will be yeeted on app close",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		debug = c.Bool("debug")
		dir := c.String("data-dir")
		if dir == "!tmp" {
			var err error
			dir, err = os.MkdirTemp(os.TempDir(), "ush"+fmt.Sprintf("-%d-", os.Getpid()))
			if err != nil {
				log.Panicf("cannot create tmpdir [%s]: %s and different data dir was not specified in cmdline", dir, err)
			}
			log.Infof("temporary data dir in [%s]. It will be removed on app stop", dir)
			stopper["tmpdir"] = func() {
				log.Infof("removing tmp dir %s", dir)
				err := os.RemoveAll(dir)
				if err != nil {
					log.Panicf("cannot cleanup tmpdir [%s]: %s", dir, err)
				}
			}
		} else {
			log.Infof("data dir: %s", dir)
		}
		storage, err := file.New(file.Config{
			RootDir: dir,
			Logger:  log,
		})
		if err != nil {
			log.Panicf("error opening storage: %s", err)
		}
		accessLogCfg := zap.NewDevelopmentConfig()
		accessLogCfg.EncoderConfig.TimeKey = "T"
		accessLogCfg.EncoderConfig.LevelKey = ""
		accessLogCfg.EncoderConfig.NameKey = ""
		accessLogCfg.EncoderConfig.CallerKey = ""
		accessLog, _ := accessLogCfg.Build()
		w, err := web.New(web.Config{
			Logger:       log,
			AccessLogger: accessLog.Sugar(),
			ListenAddr:   c.String("listen-addr"),
			Storage:      storage,
		}, webContent)
		if err != nil {
			log.Panicf("error starting web listener: %s", err)
		}
		return w.Run()
	}
	// optional commands
	//app.Commands = []cli.Command{
	//	{
	//		Name:    "rem",
	//		Aliases: []string{"a"},
	//		Usage:   "example cmd",
	//		Action: func(c *cli.Context) error {
	//			log.Warnf("running example cmd")
	//			return nil
	//		},
	//	},
	//	{
	//		Name:    "add",
	//		Aliases: []string{"a"},
	//		Usage:   "example cmd",
	//		Action: func(c *cli.Context) error {
	//			log.Warnf("running example cmd")
	//			return nil
	//		},
	//	},
	//}
	// to sort do that
	// sort.Sort(cli.FlagsByName(app.Flags))
	// sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("err: %s", err)
	}
}
