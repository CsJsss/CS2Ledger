// Package factory resolves the import cycle between platform (interface+types)
// and platform subpackages (buff/youpin/c5/igxe). The subpackages import
// platform to use BaseClient, TradeRecord, etc.; platform cannot import the
// subpackages without creating a cycle. Factory sits between them — it
// imports both sides and provides a single NewPlatformClient entry point.
package factory

import (
	"fmt"
	"strings"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/buff"
	"github.com/CsJsss/CS2Ledger/pkg/platform/c5"
	"github.com/CsJsss/CS2Ledger/pkg/platform/igxe"
	"github.com/CsJsss/CS2Ledger/pkg/platform/youpin"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

type PlatformFactory struct {
}

func NewPlatformFactory() *PlatformFactory {
	return &PlatformFactory{}
}

// New creates a platform client for the given name and credential.
func (f *PlatformFactory) New(platformName, credential string, logger *logfx.Logger) (platform.Client, error) {
	switch strings.ToLower(platformName) {
	case platform.PlatformBuff:
		log := logger.WithComponent(platform.PlatformBuff)
		return buff.New(credential, log), nil
	case platform.PlatformYoupin:
		log := logger.WithComponent(platform.PlatformYoupin)
		return youpin.New(credential, log), nil
	case platform.PlatformC5:
		log := logger.WithComponent(platform.PlatformC5)
		return c5.New(credential, log), nil
	case platform.PlatformIGXE:
		log := logger.WithComponent(platform.PlatformIGXE)
		c, err := igxe.New(credential, log)
		if err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, fmt.Errorf("unknown platform: %s", platformName)
	}
}
