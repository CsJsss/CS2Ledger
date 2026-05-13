package platform

import (
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"platform",
	logfx.WithComponent("platform"),
)
