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

// Package kind provides a Terraform provider for managing KIND (Kubernetes IN Docker) clusters.
package kind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Compile-time check to ensure KindProvider satisfies the provider.Provider interface.
var (
	_ provider.Provider = &KindProvider{}
)

// KindProvider is the provider implementation using Plugin Framework.
type KindProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance tests
	version string
}

// Configure prepares the provider for data sources and resources.
//
//nolint:gocritic // it's an internal stub
func (*KindProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
	// Provider has no configuration, so nothing to do here
	// If we needed to configure clients, we would do it here and store in resp.ResourceData
}

// DataSources defines the data sources implemented in the provider.
func (*KindProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	// No data sources yet
	return make([]func() datasource.DataSource, 0)
}

// Metadata returns the provider type name.
func (p *KindProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kind"
	resp.Version = p.version
}

// Resources defines the resources implemented in the provider.
func (*KindProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
	}
}

// Schema defines the provider-level schema for configuration data.
func (*KindProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Kind provider is used to manage Kind (Kubernetes IN Docker) clusters.",
	}
}

// New returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KindProvider{
			version: version,
		}
	}
}
