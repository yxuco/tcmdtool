package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.

Test command: ./tcmdtool import -i test-data/slack_events_api.json
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	input string
	root  string
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an API spec to TCMD",
	Long:  `Import an API spec to TCMD`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import", input)
		data, err := ioutil.ReadFile(input)
		if err != nil {
			panic(err)
		}
		// set root asset name as input file name if it is not specified
		if root == "" {
			fn := filepath.Base(input)
			root = fn[0:strings.Index(fn, ".")]
		}
		var spec map[string]interface{}
		if err = decode(data, &spec); err != nil {
			panic(err)
		}
		if spec["asyncapi"] != nil {
			fmt.Printf("Read asyncapi spec version %s\n", spec["asyncapi"])
			if err := importAsyncAPISpec(spec); err != nil {
				panic(err)
			}
		}
		if spec["openapi"] != nil {
			fmt.Printf("Read openapi spec version %s\n", spec["openapi"])
			if err := importOpenAPISpec(spec); err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&input, "input", "i", "", "name of the file to be imported")
	importCmd.Flags().StringVarP(&root, "root", "r", "", "root asset name to be created from input file")
	importCmd.MarkFlagRequired("input")
}

func get(path string, params map[string]string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s", url, path)
	reqAuth := fmt.Sprintf("Basic %s", authtoken)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create GET request %s", reqURL)
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	fmt.Printf("GET %s using token %s\n", req.URL, reqAuth)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", reqAuth)

	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed http GET %s", req.URL)
	}
	fmt.Printf("TCMD GET status: %d\n", resp.StatusCode)

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("HTTP GET returned status %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func post(path string, data interface{}) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s", url, path)
	reqAuth := fmt.Sprintf("Basic %s", authtoken)
	fmt.Printf("POST %s using token %s\n", reqURL, reqAuth)

	jsonReq, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize assets")
	}
	fmt.Println(string(jsonReq))

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create POST request %s", reqURL)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", reqAuth)
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed http POST %s", reqURL)
	}
	fmt.Printf("TCMD POST status: %d\n", resp.StatusCode)

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func decode(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		if err := yaml.Unmarshal(data, v); err != nil {
			return err
		}
	}
	return nil
}

// returns data type ID if it exists, 0 otherwise
func getAssetDataType(dataType string) int {
	path := fmt.Sprintf("%s/%s/datatype", TCDataspace, TCDataset)
	params := map[string]string{
		"predicate": fmt.Sprintf("name='%s'", dataType),
	}
	if resp, err := get(path, params); err == nil {
		var result []DataType
		if err = json.Unmarshal(resp, &result); err == nil {
			if len(result) > 0 {
				return result[0].ID
			}
		}
	}
	return 0
}

// find or create asset datatype by name, and return the ID
func findOrCreateAssetDataType(dataType string, complexType bool) (int, error) {
	tid := getAssetDataType(dataType)
	if tid > 0 {
		return tid, nil
	}
	return createAssetDataType(dataType, complexType)
}

// create new asset datatype by name, and return the ID
func createAssetDataType(dataType string, complexType bool) (int, error) {
	// create asset data type
	path := fmt.Sprintf("%s/%s/datatype", TCDataspace, TCDataset)
	data := DataType{
		Name:        dataType,
		Label:       dataType,
		BuiltIn:     false,
		ComplexType: complexType,
	}
	resp, err := post(path, data)
	if err != nil {
		return 0, err
	}
	var result DataType
	if err = json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// returns asset ID if it exists, 0 otherwise
func getAsset(name string) int {
	params := map[string]string{
		"predicate": fmt.Sprintf("name='%s'", name),
	}
	if resp, err := get("asset", params); err == nil {
		var result []Asset
		if err = json.Unmarshal(resp, &result); err == nil {
			if len(result) > 0 {
				return result[0].ID
			}
		}
	}
	return 0
}

// create or find asset by name, and return the ID
func createAsset(asset Asset) (int, error) {
	resp, err := post("asset", asset)
	if err != nil {
		return 0, err
	}
	var result Asset
	if err = json.Unmarshal(resp, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}
