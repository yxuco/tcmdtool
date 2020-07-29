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
		id, err := createAssetDataType(t, false)
		if err != nil {
			return err
		}
		AssetDataTypes[t] = id
	}
	return nil
}

func importAsyncAPISpec(spec map[string]interface{}) error {
	//if err := expandComponents(spec); err != nil {
	//	return errors.Wrap(err, "Faied to expand components")
	//}
	//data, _ := json.Marshal(spec)
	//fmt.Println(string(data))
	if err := initializeAssetDataTypes(); err != nil {
		return err
	}

	rid, err := createRootAsset()
	if err != nil {
		return nil
	}
	if id, ok := spec["id"]; ok {
		createSimpleAsset("id", id.(string), rid)
	}
	if asyncapi, ok := spec["asyncapi"]; ok {
		createSimpleAsset("asyncapi", asyncapi.(string), rid)
	}

	if info, ok := spec["info"]; ok {
		createInfoAsset(info, rid)
	}
	if externalDocs, ok := spec["externalDocs"]; ok {
		createExternalDocsAsset(externalDocs, rid)
	}
	if tags, ok := spec["tags"]; ok {
		createTagsAsset(tags, rid)
	}

	if components, ok := spec["components"]; ok {
		createComponentsAsset(components, rid)
	}

	if channels, ok := spec["channels"].(map[string]interface{}); ok {
		createChannelsAsset(channels, rid)
	}

	if servers, ok := spec["servers"]; ok {
		createServersAsset(servers, rid)
	}
	return nil
}

