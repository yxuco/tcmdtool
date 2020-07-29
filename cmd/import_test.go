package cmd

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testid = 211
)

func init() {
	cfgFile = "../.tcmdtool"
	initConfig()
}

func TestTCMDPostAsset(t *testing.T) {
	asset := Asset{
		Name:        "test-api",
		Label:       "test-api",
		Description: "test REST API from import tool",
		AssetType:   "24",
		Logo: struct {
			Attachment string `json:"attachment"`
		}{
			Attachment: "",
		},
		DataElementAutoAssigned: false,
		IsDisabled:              false,
		Version:                 "1.0.0",
	}
	resp, err := post("asset", asset)
	assert.NoError(t, err, "POST asset should not return error %v", err)
	assert.NotNil(t, resp, "POST asset should not return nil")
	fmt.Println(string(resp))

	var result Asset
	err = json.Unmarshal(resp, &result)
	assert.NoError(t, err, "POST asset result is not a valid asset %v", err)

	testid = result.ID
	assert.Equal(t, "test-api", result.Label, "Asset label does not match")
}

func TestTCMDGet(t *testing.T) {
	path := fmt.Sprintf("asset/%d", testid)

	resp, err := get(path, nil)
	assert.NoError(t, err, "GET %s should not return error %v", path, err)
	assert.NotNil(t, resp, "GET %s should not return nil", path)
	fmt.Println(string(resp))

	var asset Asset
	err = json.Unmarshal(resp, &asset)
	assert.NoError(t, err, "GET %s result is not a valid asset %v", path, err)

	assert.Equal(t, testid, asset.ID, "Asset ID does not match")
	assert.Equal(t, "test-api", asset.Label, "Asset label does not match")
}

func TestTCMDPostDataType(t *testing.T) {
	data := DataType{
		Name:        "test-type",
		Label:       "test-type",
		Description: "test asset data type from import tool",
		BuiltIn:     false,
		ComplexType: false,
	}
	resp, err := post(fmt.Sprintf("%s/%s/datatype", TCDataspace, TCDataset), data)
	assert.NoError(t, err, "POST data type should not return error %v", err)
	assert.NotNil(t, resp, "POST data type should not return nil")
	fmt.Println(string(resp))

	var result DataType
	err = json.Unmarshal(resp, &result)
	assert.NoError(t, err, "POST data type result is not a valid JSON %v", err)

	assert.Lessf(t, 0, result.ID, "New data type ID $d should be greater than 0", result.ID)
	assert.Equal(t, "test-type", result.Label, "Data type label does not match")
}

func TestTCMDQuery(t *testing.T) {
	path := fmt.Sprintf("%s/%s/datatype", TCDataspace, TCDataset)
	params := map[string]string{
		"predicate": "name='Undefined'",
	}
	resp, err := get(path, params)
	assert.NoError(t, err, "QUERY %s should not return error %v", path, err)
	assert.NotNil(t, resp, "QUERY %s should not return nil", path)
	fmt.Println(string(resp))

	var result []DataType
	err = json.Unmarshal(resp, &result)
	assert.NoError(t, err, "QUERY %s result is not a valid array %v", path, err)
	assert.Equal(t, 999, result[0].ID, "data type ID does not match")
}

func TestExtractProperties(t *testing.T) {
	js := `{
        "additionalProperties": true,
        "properties": {
        	"event_ts": {
            	"title": "When the event was dispatched",
            	"type": "string"
            },
            "type": {
                "title": "The specific name of the event",
                "type": "string"
            }
        },
        "required": ["type", "event_ts"],
        "title": "The actual event, an object, that happened",
        "type": "object",
        "x-examples": [{
            "channel": "D0PNCRP9N",
            "channel_type": "app_home",
            "event_ts": "1525215129.000001",
            "text": "How many cats did we herd yesterday?",
            "ts": "1525215129.000001",
            "type": "message",
            "user": "U061F7AUR"
        }]
	}`
	data := make(map[string]interface{})
	json.Unmarshal([]byte(js), &data)
	result := extractExtraProperties(data, []string{"properties", "type", "x-examples"})
	expected := `{
    "additionalProperties": true,
    "required": [
        "type",
        "event_ts"
    ],
    "title": "The actual event, an object, that happened"
}`
	assert.Equal(t, expected, result, "data does not match")
}
