package cmd

/*
Copyright © 2020 Yueming Xu <yxu@tibco.com>
This file is subject to the license terms contained in the license file that is distributed with this file.
*/

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// AssetTypes maps type --> type ID
var AssetTypes = map[string]string{
	"JSON Element":  "24",
	"JSON Property": "25",
}

// AssetDataTypes maps dataType --> ID
var AssetDataTypes map[string]int

func initializeAssetDataTypes() error {
	AssetDataTypes = make(map[string]int)
	basicTypes := []string{"string", "integer", "boolean", "array"}
	for _, t := range basicTypes {
		id, err := findOrCreateAssetDataType(t, false)
		if err != nil {
			return err
		}
		AssetDataTypes[t] = id
	}
	return nil
}

func cleanAsyncAPISpec(spec interface{}) error {
	if rid := getAsset(root); rid > 0 {
		// remove root asset
		fmt.Printf("cleanup asset %d -> %s\n", rid, root)
		deleteAsset(rid)
	}

	if components := getRef(spec, "#/components"); components != nil {
		cm, ok := components.(map[string]interface{})
		if !ok {
			return errors.Errorf("components type %T is not a map", components)
		}
		for cat, val := range cm {
			om, ok := val.(map[string]interface{})
			if !ok {
				continue
			}
			for k := range om {
				if tid := getAssetDataType(fmt.Sprintf("#/components/%s/%s", cat, k)); tid > 0 {
					// remove asset data types
					fmt.Printf("cleanup data type %d -> %s\n", tid, k)
					deleteAssetDataType(tid)
				}
			}
		}
	}
	return nil
}

func importAsyncAPISpec(spec map[string]interface{}) error {
	if err := initializeAssetDataTypes(); err != nil {
		return err
	}

	rid, err := createAsyncAPIAsset(spec)
	if err != nil {
		return nil
	}
	if asyncapi, ok := spec["asyncapi"]; ok {
		createSimpleAsset("asyncapi", asyncapi.(string), rid, "string")
	}
	if id, ok := spec["id"]; ok {
		createSimpleAsset("id", id.(string), rid, "string")
	}

	if info, ok := spec["info"]; ok {
		createInfoAsset(info, rid)
	}

	if components, ok := spec["components"]; ok {
		createComponentsAsset(components, rid)
	}

	if servers, ok := spec["servers"]; ok {
		createServersAsset(servers, rid)
	}

	if channels, ok := spec["channels"].(map[string]interface{}); ok {
		createChannelsAsset(channels, rid)
	}

	if tags, ok := spec["tags"]; ok {
		createTagsAsset(tags, rid)
	}

	if externalDocs, ok := spec["externalDocs"]; ok {
		createExternalDocsAsset(externalDocs, rid)
	}
	return nil
}

