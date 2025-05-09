package main

import (
	in "main/internal"
	h "main/internal/handler"
	log "main/internal/logic"
	rep "main/internal/repository"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(h.NewService),
		fx.Provide(log.NewService),
		fx.Provide(rep.NewService),
		fx.Provide(in.NewApp),

		fx.Invoke(func(*in.App) {}),
	).Run()
}
