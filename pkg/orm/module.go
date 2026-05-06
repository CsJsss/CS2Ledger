package orm

import "go.uber.org/fx"

var Module = fx.Module("orm",
	fx.Provide(NewORM),
)
