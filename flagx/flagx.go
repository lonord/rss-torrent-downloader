package flagx

import (
	"flag"
	"regexp"
	"strings"
)

type option struct {
	fileFlag     string
	envPrefix    string
	excludeFlags []string
}

type OptionFn func(*option)

func EnableFile(fileFlag string) OptionFn {
	return func(o *option) {
		o.fileFlag = fileFlag
	}
}

func EnableEnv(envPrefix string) OptionFn {
	return func(o *option) {
		o.envPrefix = envPrefix
	}
}

func ExcludeFlag(flag string) OptionFn {
	return func(o *option) {
		o.excludeFlags = append(o.excludeFlags, flag)
	}
}

func Parse(ofs ...OptionFn) {
	opt := parseOptions(ofs...)
	if opt.fileFlag != "" && flag.Lookup(opt.fileFlag) == nil {
		// if fileFlag not exists, create a hidden flag
		flag.String(opt.fileFlag, "", "config `file_path`")
	}

	flag.Parse()

	allFlags := getAllFlags()
	parsedFlags := getParsedFlags()
	convertedFlags := []string{}
	convertedFlagMapping := map[string]string{}
	for _, f := range allFlags {
		skip := false
		for _, excludeFlag := range opt.excludeFlags {
			if f == excludeFlag {
				skip = true
				continue
			}
		}
		for _, parsedFlag := range parsedFlags {
			if f == parsedFlag {
				skip = true
				continue
			}
		}
		if skip {
			continue
		}
		cf := convertFlag(f)
		convertedFlags = append(convertedFlags, cf)
		convertedFlagMapping[cf] = f
	}

	if opt.envPrefix != "" {
		envFlags := parseEnv(opt.envPrefix, convertedFlags)
		for cf, value := range envFlags {
			flag.Set(convertedFlagMapping[cf], value)
		}
	}
	if opt.fileFlag != "" {
		if fileFlag := flag.Lookup(opt.fileFlag); fileFlag != nil && fileFlag.Value.String() != "" {
			fileFlags := parseFile(fileFlag.Value.String(), convertedFlags)
			for cf, value := range fileFlags {
				flag.Set(convertedFlagMapping[cf], value)
			}
		}
	}
}

func parseOptions(ofs ...OptionFn) *option {
	o := &option{}
	for _, of := range ofs {
		of(o)
	}
	return o
}

func getAllFlags() []string {
	flags := []string{}
	flag.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})
	return flags
}

func getParsedFlags() []string {
	flags := []string{}
	flag.Visit(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})
	return flags
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func convertFlag(f string) string {
	if strings.Contains(f, "-") {
		return strings.ReplaceAll(f, "-", "_")
	}
	// camelCase to snake_case
	snake := matchFirstCap.ReplaceAllString(f, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
