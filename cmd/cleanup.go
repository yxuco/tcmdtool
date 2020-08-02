package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.

Test command: ./tcmdtool clean -i test-data/slack_events_api.json
*/

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Cleanup an API spec in TCMD",
	Long:  `Cleanup an API spec in TCMD`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("clean", input)
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
			if err := cleanAsyncAPISpec(spec); err != nil {
				panic(err)
			}
		}
		if spec["openapi"] != nil {
			fmt.Printf("Read openapi spec version %s\n", spec["openapi"])
			if err := cleanOpenAPISpec(spec); err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVarP(&input, "input", "i", "", "name of the file to be cleaned")
	cleanCmd.Flags().StringVarP(&root, "root", "r", "", "name of root asset created from input file")
	cleanCmd.MarkFlagRequired("input")
}

func delete(path string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s", url, path)
	reqAuth := fmt.Sprintf("Basic %s", authtoken)

	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create DELETE request %s", reqURL)
	}
	fmt.Printf("DELETE %s using token %s\n", req.URL, reqAuth)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", reqAuth)

	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed http DELETE %s", req.URL)
	}
	fmt.Printf("TCMD DELETE status: %d\n", resp.StatusCode)

	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.Errorf("HTTP DELETE returned status %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

// delete asset data type of specified ID
func deleteAssetDataType(tid int) error {
	path := fmt.Sprintf("%s/%s/datatype/%d", TCDataspace, TCDataset, tid)
	_, err := delete(path)
	return err
}

// delete asset of specified ID
func deleteAsset(tid int) error {
	path := fmt.Sprintf("asset/%d", tid)
	_, err := delete(path)
	return err
}