func createRootAsset() (int, error) {
	asset := Asset{
		Name:                    root,
		Label:                   root,
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	return createAsset(asset)
}

func createSimpleAsset(name, value string, parent int) (int, error) {
	if len(value) == 0 {
		// no value, so do not create it
		return 0, nil
	}
	asset := Asset{
		Name:                    name,
		Label:                   name,
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(AssetDataTypes["string"]),
		Comment:                 value,
		DataElementAutoAssigned: false,
		IsDisabled:              false,
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
	createSimpleAsset("version", getString(info, "#/version"), pid)
	if contact := getRef(info, "#/contact"); contact != nil {
		createContactAsset(contact, pid)
	}
	return nil
}

func createContactAsset(contact interface{}, parent int) error {
	comment := extractExtraProperties(contact.(map[string]interface{}), []string{"name"})
	asset := Asset{
		Name:                    getString(contact, "#/name"),
		Label:                   "contact",
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	if len(asset.Name) == 0 {
		asset.Name = "contact"
	}
	_, err := createAsset(asset)
	return err
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
		if len(cat) == 0 {
			continue
		}
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
		for k, v := range om {
			name := fmt.Sprintf("#/components/%s/%s", cat, k)
			createComponentAsset(v, name, cid)
		}
	}
	return nil
}

func createComponentAsset(component interface{}, name string, parent int) error {
	// create or find reusable data type
	tid := getAssetDataType(name)
	typeExists := true
	var err error
	if tid == 0 {
		if tid, err = createAssetDataType(name, true); err != nil {
			return err
		}
		typeExists = false
	}
	AssetDataTypes[name] = tid

	// create component
	comment := extractExtraProperties(component.(map[string]interface{}), []string{"description", "properties", "type", "x-examples"})
	label := name[strings.LastIndex(name, "/")+1:]
	asset := Asset{
		Name:                    label,
		Label:                   label,
		Description:             getString(component, "#/description"),
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		AssetDataType:           strconv.Itoa(tid),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if !typeExists {
		// create component properties only if this is a new reusable asset
		if props := getRef(component, "#/properties"); props != nil {
			if pm, ok := props.(map[string]interface{}); ok {
				for k, v := range pm {
					createPropertyAsset(k, v, pid)
				}
			}
		}
	}
	return err
}

func createPropertyAsset(name string, data interface{}, parent int) error {
	props, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	ptype, ok := props["type"].(string)
	if !ok {
		return nil
	}
	var err error
	// TODO: handle nested #ref in components
	switch ptype {
	case "object":
		err = createObjectPropertyAsset(name, props, parent)
	case "array":
		err = createArrayPropertyAsset(name, props, parent)
	default:
		err = createPrimitivePropertyAsset(name, props, parent)
	}
	return err
}

func createPrimitivePropertyAsset(name string, data map[string]interface{}, parent int) error {
	if tid, ok := AssetDataTypes[data["type"].(string)]; ok && tid > 0 {
		comment := extractExtraProperties(data, []string{"description", "type", "x-examples"})
		description := ""
		if desc, ok := data["description"]; ok {
			description = desc.(string)
		}
		asset := Asset{
			Name:                    name,
			Label:                   name,
			Description:             description,
			Comment:                 comment,
			Parent:                  strconv.Itoa(parent),
			AssetType:               AssetTypes["JSON Property"],
			AssetDataType:           strconv.Itoa(tid),
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		_, err := createAsset(asset)
		return err
	}
	return nil
}

func createObjectPropertyAsset(name string, data map[string]interface{}, parent int) error {
	comment := extractExtraProperties(data, []string{"description", "properties", "type", "x-examples"})
	description := ""
	if desc, ok := data["description"]; ok {
		description = desc.(string)
	}
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Description:             description,
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Property"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	// recursively create properties
	if props, ok := data["properties"]; ok {
		if pm, ok := props.(map[string]interface{}); ok {
			for k, v := range pm {
				createPropertyAsset(k, v, pid)
			}
		}
	}
	return nil
}

func createArrayPropertyAsset(name string, data map[string]interface{}, parent int) error {
	tid := AssetDataTypes["array"]
	// TODO: assume primitive array, need to process items for complex array
	comment := extractExtraProperties(data, []string{"type"})
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Comment:                 comment,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Property"],
		AssetDataType:           strconv.Itoa(tid),
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	_, err := createAsset(asset)
	return err
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
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	ops, ok := channel.(map[string]interface{})
	if !ok {
		return errors.Errorf("No operation for channel %s", name)
	}
	for k, v := range ops {
		if err := createOperationAsset(k, v, pid); err != nil {
			return err
		}
	}
	return nil
}

func createOperationAsset(name string, operation interface{}, parent int) error {
	asset := Asset{
		Name:                    name,
		Label:                   name,
		Parent:                  strconv.Itoa(parent),
		AssetType:               AssetTypes["JSON Element"],
		DataElementAutoAssigned: false,
		IsDisabled:              false,
	}
	pid, err := createAsset(asset)
	if err != nil {
		return err
	}

	if msg := getRef(operation, "#/message"); msg != nil {
		comment := extractExtraProperties(msg.(map[string]interface{}), []string{"externalDocs", "description", "tags", "payload"})
		description := ""
		if desc := getString(msg, "#/description"); len(desc) > 0 {
			description = desc
		}
		asset := Asset{
			Name:                    "message",
			Label:                   "message",
			Description:             description,
			Comment:                 comment,
			Parent:                  strconv.Itoa(pid),
			AssetType:               AssetTypes["JSON Element"],
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		mid, err := createAsset(asset)
		if err != nil {
			return err
		}

		if externalDocs := getRef(msg, "#/externalDocs"); externalDocs != nil {
			createExternalDocsAsset(externalDocs, mid)
		}
		if tags := getRef(msg, "#/tags"); tags != nil {
			createTagsAsset(tags, mid)
		}
		if payload := getRef(msg, "#/payload"); payload != nil {
			createPayloadAsset(payload, mid)
		}
	}
	return nil
}

// TODO: assumes payload is predefined in components, which may not be true
func createPayloadAsset(payload interface{}, parent int) error {
	if ref := getString(payload, "#/$ref"); len(ref) > 0 {
		tid, ok := AssetDataTypes[ref]
		if !ok {
			return errors.Errorf("payload ref %s is not defined", ref)
		}
		asset := Asset{
			Name:                    "payload",
			Label:                   "payload",
			Parent:                  strconv.Itoa(parent),
			AssetType:               AssetTypes["JSON Element"],
			AssetDataType:           strconv.Itoa(tid),
			DataElementAutoAssigned: false,
			IsDisabled:              false,
		}
		_, err := createAsset(asset)
		return err
	}
	return errors.New("payload is not $ref, which is not implemented")
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
			comment := extractExtraProperties(svr, []string{"description"})
			asset := Asset{
				Name:                    k,
				Label:                   k,
				Description:             svr["description"].(string),
				Comment:                 comment,
				Parent:                  strconv.Itoa(pid),
				AssetType:               AssetTypes["JSON Element"],
				DataElementAutoAssigned: false,
				IsDisabled:              false,
			}
			if _, err := createAsset(asset); err != nil {
				return err
			}
		}
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
