package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.
*/
import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	url       string
	user      string
	password  string
	authtoken string
)

var (
	// TCDataspace TCMD dataspace name
	TCDataspace = "Tabula"
	// TCDataset TCMD dataset name
	TCDataset = "Tabula"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tcmdtool",
	Short: "Utility CLI for TIBCO Cloud Metadata",
	Long:  `Utility CLI for TIBCO Cloud Metadata`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .tcmdtool)")
	rootCmd.PersistentFlags().StringVar(&url, "url", "", "TCMD REST API URL of format https://${host}${basepath}/${dataspace}/${dataset}")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "TCMD technical user to invoke REST API")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password of TCMD technical user to invoke REST API")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType("yaml")
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tcmdtool" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".tcmdtool")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())

		// use parameters from config file if they are not specified in command-line
		if url == "" {
			url = fmt.Sprintf("%s%s", viper.Get("url"), viper.Get("basepath"))
		}
		fmt.Println("TCMD URL", url)

		if user == "" {
			user = viper.GetString("ebxuser")
		}
		fmt.Println("TCMD user name", user)

		if password == "" {
			password = viper.GetString("password")
		}

		if ds, ok := viper.Get("dataspace").(string); ok {
			TCDataspace = ds
		}

		if ds, ok := viper.Get("dataset").(string); ok {
			TCDataset = ds
		}
	} else {
		fmt.Printf("Viper failed to read config file %v\n", err)
	}

	// encode user:password to be used as service call header 'Authorization: Basic ${autotoken}'
	if user != "" && password != "" {
		authtoken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password)))
		fmt.Println("Basic auth token", authtoken)
	}
}

// Asset difines asset in TCMD
type Asset struct {
	ID            int    `json:"id,omitempty"`
	Name          string `json:"name"`
	Label         string `json:"label"`
	Description   string `json:"description,omitempty"`
	AssetType     string `json:"assetType"`
	AssetDataType string `json:"assetDataType,omitempty"`
	Logo          struct {
		Attachment string `json:"attachment"`
	} `json:"logo,omitempty"`
	Comment                 string `json:"comment,omitempty"`
	Parent                  string `json:"parent,omitempty"`
	Instance                string `json:"instance,omitempty"`
	DataElementAutoAssigned bool   `json:"dataElementAutoAssigned"`
	IsDisabled              bool   `json:"isDisabled"`
	Version                 string `json:"version,omitempty"`
}

// DataType defines asset data type in TCMD
type DataType struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	BuiltIn     bool   `json:"builtIn"`
	ComplexType bool   `json:"complexType"`
}
