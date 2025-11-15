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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// kindConfigBlocks returns the blocks for the resource schema.
func kindConfigBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"kind_config": schema.ListNestedBlock{
			Description: "The kind_config that kind will use to bootstrap the cluster.",
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: kindConfigFieldsFramework(),
				Blocks:     kindConfigNestedBlocks(),
			},
		},
	}
}

// kindConfigFieldsFramework returns the schema for kind_config using Plugin Framework types
// Based on sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster.
func kindConfigFieldsFramework() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kind": schema.StringAttribute{
			Required:    true,
			Description: "Kind cluster configuration kind (should be 'Cluster').",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"api_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("kind.x-k8s.io/v1alpha4"),
			Description: "Kind cluster configuration API version. Defaults to 'kind.x-k8s.io/v1alpha4'.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"containerd_config_patches": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Containerd configuration patches in TOML format.",
		},
		"runtime_config": schema.MapAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Runtime configuration options (underscores in keys are converted to slashes).",
		},
		"feature_gates": schema.MapAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Feature gates to enable/disable.",
		},
	}
}

// kindConfigNestedBlocks returns nested blocks for kind_config.
func kindConfigNestedBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"node": schema.ListNestedBlock{
			Description: "Nodes to create in the cluster.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"role": schema.StringAttribute{
						Optional:    true,
						Description: "Node role: 'control-plane' or 'worker'.",
					},
					"image": schema.StringAttribute{
						Optional:    true,
						Description: "Node image to use (overrides cluster-level node_image).",
					},
					"labels": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Labels to apply to the node.",
					},
					"kubeadm_config_patches": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Kubeadm config patches for this node.",
					},
					"extra_mounts": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Extra mounts for the node container.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"container_path": schema.StringAttribute{
									Optional:    true,
									Description: "Path in the container.",
								},
								"host_path": schema.StringAttribute{
									Optional:    true,
									Description: "Path on the host.",
								},
								"read_only": schema.BoolAttribute{
									Optional:    true,
									Description: "Mount as read-only.",
								},
								"selinux_relabel": schema.BoolAttribute{
									Optional:    true,
									Description: "Enable SELinux relabeling.",
								},
								"propagation": schema.StringAttribute{
									Optional:    true,
									Description: "Mount propagation: 'None', 'HostToContainer', or 'Bidirectional'.",
								},
							},
						},
					},
					"extra_port_mappings": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Extra port mappings for the node container.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"container_port": schema.Int64Attribute{
									Optional:    true,
									Description: "Port in the container.",
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.RequiresReplace(),
									},
								},
								"host_port": schema.Int64Attribute{
									Optional:    true,
									Description: "Port on the host.",
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.RequiresReplace(),
									},
								},
								"listen_address": schema.StringAttribute{
									Optional:    true,
									Description: "Listen address on the host.",
								},
								"protocol": schema.StringAttribute{
									Optional:    true,
									Description: "Protocol: 'TCP', 'UDP', or 'SCTP'.",
								},
							},
						},
					},
				},
			},
		},
		"networking": schema.SingleNestedBlock{
			Description: "Networking configuration for the cluster.",
			Attributes: map[string]schema.Attribute{
				"api_server_address": schema.StringAttribute{
					Optional:    true,
					Description: "API server listen address.",
				},
				"api_server_port": schema.Int64Attribute{
					Optional:    true,
					Description: "API server port.",
				},
				"pod_subnet": schema.StringAttribute{
					Optional:    true,
					Description: "Pod subnet CIDR.",
				},
				"service_subnet": schema.StringAttribute{
					Optional:    true,
					Description: "Service subnet CIDR.",
				},
				"disable_default_cni": schema.BoolAttribute{
					Optional:    true,
					Description: "Disable the default CNI.",
				},
				"kube_proxy_mode": schema.StringAttribute{
					Optional:    true,
					Description: "Kube-proxy mode: 'iptables', 'ipvs', or 'none'.",
				},
				"ip_family": schema.StringAttribute{
					Optional:    true,
					Description: "IP family: 'ipv4', 'ipv6', or 'dual'.",
				},
				"dns_search": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "DNS search domains.",
				},
			},
		},
	}
}
