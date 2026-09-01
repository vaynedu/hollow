package main

import (
	"log"

	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/example/config"
	"github.com/vaynedu/hollow/example/router"
)

func run() error {
	app, err := hollow.NewApp(hollow.AppOption{
		ConfigPath: ".",
		ConfigName: "conf",
	})
	if err != nil {
		return err
	}

	app.Startup(config.Startup)
	app.Shutdown(config.Shutdown)
	router.Register(app)
	return app.Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