func createAsyncAPIAsset(doc map[string]interface{}) (int, error) {
	comment := extractExtraProperties(doc, []string{"id", "asyncapi", "info", "externalDocs", "tags", "components", "channels", "servers"})
	asset := Asset{
		Name:                    root,
		Label:                   root,
		Comment:                 comment,
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	return createAsset(asset)
}

func createSimpleAsset(name, value string, parent int, dataType string) (int, error) {
	if len(value) == 0 {
		// no value, so do not create it
		return 0, nil
	}
	asset := Asset{
		Name:                    name,
		Label:                   name,
		AssetType:               AssetTypes["JSON Element"],
		Comment:                 value,
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if dataType == "string" {
		asset.AssetDataType = strconv.Itoa(AssetDataTypes["string"])
	}
	if parent > 0 {
		asset.Parent = strconv.Itoa(parent)
	}
	return createAsset(asset)
}

func createInfoAsset(info interface{}, parent int) error {
	comment := extractExtraProperties(info.(map[string]interface{}), []string{"description", "contact", "version"})
	asset := Asset{
		Name:                    "info",
		Label:                   "info",
		Description:             getString(info, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}
	createSimpleAsset("version", getString(info, "#/version"), pid, "string")
	if contact := getRef(info, "#/contact"); contact != nil {
		if value, err := json.MarshalIndent(contact, "", "    "); err == nil {
			createSimpleAsset("contact", string(value), pid, "")
		}
	}
	return nil
}

func createExternalDocsAsset(docs interface{}, parent int) error {
	comment := extractExtraProperties(docs.(map[string]interface{}), []string{"description"})
	asset := Asset{
		Name:                    "externalDocs",
		Label:                   "externalDocs",
		Description:             getString(docs, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	_, err := createAsset(asset)
	return err
}

func createTagsAsset(tags interface{}, parent int) error {
	asset := Asset{
		Name:                    "tags",
		Label:                   "tags",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(AssetDataTypes["array"]),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}
	if tagList, ok := tags.([]interface{}); ok {
		for _, tag := range tagList {
			name := getString(tag, "#/name")
			asset := Asset{
				Name:                    name,
				Label:                   name,
				Description:             getString(tag, "#/description"),
				Parent:                  strconv.Itoa(pid),
				AssetType:               AssetTypes["JSON Property"],
				DataElementAutoAssigned: false,
				IsDisabled:              false,
			}
			createAsset(asset)
		}
	}
	return nil
}

func createComponentsAsset(components interface{}, parent int) error {
	asset := Asset{
		Name:                    "components",
		Label:                   "components",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}
	cm, ok := components.(map[string]interface{})
	if !ok {
		return errors.Errorf("components type %T is not a map", components)
	}
	for cat, list := range cm {
		asset := Asset{
			Name:                    cat,
			Label:                   cat,
			Parent:                  strconv.Itoa(pid),
			AssetType:               AssetTypes["JSON Element"],
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		cid, err := createAsset(asset)
		if err != nil {
			continue
		}
		om, ok := list.(map[string]interface{})
		if !ok {
			continue
		}
		// create reusable data types
		for k, v := range om {
			tid := setRef(fmt.Sprintf("#/components/%s/%s", cat, k))

			switch cat {
			case "schemas":
				createSchemaAsset(k, v, tid, cid, false)
			case "messages":
				createMessageAsset(k, v, tid, cid)
			case "securitySchemes":
				createSecuritySchemeAsset(k, v, tid, cid)
			case "parameters":
				createParameterAsset(k, v, tid, cid)
			case "operationTraits":
				createOperationTraitAsset(k, v, tid, cid)
			case "messageTraits":
				createMessageTraitAsset(k, v, tid, cid)
			default:
				fmt.Printf("component type %s not implemented", cat)
			}

		}
	}
	return nil
}

func createSchemaAsset(name string, data interface{}, tid int, parent int, isProperty bool) error {
	comment := extractExtraProperties(data.(map[string]interface{}), []string{"$ref", "description", "properties", "x-examples", "examples"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(data, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if isProperty {
		asset.AssetType = AssetTypes["JSON Property"]
	} else {
		asset.AssetType = AssetTypes["JSON Element"]
	}
	dtid := tid
	if tid == 0 {
		// not a component type, so set primitive data type
		dtype := getString(data, "#/type")
		if len(dtype) > 0 && dtype != "object" {
			if t, ok := AssetDataTypes[dtype]; ok {
				dtid = t
			}
		}
	}
	if dtid > 0 {
		asset.AssetDataType = strconv.Itoa(dtid)
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if props := getRef(data, "#/properties"); props != nil {
		if pm, ok := props.(map[string]interface{}); ok {
			for k, v := range pm {
				ctid := 0
				if ref := getString(v, "#/$ref"); len(ref) > 0 {
					ctid = setRef(ref)
				}
				//TODO: array type is assumed as simple primitive types
				createSchemaAsset(k, v, ctid, pid, true)
				//				createPropertyAsset(k, v, pid)
			}
		}
	}
	return nil
}

// return JSON of data excluding specified properties
func extractExtraProperties(data map[string]interface{}, exclude []string) string {
	exmap := make(map[string]bool)
	for _, v := range exclude {
		exmap[v] = true
	}
	result := make(map[string]interface{})
	for k, v := range data {
		if _, ok := exmap[k]; !ok {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return ""
	}
	props, _ := json.MarshalIndent(result, "", "    ")
	return string(props)
}

func createChannelsAsset(channels map[string]interface{}, parent int) error {
	asset := Asset{
		Name:                    "channels",
		Label:                   "channels",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	for k, v := range channels {
		if err := createChannelAsset(k, v, pid); err != nil {
			return err
		}
	}
	return nil
}

func createChannelAsset(name string, channel interface{}, parent int) error {
	props, ok := channel.(map[string]interface{})
	if !ok {
		return errors.Errorf("No properties for channel %s", name)
	}
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(channel, "#/description"),
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if ref := getString(channel, "#/$ref"); len(ref) > 0 {
		if tid := setRef(ref); tid > 0 {
			asset.AssetDataType = strconv.Itoa(tid)
		}
	}

	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if params, ok := props["parameters"]; ok {
		createParametersAsset(params, pid)
	}

	for _, op := range []string{"subscribe", "publish"} {
		if val, ok := props[op]; ok {
			createOperationAsset(op, val, pid)
		}
	}

	//TODO: process channel binding object
	return nil
}

func createParametersAsset(params interface{}, parent int) error {
	asset := Asset{
		Name:                    "parameters",
		Label:                   "parameters",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}
	if ps, ok := params.(map[string]interface{}); ok {
		for k, v := range ps {
			tid := 0
			if ref := getString(v, "#/$ref"); len(ref) > 0 {
				tid = setRef(ref)
			}
			createParameterAsset(k, v, tid, pid)
		}
	}
	return nil
}

func createParameterAsset(name string, parameter interface{}, tid int, parent int) error {
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(parameter, "#/description"),
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if tid > 0 {
		asset.AssetDataType = strconv.Itoa(tid)
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if loc := getString(parameter, "#/location"); len(loc) > 0 {
		createSimpleAsset("location", loc, pid, "string")
	}

	if schema := getRef(parameter, "#/schema"); schema != nil {
		tid := 0
		if ref := getString(schema, "#/$ref"); len(ref) > 0 {
			tid = setRef(ref)
		}
		createSchemaAsset("schema", schema, tid, pid, false)
	}
	return nil
}

func createOperationAsset(name string, operation interface{}, parent int) error {
	comment := extractExtraProperties(operation.(map[string]interface{}), []string{"description", "tags", "externalDocs", "traits", "message", "bindings"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(operation, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if tags := getRef(operation, "#/tags"); tags != nil {
		createTagsAsset(tags, pid)
	}

	if externalDocs := getRef(operation, "#/externalDocs"); externalDocs != nil {
		createExternalDocsAsset(externalDocs, pid)
	}

	if traits := getRef(operation, "#/traits"); traits != nil {
		createOperationTraitsAsset(traits, pid)
	}

	if msg := getRef(operation, "#/message"); msg != nil {
		tid := 0
		if ref := getString(msg, "#/$ref"); len(ref) > 0 {
			tid = setRef(ref)
		}
		createMessageAsset("message", msg, tid, pid)
	}
	return nil
}

func createOperationTraitsAsset(traits interface{}, parent int) error {
	ts, ok := traits.([]interface{})
	if !ok {
		return errors.Errorf("operation traits %T is not an array", traits)
	}
	asset := Asset{
		Name:                    "traits",
		Label:                   "traits",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(AssetDataTypes["array"]),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	for i, trait := range ts {
		name := fmt.Sprintf("trait-%d", i)
		if nm := getString(trait, "#/name"); len(nm) > 0 {
			name = nm
		}
		tid := 0
		if ref := getString(trait, "#/$ref"); len(ref) > 0 {
			tid = setRef(ref)
			name = ref[strings.LastIndex(ref, "/")+1:]
		}
		createOperationTraitAsset(name, trait, tid, pid)
	}
	return nil
}

func createOperationTraitAsset(name string, trait interface{}, tid int, parent int) error {
	comment := extractExtraProperties(trait.(map[string]interface{}), []string{"$ref", "externalDocs", "description", "tags", "bindings"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(trait, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if tid > 0 {
		asset.AssetDataType = strconv.Itoa(tid)
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if externalDocs := getRef(trait, "#/externalDocs"); externalDocs != nil {
		createExternalDocsAsset(externalDocs, pid)
	}
	if tags := getRef(trait, "#/tags"); tags != nil {
		createTagsAsset(tags, pid)
	}
	if bindings := getRef(trait, "#/bindings"); bindings != nil {
		if bv, err := json.MarshalIndent(bindings, "", "    "); err == nil {
			createSimpleAsset("bindings", string(bv), pid, "")
		}
	}
	return nil
}

func createMessageAsset(name string, message interface{}, tid int, parent int) error {
	comment := extractExtraProperties(message.(map[string]interface{}), []string{"$ref", "headers", "correlationId", "externalDocs", "description", "tags", "payload", "bindings", "examples", "traits"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(message, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if tid > 0 {
		asset.AssetDataType = strconv.Itoa(tid)
	}
	mid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if externalDocs := getRef(message, "#/externalDocs"); externalDocs != nil {
		createExternalDocsAsset(externalDocs, mid)
	}
	if tags := getRef(message, "#/tags"); tags != nil {
		createTagsAsset(tags, mid)
	}
	if payload := getRef(message, "#/payload"); payload != nil {
		tid := 0
		if ref := getString(payload, "#/$ref"); len(ref) > 0 {
			tid = setRef(ref)
		}
		createSchemaAsset("payload", payload, tid, mid, false)
	}

	if traits := getRef(message, "#/traits"); traits != nil {
		createMessageTraitsAsset(traits, mid)
	}

	//TODO: ignored headers, correlationId, bindings, examples
	return nil
}

func createSecuritySchemeAsset(name string, data interface{}, tid int, parent int) error {
	comment := extractExtraProperties(data.(map[string]interface{}), []string{"$ref", "description", "flows"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(data, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if tid > 0 {
		asset.AssetDataType = strconv.Itoa(tid)
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if flows := getRef(data, "#/flows"); flows != nil {
		createOAuthFlowsAsset(flows, pid)
	}
	return nil
}

func createOAuthFlowsAsset(flows interface{}, parent int) error {
	asset := Asset{
		Name:                    "flows",
		Label:                   "flows",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	fm, ok := flows.(map[string]interface{})
	if !ok {
		return errors.Errorf("flows %T is not a map", flows)
	}

	for k, v := range fm {
		comment := extractExtraProperties(v.(map[string]interface{}), []string{"scopes"})
		asset := Asset{
			Name:                    k,
			Label:                   k,
			Comment:                 comment,
			Parent:                  strconv.Itoa(pid),
			AssetType:               AssetTypes["JSON Element"],
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		fid, err := createAsset(asset)
		if err != nil {
			continue
		}
		if scopes := getRef(v, "#/scopes"); scopes != nil {
			createOAuthFlowScopesAsset(scopes, fid)
		}
	}
	return nil
}

func createOAuthFlowScopesAsset(scopes interface{}, parent int) error {
	asset := Asset{
		Name:                    "scopes",
		Label:                   "scopes",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	sm, ok := scopes.(map[string]interface{})
	if !ok {
		return errors.Errorf("flow scopes %T is not a map", scopes)
	}

	for k, v := range sm {
		createSimpleAsset(k, v.(string), pid, "string")
	}
	return nil
}

// set asset data type for a ref name, create the type if necessary.
// return type id if succesful, 0 otherwise
func setRef(ref string) int {
	tid, ok := AssetDataTypes[ref]
	var err error
	if !ok {
		if tid, err = findOrCreateAssetDataType(ref, true); err != nil {
			return 0
		}
		AssetDataTypes[ref] = tid
	}
	return tid
}

func createMessageTraitsAsset(traits interface{}, parent int) error {
	ts, ok := traits.([]interface{})
	if !ok {
		return errors.Errorf("message traits %T is not an array", traits)
	}
	asset := Asset{
		Name:                    "traits",
		Label:                   "traits",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(AssetDataTypes["array"]),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	for i, trait := range ts {
		name := fmt.Sprintf("trait-%d", i)
		if nm := getString(trait, "#/name"); len(nm) > 0 {
			name = nm
		}
		tid := 0
		if ref := getString(trait, "#/$ref"); len(ref) > 0 {
			tid = setRef(ref)
			name = ref[strings.LastIndex(ref, "/")+1:]
		}
		createMessageTraitAsset(name, trait, tid, pid)
	}
	return nil
}

func createMessageTraitAsset(name string, trait interface{}, tid int, parent int) error {
	comment := extractExtraProperties(trait.(map[string]interface{}), []string{"$ref", "headers", "correlationId", "externalDocs", "description", "tags", "bindings", "examples"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             getString(trait, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if tid > 0 {
		asset.AssetDataType = strconv.Itoa(tid)
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if externalDocs := getRef(trait, "#/externalDocs"); externalDocs != nil {
		createExternalDocsAsset(externalDocs, pid)
	}
	if tags := getRef(trait, "#/tags"); tags != nil {
		createTagsAsset(tags, pid)
	}
	if headers := getRef(trait, "#/headers"); headers != nil {
		createSchemaAsset("headers", headers, 0, pid, false)
	}

	//TODO: ignored correlationId, bindings, examples
	return nil
}

func createServersAsset(servers interface{}, parent int) error {
	asset := Asset{
		Name:                    "servers",
		Label:                   "servers",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	for k, v := range servers.(map[string]interface{}) {
		if svr, ok := v.(map[string]interface{}); ok {
			err := createServerAsset(k, svr, pid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createServerAsset(name string, server map[string]interface{}, parent int) error {
	comment := extractExtraProperties(server, []string{"description", "security", "bindings"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             server["description"].(string),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if security, ok := server["security"]; ok {
		createSecurityRequirementAsset(security, pid)
	}

	// TODO: handle server binding
	return nil
}

// server security requirement lists security schemes and scopes defined in #/components/securitySchemes
func createSecurityRequirementAsset(security interface{}, parent int) error {
	asset := Asset{
		Name:                    "security",
		Label:                   "security",
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(AssetDataTypes["array"]),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if schemes, ok := security.([]interface{}); ok {
		for _, s := range schemes {
			createSecurityRequirementScheme(s, pid)
		}
	}
	return nil
}

func createSecurityRequirementScheme(scheme interface{}, parent int) error {
	s, ok := scheme.(map[string]interface{})
	if !ok {
		return errors.Errorf("security requirement scheme %T is not a map", scheme)
	}
	for k, v := range s {
		asset := Asset{
			Name:                    k,
			Label:                   k,
			Parent:                  strconv.Itoa(parent),
			AssetType:               AssetTypes["JSON Property"],
			AssetDataType:           strconv.Itoa(AssetDataTypes["array"]),
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		pid, err := createAsset(asset)
		if err != nil {
			return err
		}
		createSecurityRequirementSchemeScopes(v, pid)
	}
	return nil
}

func createSecurityRequirementSchemeScopes(scopes interface{}, parent int) error {
	ss, ok := scopes.([]interface{})
	if !ok {
		return errors.Errorf("security requirement scheme scope %T is not an array", scopes)
	}
	for _, scope := range ss {
		name := fmt.Sprintf("%s", scope)
		asset := Asset{
			Name:                    name,
			Label:                   name,
			Parent:                  strconv.Itoa(parent),
			AssetType:               AssetTypes["JSON Property"],
			AssetDataType:           strconv.Itoa(AssetDataTypes["string"]),
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		createAsset(asset)
	}
	return nil
}

// NOT USED: replace component ref with actual definitions.
func expandComponents(spec map[string]interface{}) error {
	refMap := make(map[string]interface{})
	if err := dereference(spec, spec["components"], refMap); err != nil {
		return err
	}

	if err := dereference(spec, spec["channels"], nil); err != nil {
		return err
	}
	if err := dereference(spec, spec["servers"], nil); err != nil {
		return err
	}

	return nil
}

// replace $Ref with component definition in an element and its descendants.
// must initialize refMap if you want to perform deep deref on components
func dereference(root map[string]interface{}, elem interface{}, refMap map[string]interface{}) error {
	if elem == nil {
		return nil
	}

	switch v := elem.(type) {
	case []interface{}:
		return dereferenceArray(root, v, refMap)
	case map[string]interface{}:
		return dereferenceMap(root, v, refMap)
	default:
		return nil
	}
}

func dereferenceMap(root map[string]interface{}, elem map[string]interface{}, refMap map[string]interface{}) error {
	for k, v := range elem {
		if path := refPath(v); path != "" {
			ref, err := dereferencePath(root, path, refMap)
			if err != nil {
				return err
			}
			elem[k] = ref
		} else {
			if err := dereference(root, v, refMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// return the component ref path if elem is a $ref, or empty string otherwise
func refPath(elem interface{}) string {
	if elemap, ok := elem.(map[string]interface{}); ok {
		if path, ok := elemap["$ref"]; ok {
			return path.(string)
		}
	}
	return ""
}

func dereferenceArray(root map[string]interface{}, elem []interface{}, refMap map[string]interface{}) error {
	for i, v := range elem {
		if path := refPath(v); path != "" {
			ref, err := dereferencePath(root, path, refMap)
			if err != nil {
				return err
			}
			elem[i] = ref
		} else {
			if err := dereference(root, v, refMap); err != nil {
				return err
			}
		}
	}
	return nil
}

func dereferencePath(root map[string]interface{}, path string, refMap map[string]interface{}) (interface{}, error) {
	if refMap == nil {
		// deref top component object only
		return getRef(root, path), nil
	}

	ref, ok := refMap[path]
	if ok {
		// avoid circular reference in components
		return ref, nil
	}

	ref = getRef(root, path)
	refMap[path] = ref

	// deep deref for nested component refs
	if err := dereference(root, ref, refMap); err != nil {
		return nil, err
	}
	return ref, nil
}

func getRef(node interface{}, ref string) interface{} {
	path := strings.Split(ref, "/")[1:]
	c := node
	for _, k := range path {
		v, ok := c.(map[string]interface{})
		if !ok {
			return nil
		}
		c, ok = v[k]
		if !ok || c == nil {
			return nil
		}
	}
	return c
}

func getString(node interface{}, ref string) string {
	v := getRef(node, ref)
	if v == nil {
		return ""
	}

	return fmt.Sprintf("%v", v)
}
