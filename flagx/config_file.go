package flagx

import (
	"log"

	"gopkg.in/ini.v1"
)

func parseFile(configFile string, flags []string) map[string]string {
	kvs := map[string]string{}
	// only support ini file for now
	if err := parseIni(configFile, flags, kvs); err != nil {
		// warn if parse error
		log.Printf("parse config file %s error: %v\n", configFile, err)
		return nil
	}
	return kvs
}

func parseIni(configFile string, flags []string, result map[string]string) error {
	cfg, err := ini.Load(configFile)
	if err != nil {
		return err
	}
	for _, flag := range flags {
		if v := cfg.Section("").Key(flag).String(); v != "" {
			result[flag] = v
		}
	}
	return nil
}
