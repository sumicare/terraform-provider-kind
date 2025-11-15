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
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test data constants for reusability across tests.
var (
	testKey         = "k"
	emptyMap        = make(map[string]any)
	testStringValue = "test"
	testIntValue    = 42
	testFloatValue  = 3.14
)

// mustParseBigFloat parses a string to big.Float or panics on error.
//
//nolint:revive // Test helper function
func mustParseBigFloat(s string) *big.Float {
	f, _, err := big.ParseFloat(s, 10, 64, big.ToNearestEven)
	if err != nil {
		panic(err)
	}

	return f
}

// assertNilOrEqual checks if actual is nil when expected is nil, otherwise checks equality.
func assertNilOrEqual(actual, expected any, expectNil bool, message string) {
	if expectNil {
		Expect(actual).To(BeNil(), message)
	} else {
		Expect(actual).To(Equal(expected), message)
	}
}

// Test suite for utility functions used in Kind provider.
var _ = Describe("Utils", func() {
	DescribeTable("getString - extracts string values from maps",
		func(input map[string]any, key, expected string) {
			result := getString(input, key)
			Expect(result).To(Equal(expected), "getString should handle all input types correctly")
		},
		Entry("valid string value", map[string]any{testKey: testStringValue}, testKey, testStringValue),
		Entry("missing key returns empty", emptyMap, testKey, ""),
		Entry("nil value returns empty", map[string]any{testKey: nil}, testKey, ""),
		Entry("wrong type returns empty", map[string]any{testKey: testIntValue}, testKey, ""),
	)

	DescribeTable("getInt - extracts integer values from maps",
		func(input map[string]any, key string, expected int) {
			result := getInt(input, key)
			Expect(result).To(Equal(expected), "getInt should handle all input types correctly")
		},
		Entry("valid int value", map[string]any{testKey: testIntValue}, testKey, testIntValue),
		Entry("missing key returns zero", emptyMap, testKey, 0),
		Entry("nil value returns zero", map[string]any{testKey: nil}, testKey, 0),
		Entry("wrong type returns zero", map[string]any{testKey: testStringValue}, testKey, 0),
	)

	DescribeTable("getBool - extracts boolean values from maps",
		func(input map[string]any, key string, expected bool) {
			result := getBool(input, key)
			Expect(result).To(Equal(expected), "getBool should handle all input types correctly")
		},
		Entry("true value", map[string]any{testKey: true}, testKey, true),
		Entry("false value", map[string]any{testKey: false}, testKey, false),
		Entry("missing key returns false", emptyMap, testKey, false),
		Entry("nil value returns false", map[string]any{testKey: nil}, testKey, false),
		Entry("wrong type returns false", map[string]any{testKey: "true"}, testKey, false),
	)

	DescribeTable("getStringSlice - extracts string slices from maps",
		func(input map[string]any, key string, expected []string, expectNil bool) {
			result := getStringSlice(input, key)
			assertNilOrEqual(result, expected, expectNil, "getStringSlice should handle all input types correctly")
		},
		Entry("extracts all strings", map[string]any{testKey: []any{"a", "b", "c"}}, testKey, []string{"a", "b", "c"}, false),
		Entry("filters out non-strings", map[string]any{testKey: []any{"a", 123, "b"}}, testKey, []string{"a", "b"}, false),
		Entry("handles empty slice", map[string]any{testKey: make([]any, 0)}, testKey, []string{}, false),
		Entry("missing key returns nil", emptyMap, testKey, nil, true),
		Entry("nil value returns nil", map[string]any{testKey: nil}, testKey, nil, true),
		Entry("wrong type returns nil", map[string]any{testKey: testStringValue}, testKey, nil, true),
	)

	DescribeTable("getMapSlice - extracts map slices from maps",
		func(input map[string]any, key string, expectedLen int, expectNil bool) {
			result := getMapSlice(input, key)
			if expectNil {
				Expect(result).To(BeNil(), "getMapSlice should return nil for invalid inputs")
			} else {
				Expect(result).To(HaveLen(expectedLen), "getMapSlice should return correct slice length")
			}
		},
		Entry("extracts all maps", map[string]any{testKey: []any{map[string]any{"a": 1}, map[string]any{"b": 2}}}, testKey, 2, false),
		Entry("filters out non-maps", map[string]any{testKey: []any{map[string]any{"a": 1}, testStringValue, map[string]any{"b": 2}}}, testKey, 2, false),
		Entry("handles empty slice", map[string]any{testKey: make([]any, 0)}, testKey, 0, false),
		Entry("missing key returns nil", emptyMap, testKey, 0, true),
		Entry("wrong type returns nil", map[string]any{testKey: testStringValue}, testKey, 0, true),
	)

	DescribeTable("getStringMap - extracts string maps from maps",
		func(input map[string]any, key string, expected map[string]string, expectNil bool) {
			result := getStringMap(input, key)
			assertNilOrEqual(result, expected, expectNil, "getStringMap should handle all input types correctly")
		},
		Entry("extracts all string values", map[string]any{testKey: map[string]any{"a": "v1", "b": "v2"}}, testKey, map[string]string{"a": "v1", "b": "v2"}, false),
		Entry("filters out non-strings", map[string]any{testKey: map[string]any{"a": "v1", "b": 123}}, testKey, map[string]string{"a": "v1"}, false),
		Entry("handles empty map", map[string]any{testKey: make(map[string]any)}, testKey, map[string]string{}, false),
		Entry("missing key returns nil", emptyMap, testKey, nil, true),
		Entry("nil value returns nil", map[string]any{testKey: nil}, testKey, nil, true),
		Entry("wrong type returns nil", map[string]any{testKey: testStringValue}, testKey, nil, true),
	)

	DescribeTable("normalizeToml - processes TOML input",
		func(input any, expected string, expectErr, contains bool) {
			result, err := normalizeToml(input)
			if expectErr {
				Expect(err).To(HaveOccurred(), "normalizeToml should return error for invalid input")
				Expect(result).To(Equal(expected), "normalizeToml should return input on error")
			} else {
				Expect(err).NotTo(HaveOccurred(), "normalizeToml should not return error for valid input")
				if contains {
					Expect(result).To(ContainSubstring(expected), "normalizeToml should contain expected substring")
				} else {
					Expect(result).To(Equal(expected), "normalizeToml should return expected result")
				}
			}
		},
		Entry("handles empty string", "", "", false, false),
		Entry("handles nil input", nil, "", false, false),
		Entry("handles non-string input", 123, "", false, false),
		Entry("parses valid TOML", `title = "Test"`, "title", false, true),
		Entry("returns error for invalid TOML", `invalid [[[`, `invalid [[[`, true, false),
	)

	DescribeTable("objectToMap - converts Terraform objects to Go maps",
		func(input types.Object, expected map[string]any, expectNil bool) {
			result := objectToMap(input)
			assertNilOrEqual(result, expected, expectNil, "objectToMap should handle all object types correctly")
		},
		Entry("null object returns nil", types.ObjectNull(make(map[string]attr.Type)), nil, true),
		Entry("unknown object returns nil", types.ObjectUnknown(map[string]attr.Type{testKey: types.StringType}), nil, true),
		Entry("empty object returns empty map", types.ObjectValueMust(make(map[string]attr.Type), make(map[string]attr.Value)), map[string]any{}, false),
		Entry("object with fields extracts correctly",
			types.ObjectValueMust(
				map[string]attr.Type{"name": types.StringType, "age": types.Int64Type},
				map[string]attr.Value{"name": types.StringValue(testStringValue), "age": types.Int64Value(int64(testIntValue))},
			),
			map[string]any{"name": testStringValue, "age": testIntValue}, false),
	)

	DescribeTable("listToSlice - converts Terraform lists to Go slices",
		func(input types.List, expected []any, expectNil bool) {
			result := listToSlice(input)
			assertNilOrEqual(result, expected, expectNil, "listToSlice should handle all list types correctly")
		},
		Entry("null list returns nil", types.ListNull(types.StringType), nil, true),
		Entry("unknown list returns nil", types.ListUnknown(types.StringType), nil, true),
		Entry("empty list returns empty slice", types.ListValueMust(types.StringType, make([]attr.Value, 0)), []any{}, false),
		Entry("list with values extracts correctly", types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}), []any{"a", "b"}, false),
	)

	DescribeTable("setToSlice - converts Terraform sets to Go slices",
		func(input types.Set, expected []any, expectNil bool) {
			result := setToSlice(input)
			assertNilOrEqual(result, expected, expectNil, "setToSlice should handle all set types correctly")
		},
		Entry("null set returns nil", types.SetNull(types.StringType), nil, true),
		Entry("unknown set returns nil", types.SetUnknown(types.StringType), nil, true),
		Entry("empty set returns empty slice", types.SetValueMust(types.StringType, make([]attr.Value, 0)), []any{}, false),
		Entry("set with values extracts correctly", types.SetValueMust(types.StringType, []attr.Value{types.StringValue("x")}), []any{"x"}, false),
	)

	DescribeTable("mapToMap - converts Terraform maps to Go maps",
		func(input types.Map, expected map[string]any, expectNil bool) {
			result := mapToMap(input)
			assertNilOrEqual(result, expected, expectNil, "mapToMap should handle all map types correctly")
		},
		Entry("null map returns nil", types.MapNull(types.StringType), nil, true),
		Entry("unknown map returns nil", types.MapUnknown(types.StringType), nil, true),
		Entry("empty map returns empty map", types.MapValueMust(types.StringType, make(map[string]attr.Value)), map[string]any{}, false),
		Entry("map with values extracts correctly",
			types.MapValueMust(types.StringType, map[string]attr.Value{"k1": types.StringValue("v1"), "k2": types.StringValue("v2")}),
			map[string]any{"k1": "v1", "k2": "v2"},
			false),
	)

	DescribeTable("attrValueToAny - converts Terraform attribute values to Go types",
		func(input attr.Value, expected any, matcher OmegaMatcher) {
			result := attrValueToAny(input)
			Expect(result).To(matcher, "attrValueToAny should convert attribute values correctly")
			if expected != nil {
				Expect(result).To(Equal(expected), "attrValueToAny should return expected value")
			}
		},
		Entry("string value converts correctly", types.StringValue(testStringValue), testStringValue, Equal(testStringValue)),
		Entry("bool value converts correctly", types.BoolValue(true), true, BeTrue()),
		Entry("int64 value converts correctly", types.Int64Value(int64(testIntValue)), testIntValue, Equal(testIntValue)),
		Entry("float64 value converts correctly", types.Float64Value(testFloatValue), testFloatValue, Equal(testFloatValue)),
		Entry("number value converts correctly", types.NumberValue(mustParseBigFloat("42.5")), 42.5, BeNumerically("~", 42.5, 0.01)),
		Entry("null string returns nil", types.StringNull(), nil, BeNil()),
		Entry("unknown string returns nil", types.StringUnknown(), nil, BeNil()),
		Entry("list converts to slice",
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
			[]any{"a", "b"},
			And(BeAssignableToTypeOf(make([]any, 0)), HaveLen(2))),
		Entry("set converts to slice",
			types.SetValueMust(types.StringType, []attr.Value{types.StringValue("x")}),
			[]any{"x"},
			BeAssignableToTypeOf(make([]any, 0))),
		Entry("map converts to map[string]any",
			types.MapValueMust(types.StringType, map[string]attr.Value{testKey: types.StringValue("v")}),
			map[string]any{testKey: "v"},
			BeAssignableToTypeOf(make(map[string]any))),
		Entry("object converts to map[string]any",
			types.ObjectValueMust(map[string]attr.Type{"n": types.StringType}, map[string]attr.Value{"n": types.StringValue("t")}),
			map[string]any{"n": "t"},
			BeAssignableToTypeOf(make(map[string]any))),
	)

	Describe("parseKindConfigFromFramework", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()
		})

		It("handles null and empty lists correctly", func() {
			// Test null list
			result, err := parseKindConfigFromFramework(ctx, types.ListNull(types.ObjectType{}))
			Expect(err).NotTo(HaveOccurred(), "should handle null list without error")
			Expect(result).To(BeNil(), "should return nil for null list")

			// Test empty list
			result, err = parseKindConfigFromFramework(ctx, types.ListValueMust(types.ObjectType{}, make([]attr.Value, 0)))
			Expect(err).NotTo(HaveOccurred(), "should handle empty list without error")
			Expect(result).To(BeNil(), "should return nil for empty list")
		})

		It("parses valid kind configuration correctly", func() {
			// Create object type and value for Kind configuration
			objType := map[string]attr.Type{
				"kind":        types.StringType,
				"api_version": types.StringType,
			}
			obj := types.ObjectValueMust(objType, map[string]attr.Value{
				"kind":        types.StringValue("Cluster"),
				"api_version": types.StringValue("kind.x-k8s.io/v1alpha4"),
			})
			list := types.ListValueMust(types.ObjectType{AttrTypes: objType}, []attr.Value{obj})

			result, err := parseKindConfigFromFramework(ctx, list)
			Expect(err).NotTo(HaveOccurred(), "should parse valid configuration without error")
			Expect(result).NotTo(BeNil(), "should return non-nil result for valid configuration")
			Expect(result.Kind).To(Equal("Cluster"), "should extract Kind field correctly")
			Expect(result.APIVersion).To(Equal("kind.x-k8s.io/v1alpha4"), "should extract APIVersion field correctly")
		})
	})
})
