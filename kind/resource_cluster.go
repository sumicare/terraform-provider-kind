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
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

const (
	// defaultTimeout is the default timeout for cluster operations (create/delete).
	defaultTimeout = 5 * time.Minute
	// defaultNodeImage is the default Kubernetes node image used for KIND clusters.
	defaultNodeImage = "kindest/node:v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a"
	// maxRetries is the maximum number of retry attempts for cluster creation.
	maxRetries = 2
	// retryDelay is the delay between retry attempts.
	retryDelay = 5 * time.Second
	// kubeProxyModeNone represents the "none" kube-proxy mode.
	kubeProxyModeNone = "none"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ClusterResource{}
	_ resource.ResourceWithConfigure   = &ClusterResource{}
	_ resource.ResourceWithImportState = &ClusterResource{}

	errDeleteTimeout = fmt.Errorf("delete operation timed out after %v", defaultTimeout)
)

// NewClusterResource is a helper function to simplify the provider implementation.
//
//nolint:ireturn // false positive
func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource is the resource implementation.
// ClusterResourceModel describes the resource data model.
type (
	ClusterResource struct{}

	ClusterResourceModel struct {
		KindConfig           types.List   `tfsdk:"kind_config"`
		ID                   types.String `tfsdk:"id"`
		Name                 types.String `tfsdk:"name"`
		NodeImage            types.String `tfsdk:"node_image"`
		KubeconfigPath       types.String `tfsdk:"kubeconfig_path"`
		Kubeconfig           types.String `tfsdk:"kubeconfig"`
		ClientCertificate    types.String `tfsdk:"client_certificate"`
		ClientKey            types.String `tfsdk:"client_key"`
		ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
		Endpoint             types.String `tfsdk:"endpoint"`
		WaitForReady         types.Bool   `tfsdk:"wait_for_ready"`
		Completed            types.Bool   `tfsdk:"completed"`
	}
)

// Configure adds the provider configured client to the resource.
func (*ClusterResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	// Provider has no configuration, so nothing to configure
}

// Create creates the resource and sets the initial Terraform state.
//
//nolint:gocritic // false positive
func (clusterResource *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	nodeImage := data.NodeImage.ValueString()

	// Use default node image if not provided
	if nodeImage == "" {
		nodeImage = defaultNodeImage
	}

	waitForReady := data.WaitForReady.ValueBool()
	kubeconfigPath := data.KubeconfigPath.ValueString()

	var copts []cluster.CreateOption

	if kubeconfigPath != "" {
		copts = append(copts, cluster.CreateWithKubeconfigPath(kubeconfigPath))
	}

	// Handle kind_config if provided
	if !data.KindConfig.IsNull() && len(data.KindConfig.Elements()) > 0 {
		kindConfig, err := parseKindConfigFromFramework(ctx, data.KindConfig)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing kind_config",
				"Could not parse kind_config: "+err.Error(),
			)

			return
		}

		if kindConfig != nil {
			copts = append(copts, cluster.CreateWithV1Alpha4Config(kindConfig))
		}
	}

	// Always set node image (either user-provided or default)
	copts = append(copts, cluster.CreateWithNodeImage(nodeImage))

	if waitForReady {
		copts = append(copts, cluster.CreateWithWaitForReady(defaultTimeout))
	}

	provider := cluster.NewProvider(cluster.ProviderWithLogger(cmd.NewLogger()))

	// Retry cluster creation for transient failures
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delErr := provider.Delete(name, "")
			if delErr != nil {
				tflog.Warn(ctx, fmt.Sprintf("Failed to delete cluster during retry: %v", delErr))
			}

			time.Sleep(retryDelay)
		}

		err = provider.Create(name, copts...)
		if err == nil {
			break
		}
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Kind cluster",
			fmt.Sprintf("Could not create cluster %s after %d attempts: %s", name, maxRetries+1, err.Error()),
		)

		return
	}

	// Set node_image to the actual value used (either user-provided or default)
	data.NodeImage = types.StringValue(nodeImage)

	// Set ID
	data.ID = types.StringValue(fmt.Sprintf("%s-%s", name, nodeImage))

	// Read the cluster state
	clusterResource.readClusterState(ctx, &data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Metadata returns the resource type name.
func (*ClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Read refreshes the Terraform state with the latest data.
//
//nolint:gocritic // it's an internal stub
func (clusterResource *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterResource.readClusterState(ctx, &data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Schema defines the schema for the resource.
func (*ClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kind (Kubernetes IN Docker) cluster.",
		Blocks:      kindConfigBlocks(),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the cluster resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The kind name that is given to the created cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"node_image": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The node_image that kind will use (ex: kindest/node:v1.29.7).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_for_ready": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Defines whether or not the provider will wait for the control plane to be ready. Defaults to false.",
			},
			"kubeconfig_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Kubeconfig path set after the cluster is created or by the user to override defaults.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kubeconfig": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Kubeconfig set after the cluster is created.",
			},
			"client_certificate": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client certificate for authenticating to cluster.",
			},
			"client_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client key for authenticating to cluster.",
			},
			"cluster_ca_certificate": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client verifies the server certificate with this CA cert.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Kubernetes APIServer endpoint.",
			},
			"completed": schema.BoolAttribute{
				Computed:    true,
				Description: "Cluster successfully created.",
			},
		},
	}
}

