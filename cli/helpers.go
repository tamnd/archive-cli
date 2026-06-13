package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func itoa(n int) string     { return strconv.Itoa(n) }
func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

// humanBytes renders a byte count in a compact human form.
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func boolWord(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

// parseKV splits "key:value" or "key=value" pairs into a map (for --metadata).
func parseKV(pairs []string) (map[string]string, error) {
	out := map[string]string{}
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, ":")
		if !ok {
			k, v, ok = strings.Cut(p, "=")
		}
		if !ok {
			return nil, fmt.Errorf("expected key:value, got %q", p)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}
