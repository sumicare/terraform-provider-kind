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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test data constants for schema validation.
var (
	// Schema block names.
	kindConfigBlockName = "kind_config"
	nodeBlockName       = "node"
	networkingBlockName = "networking"

	// Schema field names.
	kindFieldName                    = "kind"
	apiVersionFieldName              = "api_version"
	containerdConfigPatchesFieldName = "containerd_config_patches"
	runtimeConfigFieldName           = "runtime_config"
	featureGatesFieldName            = "feature_gates"
)

// assertSchemaNotNil checks that schema result is not nil.
func assertSchemaNotNil(schema any, message string) {
	Expect(schema).NotTo(BeNil(), message)
}

// assertSchemaHasKey checks that schema map contains the expected key (generic version).
func assertSchemaHasKey(schema any, key, message string) {
	Expect(schema).To(HaveKey(key), message)
}

// Test suite for Kind configuration schema validation.
var _ = Describe("Schema Kind Config", func() {
	DescribeTable("kindConfigBlocks - returns kind_config block schema",
		func(expectedKey, description string) {
			blocks := kindConfigBlocks()
			assertSchemaNotNil(blocks, "blocks should not be nil")
			assertSchemaHasKey(blocks, expectedKey, description)
		},
		Entry("has kind_config block", kindConfigBlockName, "blocks should have kind_config key"),
	)

	DescribeTable("kindConfigFieldsFramework - returns kind_config field attributes",
		func(expectedKey, description string) {
			fields := kindConfigFieldsFramework()
			assertSchemaNotNil(fields, "fields should not be nil")
			assertSchemaHasKey(fields, expectedKey, description)
		},
		Entry("has kind field", kindFieldName, "fields should have kind key"),
		Entry("has api_version field", apiVersionFieldName, "fields should have api_version key"),
		Entry("has containerd_config_patches field", containerdConfigPatchesFieldName, "fields should have containerd_config_patches key"),
		Entry("has runtime_config field", runtimeConfigFieldName, "fields should have runtime_config key"),
		Entry("has feature_gates field", featureGatesFieldName, "fields should have feature_gates key"),
	)

	DescribeTable("kindConfigNestedBlocks - returns nested block schemas",
		func(expectedKey, description string) {
			blocks := kindConfigNestedBlocks()
			assertSchemaNotNil(blocks, "blocks should not be nil")
			assertSchemaHasKey(blocks, expectedKey, description)
		},
		Entry("has node block", nodeBlockName, "blocks should have node key"),
		Entry("has networking block", networkingBlockName, "blocks should have networking key"),
	)

	DescribeTable("kindConfigNestedBlocks - validates individual block schemas",
		func(blockKey, description string) {
			blocks := kindConfigNestedBlocks()
			block := blocks[blockKey]
			assertSchemaNotNil(block, description)
		},
		Entry("node block is properly configured", nodeBlockName, "node block should not be nil"),
		Entry("networking block is properly configured", networkingBlockName, "networking block should not be nil"),
	)
})
