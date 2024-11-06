package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	ds "github.com/roundfeather/paven-config-tools/internal/datastore"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Sources struct {
	Envs map[string]struct {
		Project      string `yaml:"project"`
		ConfigTables []struct {
			Namespace string `yaml:"namespace"`
			Kind      string `yaml:"kind"`
			ConfigId  string `yaml:"config-id"`
		} `yaml:"config-tables"`
	} `yaml:"env"`
}

type diffStruct struct {
	Property interface{}
	Name     string
	Key      string
}

func main() {
	sourcesData, err := os.ReadFile("sources.yml")
	if err != nil {
		fmt.Println(err)
	}
	var sources Sources
	err = yaml.Unmarshal(sourcesData, &sources)
	if err != nil {
		fmt.Println(err)
	}

	env1 := os.Args[1]
	env2 := os.Args[2]

	ctx := context.Background()
	client1, err := datastore.NewClient(ctx, sources.Envs[env1].Project)
	if err != nil {
		fmt.Println(err)
	}
	client2, err := datastore.NewClient(ctx, sources.Envs[env2].Project)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create("report.md")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	f.WriteString("# Configuration Comparison\n")
	f.WriteString("\n")

	f.WriteString("## Summary\n")
	f.WriteString(fmt.Sprintf("- **Environment 1**: `%s`\n", env1))
	f.WriteString(fmt.Sprintf("- **Environment 2**: `%s`\n", env2))
	f.WriteString("\n")

	f.WriteString("## Table of Contents\n")
	for _, ct := range sources.Envs[env1].ConfigTables {
		f.WriteString(fmt.Sprintf("1. [%s](#%s)\n", ct.Kind, ct.Kind))
	}

	f.WriteString("\n")
	f.WriteString("---\n")
	f.WriteString("\n")

	for ctk, ct := range sources.Envs[env1].ConfigTables {
		key := ds.NewKey(ct.ConfigId)
		reader1 := ds.ReadConfig(ctx, client1, ct.Namespace, ct.Kind)
		reader2 := ds.ReadConfig(ctx, client2, sources.Envs[env2].ConfigTables[ctk].Namespace, sources.Envs[env2].ConfigTables[ctk].Kind)

		diffProperties := make([]diffStruct, 0)

		f.WriteString(fmt.Sprintf("## %s\n", ct.Kind))
		f.WriteString("\n")
		f.WriteString("### Overview\n")
		f.WriteString(fmt.Sprintf("- **Table Name**: `%s`\n", ct.Kind))
		f.WriteString(fmt.Sprintf("- **Total Entries Compared**: `%d`\n", len(reader1.Entities)))
		f.WriteString(fmt.Sprintf("- **Key Type**: `%s`\n", ct.ConfigId))

		f.WriteString("\n")
		f.WriteString("### Differences\n")
		f.WriteString("\n")

		f.WriteString(fmt.Sprintf("| **Key** | **Field** | **%s Value** | **%s Value** | **Difference** |\n", env1, env2))
		f.WriteString("|---------|-----------|-------------------------|-------------------------|------------|\n")

		for _, e := range reader1.Entities {
			k := key.Extract(e)

			e2, err := getEntity(k, key, reader2.Entities)
			if err != nil {
				fmt.Println(err)
			}

			for kp, p := range e.Properties {
				s := convertToMDString(p)

				var s2 string
				if e2.Properties == nil {
					s2 = convertToMDString("")
				} else {
					_p, ok := e2.Properties[kp]
					if ok {
						s2 = convertToMDString(_p)
					} else {
						s2 = convertToMDString("")
					}
				}

				diff := getDiffString(s, s2)
				if diff != "" {
					diffProperties = append(diffProperties, diffStruct{e.DatastoreProperties[kp], kp, k})
					diffProperties = append(diffProperties, diffStruct{e.DatastoreProperties[kp], kp, k})
				}

				f.WriteString(fmt.Sprintf("| %s | %s | <pre>%s</pre> | <pre>%s</pre> | <pre>%s</pre> |\n", k, kp, s, s2, diff))
			}
		}

		if len(diffProperties) > 0 {

			f.WriteString("\n")
			f.WriteString("### Update Values\n")
			f.WriteString("\n")

			f.WriteString("| **Key** | **Field** | **Value** |\n")
			f.WriteString("|---------|-----------|-----------|\n")

			for _, dp := range diffProperties {
				f.WriteString(fmt.Sprintf("| %s | %s | <pre>%s</pre> |\n", dp.Key, dp.Name, convertToMDString(dp.Property)))
			}
		}

		f.WriteString("\n")
		f.WriteString("---\n")
		f.WriteString("\n")
	}
}

func getEntity(keyLiteral string, key ds.Key, entities []ds.GenericEntiy) (ds.GenericEntiy, error) {
	for _, e := range entities {
		k := key.Extract(e)
		if k == keyLiteral {
			return e, nil
		}
	}
	return ds.GenericEntiy{}, errors.New("no entity with matching key")
}

func convertToMDString(property interface{}) string {
	data, _ := json.MarshalIndent(property, "", "&nbsp;")
	s := string(data)
	return strings.ReplaceAll(s, "\n", "<br>")
}

func getDiffString(env1String string, env2String string) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(env1String, env2String, false)

	diffString := ""

	for _, d := range diffs {
		if d.Type == diffmatchpatch.DiffEqual {
			diffString = fmt.Sprintf("%s%s", diffString, d.Text)
		} else if d.Type == diffmatchpatch.DiffDelete {
			diffString = fmt.Sprintf("%s<span style=\"color:red\">%s</span>", diffString, d.Text)
		} else {
			diffString = fmt.Sprintf("%s<span style=\"color:green\">%s</span>", diffString, d.Text)
		}
	}

	if diffString == env1String {
		diffString = ""
	}

	return diffString
}
