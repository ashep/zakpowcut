package main

import (
	"github.com/ashep/go-apprun"

	"zakpowcut/internal/app"
)

func main() {
	apprun.Run(app.New, app.Config{})
}
