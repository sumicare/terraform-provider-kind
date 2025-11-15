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
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test data constants for provider testing.
var (
	testVersion = "test"
)

// assertProviderNotNil checks that provider result is not nil.
func assertProviderNotNil(provider any, message string) {
	Expect(provider).NotTo(BeNil(), message)
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
//
//nolint:gochecknoglobals // This is a global factory used for testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kind": providerserver.NewProtocol6WithError(New("test")()),
}

var _ = Describe("Provider Framework", func() {
	DescribeTable("New - creates provider factory and instances",
		func(version string, validator func(any)) {
			factory := New(version)
			assertProviderNotNil(factory, "factory should not be nil")

			provider := factory()
			assertProviderNotNil(provider, "provider should not be nil")
			validator(provider)
		},
		Entry("creates provider with test version", testVersion, func(provider any) {
			// Additional validation can be added here if needed
			Expect(provider).To(BeAssignableToTypeOf(&KindProvider{}), "provider should be of type KindProvider")
		}),
	)

	DescribeTable("KindProvider - validates provider metadata",
		func(version string, validator func(*KindProvider)) {
			provider := &KindProvider{version: version}
			assertProviderNotNil(provider, "provider should not be nil")
			validator(provider)
		},
		Entry("has correct test version", testVersion, func(provider *KindProvider) {
			Expect(provider.version).To(Equal(testVersion), "version should match expected value")
		}),
	)
})
