// package main

// import (
// 	"os"

// 	"github.com/nxczje/vffuf/pkg"
// )

//	func main() {
//		//get input from user
//		file := os.Args[1]
//		//check if file exists and is a csv
//		check := pkg.Checkfile(file)
//		//process the file
//		if check {
//			pkg.Process_csv(file)
//		}
//	}
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kr/pretty"
)

type DirectoryTreeBuilder struct {
	DirectoryTree map[string]interface{}
}

func NewDirectoryTreeBuilder() *DirectoryTreeBuilder {
	return &DirectoryTreeBuilder{
		DirectoryTree: make(map[string]interface{}),
	}
}

func (dtb *DirectoryTreeBuilder) buildDirectoryTree(ffufOutput []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(ffufOutput, &data)
	if err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	results, ok := data["results"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid ffuf output format")
	}

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		url, ok := resultMap["url"].(string)
		if !ok || !strings.HasPrefix(url, "https://") {
			continue
		}

		parts := strings.Split(url, "/")[3:]
		dtb.addPathToTree(parts, resultMap)
	}
	return nil
}

func (dtb *DirectoryTreeBuilder) addPathToTree(parts []string, result map[string]interface{}) {
	currentLevel := dtb.DirectoryTree

	for i, part := range parts {
		if i == len(parts)-1 {
			status, _ := result["status"].(float64)
			length, _ := result["length"].(float64)
			currentLevel[part] = map[string]interface{}{"_status": strconv.Itoa(int(status)), "_length": strconv.Itoa(int(length))}
		} else {
			_, exists := currentLevel[part]
			if !exists {
				currentLevel[part] = make(map[string]interface{})
			}
			currentLevel = currentLevel[part].(map[string]interface{})
		}
	}
}

func (dtb *DirectoryTreeBuilder) printDirectoryTree(tree map[string]interface{}, indent int, lastItem bool) {
	keys := make([]string, 0, len(tree))
	for key := range tree {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		lastItemFlag := i == len(keys)-1
		symbol := "└── "
		if !lastItemFlag {
			symbol = "├── "
		}
		currentPath := fmt.Sprintf("%s%s", strings.Repeat("    ", indent), symbol)

		if subTree, ok := tree[key].(map[string]interface{}); ok {
			status, ok := subTree["_status"].(string)
			length, ok2 := subTree["_length"].(string)
			if ok && ok2 {
				currentPath += fmt.Sprintf("%s (Status: %s), (Length: %s)", key, status, length)
				delete(subTree, "_status")
			} else {
				currentPath += key
			}
			pretty.Println(currentPath)
			nextIndent := indent + 1
			lastItemChild := lastItemFlag && len(subTree) == 0
			dtb.printDirectoryTree(subTree, nextIndent, lastItemChild)
		}
	}
}

func readFFUFOutput(jsonFile string) ([]byte, error) {
	ffufOutput, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", jsonFile, err)
	}
	return ffufOutput, nil
}

func main() {
	jsonFile := flag.String("f", "", "Path to the JSON file containing ffuf output.")
	flag.Parse()

	if *jsonFile == "" {
		fmt.Println("Error: json_file flag is required.")
		flag.Usage()
		os.Exit(1)
	}

	ffufOutput, err := readFFUFOutput(*jsonFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	dtb := NewDirectoryTreeBuilder()
	err = dtb.buildDirectoryTree(ffufOutput)
	if err != nil {
		log.Fatalf("%v", err)
	}

	dtb.printDirectoryTree(dtb.DirectoryTree, 0, false)
}