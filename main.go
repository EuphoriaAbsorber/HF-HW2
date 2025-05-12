package main

import (
	in "main/internal"
	h "main/internal/handler"
	logic "main/internal/logic"
	rep "main/internal/repository"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(h.NewService),
		fx.Provide(logic.NewService),
		fx.Provide(rep.NewService),
		fx.Provide(in.NewApp),

		fx.Invoke(func(*in.App) {}),
	).Run()
}
