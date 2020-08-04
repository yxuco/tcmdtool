package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.

Test command: ./tcmdtool export -r slack_events_api
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	output string
	format string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export API spec from TCMD",
	Long:  `export API spec from TCMD`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("export", root)

		// set output file name to match root element name if the file is not specified
		if output == "" {
			output = fmt.Sprintf("%s.%s", root, format)
		}

		spec, err := exportAsyncAPISpec(root)
		if err != nil {
			panic(err)
		}
		data, err := encode(spec)
		if err != nil {
			panic(err)
		}
		ioutil.WriteFile(output, data, 0644)
		fmt.Println("API spec exported in file", output)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&root, "root", "r", "", "name of root asset to be exported")
	exportCmd.Flags().StringVarP(&input, "output", "o", "", "name of the spec file to be exported")
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "output file format, json or yaml")
	exportCmd.MarkFlagRequired("root")
}

// fetch asset data type of a specified ID
func getAssetDataTypeByID(id int) (*DataType, error) {
	path := fmt.Sprintf("%s/%s/datatype/%d", TCDataspace, TCDataset, id)
	resp, err := get(path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed TCMD request")
	}

	var result DataType
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal TCMD response")
	}
	return &result, nil
}

// fetch asset of a specified ID
func getAssetByID(id int) (*Asset, error) {
	path := fmt.Sprintf("asset/%d", id)
	resp, err := get(path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed TCMD request")
	}

	var result Asset
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal TCMD response")
	}
	return &result, nil
}

// fetch asset of a specified name
func getAssetByName(name string) (*Asset, error) {
	params := map[string]string{
		"predicate": fmt.Sprintf("name='%s'", name),
	}
	resp, err := get("asset", params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed TCMD request")
	}

	var result []Asset
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal TCMD response")
	}

	if len(result) > 0 {
		return &result[0], nil
	}

	return nil, nil
}

// fetch children assets of a specified parent
func getChildrenAsset(id int) ([]Asset, error) {
	params := map[string]string{
		"predicate": fmt.Sprintf("parent='%d'", id),
	}
	resp, err := get("asset", params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed TCMD request")
	}

	var result []Asset
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal TCMD response")
	}

	if len(result) > 0 {
		return result, nil
	}
	return nil, nil
}

func encode(data interface{}) ([]byte, error) {
	if format == "yaml" {
		return yaml.Marshal(data)
	}
	return json.MarshalIndent(data, "", "    ")
}
