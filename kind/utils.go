/*
   Copyright 2025 Sumicare

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package kind

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	toml "github.com/pelletier/go-toml"
)

// normalizeToml normalizes a TOML string by parsing and re-serializing it.
// Returns the input string unchanged if parsing fails or input is empty.
// This function is used to ensure consistent TOML formatting across operations.
func normalizeToml(tomlString any) (string, error) {
	if tomlString == nil {
		return "", nil
	}

	tomlStr, ok := tomlString.(string)
	if !ok || tomlStr == "" {
		return "", nil
	}

	tree, err := toml.Load(tomlStr)
	if err != nil {
		return tomlStr, fmt.Errorf("failed to parse TOML: %w", err)
	}

	result, err := tree.ToTomlString()
	if err != nil {
		return tomlStr, fmt.Errorf("failed to serialize TOML: %w", err)
	}

	return result, nil
}

// getString safely extracts a string from a map, returning empty string if not found or empty.
func getString(m map[string]any, key string) string {
	val, exists := m[key]
	if exists && val != nil {
		if s, isString := val.(string); isString {
			return s
		}
	}

	return ""
}

// getInt safely extracts an int from a map, returning 0 if not found.
func getInt(m map[string]any, key string) int {
	val, exists := m[key]
	if exists && val != nil {
		if i, isInt := val.(int); isInt {
			return i
		}
	}

	return 0
}

// getBool safely extracts a bool from a map, returning false if not found.
func getBool(m map[string]any, key string) bool {
	val, exists := m[key]
	if exists && val != nil {
		if b, isBool := val.(bool); isBool {
			return b
		}
	}

	return false
}

// getStringSlice safely extracts a []string from a map.
// Returns nil if the key doesn't exist or the value is not a string slice.
func getStringSlice(m map[string]any, key string) []string {
	val, exists := m[key]
	if exists && val != nil {
		if slice, isSlice := val.([]any); isSlice {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, isString := item.(string); isString {
					result = append(result, s)
				}
			}

			return result
		}
	}

	return nil
}

// getMapSlice safely extracts a []map[string]any from a map.
// Returns nil if the key doesn't exist or the value is not a slice of maps.
func getMapSlice(m map[string]any, key string) []map[string]any {
	val, exists := m[key]
	if exists && val != nil {
		if slice, isSlice := val.([]any); isSlice {
			result := make([]map[string]any, 0, len(slice))
			for _, item := range slice {
				if itemMap, isMap := item.(map[string]any); isMap {
					result = append(result, itemMap)
				}
			}

			return result
		}
	}

	return nil
}

// getStringMap safely extracts a map[string]string from a map[string]any.
// Returns nil if the key doesn't exist or the value is not a string map.
func getStringMap(m map[string]any, key string) map[string]string {
	val, exists := m[key]
	if exists && val != nil {
		if srcMap, isMap := val.(map[string]any); isMap {
			result := make(map[string]string, len(srcMap))
			for k, v := range srcMap {
				if s, isString := v.(string); isString {
					result[k] = s
				}
			}

			return result
		}
	}

	return nil
}

// parseKindConfigFromFramework converts Framework types to v1alpha4.Cluster.
// The context parameter is reserved for future use with framework operations.
func parseKindConfigFromFramework(_ context.Context, kindConfigList types.List) (*v1alpha4.Cluster, error) {
	//nolint:nilnil // false positive
	if kindConfigList.IsNull() || len(kindConfigList.Elements()) == 0 {
		return nil, nil
	}

	// Get the first (and only) kind_config block
	elements := kindConfigList.Elements()
	//nolint:nilnil // false positive
	if len(elements) == 0 {
		return nil, nil
	}

	kindConfigObj, ok := elements[0].(types.Object)
	//nolint:nilnil // false positive
	if !ok {
		return nil, nil
	}

	// Convert to map[string]any for the existing flattener
	configMap := objectToMap(kindConfigObj)

	// Use existing flattener
	cluster, err := flattenKindConfig(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kind configuration: %w", err)
	}

	return cluster, nil
}

// objectToMap converts a Framework Object to map[string]any.
func objectToMap(obj types.Object) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	result := make(map[string]any)
	attrs := obj.Attributes()

	for key, value := range attrs {
		result[key] = attrValueToAny(value)
	}

	return result
}

// attrValueToAny converts any Framework attr.Value to Go native types.
// This handles all Terraform Framework types and converts them to standard Go types.
func attrValueToAny(value attr.Value) any {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	switch val := value.(type) {
	case types.String:
		return val.ValueString()
	case types.Bool:
		return val.ValueBool()
	case types.Int64:
		return int(val.ValueInt64())
	case types.Float64:
		return val.ValueFloat64()
	case types.Number:
		f, _ := val.ValueBigFloat().Float64()
		return f

	case types.List:
		return listToSlice(val)
	case types.Set:
		return setToSlice(val)
	case types.Map:
		return mapToMap(val)
	case types.Object:
		return objectToMap(val)
	default:
		return nil
	}
}

// listToSlice converts Framework List to []any.
func listToSlice(list basetypes.ListValue) []any {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	result := make([]any, 0, len(elements))

	for _, elem := range elements {
		result = append(result, attrValueToAny(elem))
	}

	return result
}

// setToSlice converts Framework Set to []any.
func setToSlice(set basetypes.SetValue) []any {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	elements := set.Elements()
	result := make([]any, 0, len(elements))

	for _, elem := range elements {
		result = append(result, attrValueToAny(elem))
	}

	return result
}

// mapToMap converts Framework Map to map[string]any.
func mapToMap(m basetypes.MapValue) map[string]any {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}

	elements := m.Elements()
	result := make(map[string]any, len(elements))

	for key, value := range elements {
		result[key] = attrValueToAny(value)
	}

	return result
}
