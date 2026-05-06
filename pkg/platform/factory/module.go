package factory

import "go.uber.org/fx"

var Module = fx.Module("platform-factory", fx.Provide(NewFactory))
