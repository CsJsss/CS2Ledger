// Package factory breaks the import cycle between the platform interface
// and its concrete implementations.
package factory

import (
	"fmt"
	"strings"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/buff"
	"github.com/CsJsss/CS2Ledger/pkg/platform/c5"
	"github.com/CsJsss/CS2Ledger/pkg/platform/igxe"
	"github.com/CsJsss/CS2Ledger/pkg/platform/youpin"
	"github.com/CsJsss/CS2Ledger/pkg/utils"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type Factory struct {
	log *logfx.Logger
}

func NewFactory(log *logfx.Logger) *Factory {
	return &Factory{log: log}
}

// New creates a platform client for the given name and credential.
func (f *Factory) New(platformName, credential string) (platform.Client, error) {
	switch strings.ToLower(platformName) {
	case utils.PlatformBuff:
		c := buff.New(credential)
		c.SetLogger(f.log.Logger)
		return c, nil
	case utils.PlatformYoupin, "悠悠有品":
		c := youpin.New(credential)
		c.SetLogger(f.log.Logger)
		return c, nil
	case utils.PlatformC5:
		c := c5.New(credential)
		c.SetLogger(f.log.Logger)
		return c, nil
	case utils.PlatformIGXE:
		c, err := igxe.New(credential)
		if err != nil {
			return nil, err
		}
		c.SetLogger(f.log.Logger)
		return c, nil
	default:
		return nil, fmt.Errorf("unknown platform: %s", platformName)
	}
}
