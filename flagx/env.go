package flagx

import (
	"os"
	"strings"
)

func parseEnv(prefix string, flags []string) map[string]string {
	env := map[string]string{}
	p := strings.ToUpper(prefix)
	if !strings.HasSuffix(p, "_") {
		p += "_"
	}
	for _, f := range flags {
		if v, exist := os.LookupEnv(p + strings.ToUpper(f)); exist {
			env[f] = v
		}
	}
	return env
}
