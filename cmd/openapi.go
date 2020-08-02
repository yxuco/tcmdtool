package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.
*/
import (
	"fmt"

	"github.com/pkg/errors"
)

func importOpenAPISpec(spec map[string]interface{}) error {
	return importAPIPaths(spec)
}

func importAPIPaths(spec map[string]interface{}) error {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return errors.New("paths are not defined in openapi spec")
	}
	for k, v := range paths {
		ops := v.(map[string]interface{})
		fmt.Printf("import path %s - ", k)
		for o := range ops {
			fmt.Print(o, " ")
		}
		fmt.Println()
	}
	return nil
}

func cleanOpenAPISpec(spec interface{}) error {
	fmt.Println("clean OpenAPI spec not implemented")
	return nil
}
