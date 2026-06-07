package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/CsJsss/CS2Ledger/pkg/platform"
	"github.com/CsJsss/CS2Ledger/pkg/platform/factory"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "platform-cli",
	Short: "CS2Ledger platform API debug tool",
}

var validPlatforms = []string{"buff", "youpin", "c5", "igxe", "eco"}

var tokenEnv = map[string]string{
	"buff":   "BUFF_TOKEN",
	"youpin": "YOUPIN_TOKEN",
	"c5":     "C5_TOKEN",
	"igxe":   "IGXE_TOKEN",
	"eco":    "ECO_TOKEN",
}

var debugMode bool

func init() {
	rootCmd.PersistentFlags().StringP("token", "t", "", "Platform credential")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging")

	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(buyhistoryCmd)
	rootCmd.AddCommand(sellhistoryCmd)
	rootCmd.AddCommand(billhistoryCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func resolveToken(cmd *cobra.Command, platform string) (string, error) {
	t, _ := cmd.Flags().GetString("token")
	if t != "" {
		return t, nil
	}
	envKey := tokenEnv[strings.ToLower(platform)]
	if envKey == "" {
		return "", fmt.Errorf("unknown platform: %s", platform)
	}
	t = os.Getenv(envKey)
	if t == "" {
		return "", fmt.Errorf("token required: set --token or %s env var", envKey)
	}
	return t, nil
}

func parseExtraParams(pairs []string) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			continue
		}
		m[k] = v
	}
	return m
}

func createClient(platformName, token string) (platform.Client, error) {
	level := slog.LevelWarn
	if debugMode {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	f := factory.NewPlatformFactory()
	return f.New(platformName, token, &logfx.Logger{Logger: log})
}
