package app

import (
	"github.com/cng1985/erpgo/backend/internal/api"
	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/cng1985/erpgo/backend/internal/server"
	"github.com/cng1985/erpgo/backend/internal/store"
	"github.com/cng1985/erpgo/backend/internal/web"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"erp",
	config.Module,
	store.Module,
	api.Module,
	web.Module,
	server.Module,
)
