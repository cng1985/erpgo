package main

import (
	"github.com/cng1985/erpgo/backend/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
