package mdgo

import "strings"

var (
	replacer = strings.NewReplacer(
		"_", "\\_",
		"~", "\\~",
		"*", "\\*",
		"`", "\\`",
		"_", "\\_",
		">", "\\>",
		"<", "\\<",
		"|", "\\|",
	)
)
