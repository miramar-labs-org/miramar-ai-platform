package main

import (
	"github.com/fatih/color"

	"github.com/miramar-labs-org/miramar-ai-platform/internal/doctor"
)

// color.NoColor is set once at init from NO_COLOR and whether stdout is a
// real terminal, so every helper below is a plain no-op (falls back to the
// original ASCII text) when output is piped, redirected, or in CI logs.
var (
	styleGoodColor = color.New(color.FgGreen)
	styleWarnColor = color.New(color.FgYellow)
	styleBadColor  = color.New(color.FgRed)
)

// styleOK prefixes msg with a green checkmark when color is enabled.
func styleOK(msg string) string {
	if color.NoColor {
		return msg
	}
	return styleGoodColor.Sprint("✓ ") + msg
}

// styleFail prefixes msg with a red cross when color is enabled.
func styleFail(msg string) string {
	if color.NoColor {
		return msg
	}
	return styleBadColor.Sprint("✗ ") + msg
}

// styleDoctorStatus renders a doctor.Status as a colored unicode glyph, or
// the original plain "[PASS]"/"[WARN]"/"[FAIL]" form when color is disabled.
func styleDoctorStatus(s doctor.Status) string {
	switch s {
	case doctor.Pass:
		if color.NoColor {
			return "[PASS]"
		}
		return styleGoodColor.Sprint("✓ PASS")
	case doctor.Warn:
		if color.NoColor {
			return "[WARN]"
		}
		return styleWarnColor.Sprint("⚠ WARN")
	case doctor.Fail:
		if color.NoColor {
			return "[FAIL]"
		}
		return styleBadColor.Sprint("✗ FAIL")
	default:
		return "[" + s.String() + "]"
	}
}