// Update updates the resource and sets the updated Terraform state on success.
//
//nolint:gocritic // it's an internal stub
func (*ClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Kind clusters don't support updates - everything is ForceNew
	// This method should not be called, but we implement it for completeness
	resp.Diagnostics.AddError(
		"Update not supported",
		"Kind clusters do not support updates. All changes require replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
//
//nolint:gocritic // it's an internal stub
func (*ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a context with timeout for delete operation
	deleteCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	name := data.Name.ValueString()
	kubeconfigPath := data.KubeconfigPath.ValueString()
	provider := cluster.NewProvider(cluster.ProviderWithLogger(cmd.NewLogger()))

	// Run delete in a goroutine to respect context timeout
	errChan := make(chan error, 1)
	go func() {
		errChan <- provider.Delete(name, kubeconfigPath)
	}()

	var err error

	select {
	case err = <-errChan:
		// Delete completed
	case <-deleteCtx.Done():
		err = errDeleteTimeout
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Kind cluster",
			fmt.Sprintf("Could not delete cluster %s: %s", name, err.Error()),
		)

		return
	}

	// Remove kubeconfig context, user, and cluster from default kubeconfig
	contextName := "kind-" + name

	// Helper function to safely remove context from a kubeconfig
	removeContext := func(configPath, configType string) {
		config, loadErr := clientcmd.LoadFromFile(configPath)
		if loadErr != nil {
			tflog.Warn(ctx, fmt.Sprintf("Unable to load %s kubeconfig for context cleanup: %v", configType, loadErr))
			return
		}

		if _, exists := config.Contexts[contextName]; !exists {
			return
		}

		delete(config.Contexts, contextName)
		delete(config.AuthInfos, contextName)
		delete(config.Clusters, contextName)

		if config.CurrentContext == contextName {
			config.CurrentContext = ""
		}

		writeErr := clientcmd.WriteToFile(*config, configPath)
		if writeErr != nil {
			tflog.Warn(ctx, fmt.Sprintf("Unable to write %s kubeconfig to remove context: %v", configType, writeErr))
		}
	}

	// Clean up default kubeconfig
	defaultKubeconfigPath := clientcmd.RecommendedHomeFile
	removeContext(defaultKubeconfigPath, "default")

	// Clean up custom kubeconfig if specified
	if kubeconfigPath != "" {
		removeContext(kubeconfigPath, "custom")
	}
}

// ImportState imports the resource state.
func (*ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readClusterState is a helper function to read cluster state.
func (*ClusterResource) readClusterState(ctx context.Context, data *ClusterResourceModel, diags *diag.Diagnostics) {
	name := data.Name.ValueString()
	provider := cluster.NewProvider(cluster.ProviderWithLogger(cmd.NewLogger()))

	tflog.Debug(ctx, "Reading cluster state for: "+name)

	kconfig, err := provider.KubeConfig(name, false)
	if err != nil {
		diags.AddError(
			"Error reading Kind cluster",
			fmt.Sprintf("Could not read kubeconfig for cluster %s: %s", name, err.Error()),
		)

		return
	}

	data.Kubeconfig = types.StringValue(kconfig)

	// Set kubeconfig_path if not already set
	if data.KubeconfigPath.IsNull() || data.KubeconfigPath.ValueString() == "" {
		currentPath, err := os.Getwd()
		if err != nil {
			diags.AddError("Error getting current directory", err.Error())
			return
		}

		exportPath := fmt.Sprintf("%s%s%s-config", currentPath, string(os.PathSeparator), name)

		err = provider.ExportKubeConfig(name, exportPath, false)
		if err != nil {
			diags.AddError(
				"Error exporting kubeconfig",
				fmt.Sprintf("Could not export kubeconfig for cluster %s: %s", name, err.Error()),
			)

			return
		}

		data.KubeconfigPath = types.StringValue(exportPath)
	}

	// Parse kubeconfig to extract connection details
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kconfig))
	if err != nil {
		diags.AddError("Error parsing kubeconfig", err.Error())
		return
	}

	data.ClientCertificate = types.StringValue(string(config.CertData))
	data.ClientKey = types.StringValue(string(config.KeyData))
	data.ClusterCACertificate = types.StringValue(string(config.CAData))
	data.Endpoint = types.StringValue(config.Host)
	data.Completed = types.BoolValue(true)
}
