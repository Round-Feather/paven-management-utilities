package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/go-test/deep"
	"github.com/manifoldco/promptui"
	"google.golang.org/api/iterator"
	"gopkg.in/yaml.v2"
)

// ANSI color codes for log messages and output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[1;31m"
	colorGreen  = "\033[1;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[1;34m"
	colorCyan   = "\033[1;36m"
)

// Config holds the overall configuration structure
type Config struct {
	ProjectID string       `yaml:"projectID"`
	Kinds     []KindConfig `yaml:"kinds"`
}

// KindConfig holds configuration for each kind and its namespace
type KindConfig struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// OutputEntity represents the simplified JSON output for each entity
type OutputEntity struct {
	ID     string                 `json:"id"`
	Parent string                 `json:"parent,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// Clear the console (platform-independent)
func clearConsole() {
	cmd := exec.Command("clear")
	if err := cmd.Run(); err != nil {
		fmt.Println("Unable to clear the console.")
	}
}

// Display the tool title and description
func displayHeader() {
	clearConsole()
	fmt.Println(colorBlue + "===============================")
	fmt.Println("      Datastore CLI Tool")
	fmt.Println("===============================")
	fmt.Println("Welcome to the Datastore CLI tool! This tool allows you to download, compare, and apply changes to your Google Cloud Datastore.")
	fmt.Println("Use the options below to select your desired operation:\n" + colorReset)
}

// Log functions with color coding
func logError(message string) {
	fmt.Println(colorRed + "ERROR: " + message + colorReset)
}

func logInfo(message string) {
	fmt.Println(colorYellow + "INFO: " + message + colorReset)
}

func logSuccess(message string) {
	fmt.Println(colorGreen + "SUCCESS: " + message + colorReset)
}

func main() {
	// Parse flags for YAML configuration file and output directory
	configPath := flag.String("config", "config.yaml", "Path to YAML configuration file")
	outputDir := flag.String("outputDir", "./output", "Directory to save JSON output files")
	flag.Parse()

	// Display header and clear console
	displayHeader()

	// Interactive menu
	prompt := promptui.Select{
		Label: "Select an action",
		Items: []string{"Only Download", "Download and Compare", "Apply Changes to Database"},
		Templates: &promptui.SelectTemplates{
			Selected: "\U0001F4CC " + colorCyan + "{{ . }}" + colorReset,
			Active:   colorGreen + "\U0001F4CC {{ . }}" + colorReset,
			Inactive: colorYellow + "  {{ . }}" + colorReset,
		},
	}

	_, choice, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to load configuration: %v", err))
		return
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		logError(fmt.Sprintf("Failed to create output directory: %v", err))
		return
	}

	switch choice {
	case "Only Download":
		logInfo("Starting download...")
		if err := retrieveAndSaveJSON(config, *outputDir); err != nil {
			logError(fmt.Sprintf("Error retrieving datastore data: %v", err))
			return
		}
		logSuccess("Data downloaded successfully.")

	case "Download and Compare":
		logInfo("Starting download...")
		if err := retrieveAndSaveJSON(config, *outputDir); err != nil {
			logError(fmt.Sprintf("Error retrieving datastore data: %v", err))
			return
		}
		logSuccess("Data downloaded successfully.")

		// Prompt for comparison directory
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(colorCyan + "Enter the directory path to compare against (default is ./local_changes): " + colorReset)
		compareDir, _ := reader.ReadString('\n')
		compareDir = strings.TrimSpace(compareDir)
		if compareDir == "" {
			compareDir = "./local_changes"
		}

		if err := compareOutput(*outputDir, compareDir); err != nil {
			logError(fmt.Sprintf("Error comparing output files: %v", err))
		}

	case "Apply Changes to Database":
		// Prompt for dry-run mode (Yes by default)
		dryRunPrompt := promptui.Select{
			Label: "Enable dry-run mode? (no changes will be applied to database)",
			Items: []string{"Yes", "No"},
			Templates: &promptui.SelectTemplates{
				Selected: colorCyan + "Dry-run: {{ . }}" + colorReset,
				Active:   colorGreen + "\U0001F4CC {{ . }}" + colorReset,
				Inactive: colorYellow + "  {{ . }}" + colorReset,
			},
		}
		_, dryRunChoice, err := dryRunPrompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed %v\n", err)
		}

		dryRun := dryRunChoice == "Yes"

		// Prompt for directory containing changes to apply
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(colorCyan + "Enter the directory path containing changes to apply (default is ./local_changes/): " + colorReset)
		applyDir, _ := reader.ReadString('\n')
		applyDir = strings.TrimSpace(applyDir)
		if applyDir == "" {
			applyDir = "./local_changes/"
		}

		if dryRun {
			logInfo("Dry-run mode enabled. Changes will not be applied to the database.")
		}

		if err := applyChangesToDatabase(config.ProjectID, applyDir, dryRun); err != nil {
			logError(fmt.Sprintf("Error applying changes to database: %v", err))
		} else {
			if dryRun {
				logSuccess("Dry-run completed. JSON output generated for review.")
			} else {
				logSuccess("Changes applied to the database successfully.")
			}
		}

	default:
		logError("Invalid choice. Please run the program again.")
	}
}

// loadConfig loads the yaml config to be used to obtain the datastore data
func loadConfig(path string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func applyChangesToDatabase(projectID, applyDir string, dryRun bool) error {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create datastore client: %v", err)
	}
	defer client.Close()

	dryRunDir := filepath.Join("local_changes", "dry_run")

	if dryRun {
		if _, err := os.Stat(dryRunDir); err == nil {
			if err := os.RemoveAll(dryRunDir); err != nil {
				return fmt.Errorf("failed to clear dry_run directory: %v", err)
			}
		}
		if err := os.MkdirAll(dryRunDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create dry_run directory: %v", err)
		}
	}

	namespaces, err := ioutil.ReadDir(applyDir)
	if err != nil {
		return fmt.Errorf("failed to read apply directory: %v", err)
	}

	for _, ns := range namespaces {
		if ns.Name() == "dry_run" || !ns.IsDir() {
			continue
		}

		namespaceDir := filepath.Join(applyDir, ns.Name())
		dryRunNamespaceDir := filepath.Join(dryRunDir, ns.Name())
		if dryRun {
			if err := os.MkdirAll(dryRunNamespaceDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create namespace directory %s: %v", dryRunNamespaceDir, err)
			}
		}

		files, err := ioutil.ReadDir(namespaceDir)
		if err != nil {
			logError(fmt.Sprintf("Failed to read namespace directory %s: %v", namespaceDir, err))
			continue
		}

		for _, file := range files {
			kind := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			filePath := filepath.Join(namespaceDir, file.Name())

			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				logError(fmt.Sprintf("Error reading file %s: %v", filePath, err))
				continue
			}

			var entities []OutputEntity
			if err := json.Unmarshal(data, &entities); err != nil {
				logError(fmt.Sprintf("Error unmarshalling JSON in file %s: %v", filePath, err))
				continue
			}

			var changesForKind []datastore.Entity

			// Loop through each entity in the JSON data to prepare for database actions
			// Loop through each entity in the JSON data to prepare for database actions
			for _, entity := range entities {
				var key, parentKey *datastore.Key

				// Handle parent key creation
				if entity.Parent != "" {
					parentArr := strings.Split(entity.Parent, ",")
					if parentID, err := strconv.ParseInt(parentArr[1], 10, 64); err == nil {
						parentKey = datastore.IDKey(parentArr[0], parentID, nil)
					} else {
						parentKey = datastore.NameKey(parentArr[0], parentArr[1], nil)
					}
					parentKey.Namespace = ns.Name()
				}

				// Determine the appropriate key based on entity.ID presence
				if entity.ID == "" {
					// If entity.ID is empty, use IncompleteKey to create a new entity with a generated ID
					key = datastore.IncompleteKey(kind, parentKey)
				} else {
					// If entity.ID is present, use it to define a specific key
					if idInt, err := strconv.ParseInt(entity.ID, 10, 64); err == nil {
						key = datastore.IDKey(kind, idInt, parentKey)
					} else {
						key = datastore.NameKey(kind, entity.ID, parentKey)
					}
				}
				key.Namespace = ns.Name()

				// Skip datastore fetch if entity.ID is empty (new entity)
				var existingDataMap map[string]interface{}
				var diff []string
				// Prepare the new data map for comparison or new entity creation
				newDataMap := make(map[string]interface{})
				for k, v := range entity.Data {
					newDataMap[k] = simplifyValue(v)
				}
				if entity.ID != "" {
					var existingData datastore.PropertyList
					err := client.Get(ctx, key, &existingData)
					if err != nil && err != datastore.ErrNoSuchEntity {
						logError(fmt.Sprintf("Error fetching entity with ID %s from Datastore: %v", entity.ID, err))
						continue
					}
					existingDataMap = propertyListToMap(existingData)
				}

				diff = deep.Equal(existingDataMap, newDataMap)

				// In dry-run mode, log all changes (including new entities)
				if dryRun && (entity.ID == "" || len(diff) > 0) {
					logInfo(fmt.Sprintf("Dry-run: preparing entity for creation/update with ID %s. Differences: %v", entity.ID, diff))
					properties, _ := mapToPropertyList(newDataMap)
					changesForKind = append(changesForKind, datastore.Entity{
						Key:        key,
						Properties: properties,
					})
				}

				// Apply changes in non-dry-run mode only if there are differences
				if !dryRun && len(diff) > 0 {
					logInfo(fmt.Sprintf("Applying updates for entity with ID %s", entity.ID))
					properties, err := mapToPropertyList(newDataMap)
					if err != nil {
						logError(fmt.Sprintf("Error converting data map for entity with ID %s: %v", entity.ID, err))
						continue
					}
					if _, err := client.Put(ctx, key, &properties); err != nil {
						logError(fmt.Sprintf("Error applying entity with ID %s in file %s: %v", entity.ID, filePath, err))
					} else {
						logInfo(fmt.Sprintf("Updated entity with ID %s in Datastore", entity.ID))
					}
				}
			}

			// Write out dry-run file if there are changes
			if dryRun && len(changesForKind) > 0 {
				dryRunFile := filepath.Join(dryRunNamespaceDir, fmt.Sprintf("%s_dry_run.json", kind))
				jsonData, err := json.MarshalIndent(changesForKind, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal dry-run output for kind %s: %v", kind, err)
				}

				if err := ioutil.WriteFile(dryRunFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write dry-run JSON output for kind %s: %v", kind, err)
				}
				logInfo(fmt.Sprintf("Dry-run output saved to %s", dryRunFile))
			}

			if dryRun && len(changesForKind) > 0 {
				dryRunFile := filepath.Join(dryRunNamespaceDir, fmt.Sprintf("%s_dry_run.json", kind))
				jsonData, err := json.MarshalIndent(changesForKind, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal dry-run output for kind %s: %v", kind, err)
				}

				if err := ioutil.WriteFile(dryRunFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write dry-run JSON output for kind %s: %v", kind, err)
				}
				logInfo(fmt.Sprintf("Dry-run output saved to %s", dryRunFile))
			}
		}
	}

	return nil
}

func mapToPropertyList(data map[string]interface{}) (datastore.PropertyList, error) {
	var properties datastore.PropertyList

	for key, value := range data {
		convertedValue, err := convertValueToDatastore(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property %s: %v", key, err)
		}
		properties = append(properties, datastore.Property{
			Name:  key,
			Value: convertedValue,
		})
	}
	return properties, nil
}

func convertValueToDatastore(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		// For nested maps, convert to a datastore.Entity
		propertyList, err := mapToPropertyList(v)
		if err != nil {
			return nil, err
		}
		return &datastore.Entity{Properties: propertyList}, nil
	case []interface{}:
		// For slices, recursively convert each item
		var items []interface{}
		for _, item := range v {
			convertedItem, err := convertValueToDatastore(item)
			if err != nil {
				return nil, err
			}
			items = append(items, convertedItem)
		}
		return items, nil
	case int64, string, bool, float64:
		// Return primitive types as-is
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported property type: %T", v)
	}
}

// Convert PropertyList to a simplified map[string]interface{} (used for comparison)
func propertyListToMap(pl datastore.PropertyList) map[string]interface{} {
	dataMap := make(map[string]interface{})
	for _, prop := range pl {
		dataMap[prop.Name] = simplifyValue(prop.Value)
	}
	return dataMap
}

// Recursive function to simplify complex types into JSON-friendly values
func simplifyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case float64:
		// Convert float64 to int64 if it represents a whole number
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	case map[string]interface{}:
		// Recursively simplify map values
		simplifiedMap := make(map[string]interface{})
		for key, val := range v {
			simplifiedMap[key] = simplifyValue(val)
		}
		return simplifiedMap
	case []interface{}:
		// Recursively simplify array elements
		for i := range v {
			v[i] = simplifyValue(v[i])
		}
		return v
	case datastore.PropertyList:
		return propertyListToMap(v)
	case *datastore.Entity:
		return propertyListToMap(v.Properties)
	case []*datastore.Key:
		var keys []string
		for _, key := range v {
			keys = append(keys, key.String())
		}
		return keys
	default:
		return v
	}
}

// Retrieves the entity ID from the datastore.Key as a string
func getEntityID(key *datastore.Key) string {
	if key.ID != 0 {
		return fmt.Sprintf("%d", key.ID) // Use numeric ID
	}
	return key.Name // Use named ID if available
}

// Retrieve parent key as a formatted string with both kind and ID if available
func getParentKeyString(key *datastore.Key) string {
	if key.Parent != nil {
		// Extract both kind and ID for the parent
		kind := key.Parent.Kind
		id := getEntityID(key.Parent) // Get only the ID part of the parent
		return fmt.Sprintf("%s,%s", kind, id)
	}
	return ""
}

func retrieveAndSaveJSON(config Config, outputDir string) error {
	ctx := context.Background()

	// Initialize Datastore client
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to create datastore client: %v", err)
	}
	defer client.Close()

	// Query each kind separately with the specified namespace
	for _, kindConfig := range config.Kinds {
		var outputEntities []OutputEntity
		query := datastore.NewQuery(kindConfig.Name).Namespace(kindConfig.Namespace)

		// Fetch entities
		it := client.Run(ctx, query)
		for {
			var data datastore.PropertyList
			key, err := it.Next(&data)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("error retrieving entity from kind %s in namespace %s: %v", kindConfig.Name, kindConfig.Namespace, err)
			}

			// Create OutputEntity with the entity's ID, Parent kind and ID, and Data
			outputEntity := OutputEntity{
				ID:     getEntityID(key),        // Extract only the specific entity ID part
				Parent: getParentKeyString(key), // Include both kind and ID for the parent entity
				Data:   propertyListToMap(data),
			}
			outputEntities = append(outputEntities, outputEntity)
		}

		// Create namespace directory within outputDir
		namespaceDir := filepath.Join(outputDir, kindConfig.Namespace)
		if err := os.MkdirAll(namespaceDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create namespace directory %s: %v", namespaceDir, err)
		}

		// Save JSON to file with kind name in the namespace folder
		filePath := filepath.Join(namespaceDir, fmt.Sprintf("%s.json", kindConfig.Name))
		jsonData, err := json.MarshalIndent(outputEntities, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal entities for kind %s to JSON: %v", kindConfig.Name, err)
		}
		if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write JSON file for kind %s: %v", kindConfig.Name, err)
		}
		logInfo(fmt.Sprintf("Data for kind '%s' (namespace '%s') saved to %s", kindConfig.Name, kindConfig.Namespace, filePath))
	}

	return nil
}

// compareOutput compares the generated JSON files in outputDir with the JSON files in compareDir
// and displays differences on a line-by-line basis.
func compareOutput(outputDir, compareDir string) error {
	namespaces, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %v", err)
	}

	for _, ns := range namespaces {
		if !ns.IsDir() {
			continue
		}

		namespaceDir := filepath.Join(outputDir, ns.Name())
		compareNamespaceDir := filepath.Join(compareDir, ns.Name())
		files, err := ioutil.ReadDir(namespaceDir)
		if err != nil {
			logError(fmt.Sprintf("Failed to read namespace directory %s: %v", namespaceDir, err))
			continue
		}

		for _, file := range files {
			outputFilePath := filepath.Join(namespaceDir, file.Name())
			compareFilePath := filepath.Join(compareNamespaceDir, file.Name())

			outputData, err := ioutil.ReadFile(outputFilePath)
			if err != nil {
				logError(fmt.Sprintf("Error reading output file %s: %v", outputFilePath, err))
				continue
			}

			compareData, err := ioutil.ReadFile(compareFilePath)
			if err != nil {
				logError(fmt.Sprintf("Comparison file missing or unreadable: %s", compareFilePath))
				continue
			}

			outputJSON, err := prettyPrintJSON(outputData)
			if err != nil {
				logError(fmt.Sprintf("Error formatting output JSON in file %s: %v", outputFilePath, err))
				continue
			}
			compareJSON, err := prettyPrintJSON(compareData)
			if err != nil {
				logError(fmt.Sprintf("Error formatting comparison JSON in file %s: %v", compareFilePath, err))
				continue
			}

			logInfo(fmt.Sprintf("Comparing file: %s", file.Name()))
			displayLineDiff(outputJSON, compareJSON)
			fmt.Println()
		}
	}
	return nil
}

// displayLineDiff compares two JSON strings line by line, highlighting differences in color.
func displayLineDiff(outputJSON, compareJSON string) {
	outputLines := strings.Split(outputJSON, "\n")
	compareLines := strings.Split(compareJSON, "\n")

	added, deleted, changed := 0, 0, 0
	for i := 0; i < len(outputLines) || i < len(compareLines); i++ {
		var outputLine, compareLine string
		if i < len(outputLines) {
			outputLine = outputLines[i]
		}
		if i < len(compareLines) {
			compareLine = compareLines[i]
		}

		if outputLine != compareLine {
			if compareLine == "" {
				logDiff("ADDED", outputLine, colorGreen)
				fmt.Printf("\n%sLine: %d%s\n", colorGreen, i+3, colorReset)
				added++
			} else if outputLine == "" {
				logDiff("DELETED", compareLine, colorRed)
				fmt.Printf("\n%sLine: %d%s\n", colorRed, i+3, colorReset)
				deleted++
			} else {
				logDiff("CHANGED", fmt.Sprintf("Expected: %s", compareLine), colorRed)
				logDiff("\nCHANGED", fmt.Sprintf("Got:      %s", outputLine), colorYellow)
				fmt.Printf("\n%sLine: %d%s\n", colorRed, i+3, colorReset)
				changed++
			}
		}
	}

	// Summary of differences
	fmt.Printf("\n%sSummary: %d added, %d deleted, %d changed%s\n\n",
		colorBlue, added, deleted, changed, colorReset)
}

// logDiff formats and colors diff messages based on type.
func logDiff(changeType, message, colorCode string) {
	fmt.Printf("%s%s: %s%s", colorCode, changeType, message, colorReset)
}

// prettyPrintJSON formats JSON data to a pretty-printed string with consistent formatting.
func prettyPrintJSON(data []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	prettyData, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %v", err)
	}
	return string(prettyData), nil
}
