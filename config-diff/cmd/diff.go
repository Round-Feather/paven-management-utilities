package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var diffCmd = &cobra.Command{
	Use:           "config-diff [config-directory] [source-env] [target-env]",
	Args:          cobra.ExactArgs(3),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          diffRun,
}

func init() {
}

func diffRun(cmd *cobra.Command, args []string) error {
	files, err := filepath.Glob(fmt.Sprintf("%s/*-service.yml", args[0]))
	if err != nil {
		fmt.Println(err)
	}

	report := []string{}

	for _, f := range files {
		f = strings.TrimSuffix(f, ".yml")
		fmt.Println()
		fmt.Printf("*** %s ***\n", f)
		sourceYamlName := fmt.Sprintf("%s-%s.yml", f, args[1])
		targetYamlName := fmt.Sprintf("%s-%s.yml", f, args[2])
		sourceYamlBytes, _ := os.ReadFile(sourceYamlName)
		targetYamlBytes, _ := os.ReadFile(targetYamlName)

		source := make(map[string]interface{})
		target := make(map[string]interface{})

		_ = yaml.Unmarshal(sourceYamlBytes, source)
		_ = yaml.Unmarshal(targetYamlBytes, target)

		_, missingKeys := checkKeys(source, target, "", "")
		if len(missingKeys) > 0 {
			report = append(report, fmt.Sprintf("*** %s ***", f))
			report = append(report, missingKeys...)
			report = append(report, "")
		}

	}

	fmt.Println()

	if len(report) > 0 {
		f, err := os.Create("missing_properties.txt")
		if err != nil {
			return err
		}

		for _, s := range report {
			fmt.Fprintln(f, s)
		}

		f.Close()

		return errors.New("missing some properties")
	}
	return nil
}

func checkKeys(source map[string]interface{}, target map[string]interface{}, indent string, root string) (error, []string) {
	missingKeys := []string{}
	missing := false
	for key, val := range source {
		_, subMap := val.(map[string]interface{})

		// When the the target yaml file is missing the entire object
		if target == nil {
			color.Red("%s%s: ✘", indent, key)
			missing = true

			// if the source yaml section is an object and not just a value
			if subMap {
				_, m := checkKeys(val.(map[string]interface{}), nil, indent+"  ", root+"."+key)
				missingKeys = append(missingKeys, m...)
			} else {
				_k := strings.TrimPrefix(root+"."+key, ".")
				missingKeys = append(missingKeys, _k)
			}
		} else {
			_, targetExists := target[key]
			if targetExists {
				color.Green("%s%s: ✔", indent, key)
				if subMap {
					_, m := checkKeys(val.(map[string]interface{}), target[key].(map[string]interface{}), indent+"  ", root+"."+key)
					missingKeys = append(missingKeys, m...)
				}
			} else {
				color.Red("%s%s: ✘", indent, key)
				missing = true
				if subMap {
					_, m := checkKeys(val.(map[string]interface{}), nil, indent+"  ", root+"."+key)
					missingKeys = append(missingKeys, m...)
				} else {
					_k := strings.TrimPrefix(root+"."+key, ".")
					missingKeys = append(missingKeys, _k)
				}
			}
		}
	}
	if missing {
		return errors.New(""), missingKeys
	}
	return nil, missingKeys
}

func Execute() {
	if err := diffCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
