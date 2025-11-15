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
	"bytes"
	"text/template"
)

// clusterConfigTemplate is the Terraform configuration template for kind_cluster resource.
const clusterConfigTemplate = `resource "kind_cluster" "test" {
  name = "{{ .Name }}"
{{- if .NodeImage }}
  node_image = "{{ .NodeImage }}"
{{- end }}
{{- if .WaitForReady }}
  wait_for_ready = true
{{- end }}
{{- if .KubeconfigPath }}
  kubeconfig_path = "{{ .KubeconfigPath }}"
{{- end }}
{{- if .KindConfig }}
  kind_config {
    kind = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"
{{- if .KindConfig.Networking }}

    networking {
{{- if .KindConfig.Networking.APIServerAddress }}
      api_server_address = "{{ .KindConfig.Networking.APIServerAddress }}"
{{- end }}
{{- if .KindConfig.Networking.APIServerPort }}
      api_server_port = {{ .KindConfig.Networking.APIServerPort }}
{{- end }}
{{- if .KindConfig.Networking.KubeProxyMode }}
      kube_proxy_mode = "{{ .KindConfig.Networking.KubeProxyMode }}"
{{- end }}
    }
{{- end }}
{{- if .KindConfig.RuntimeConfig }}

    runtime_config = {
{{- range $key, $value := .KindConfig.RuntimeConfig }}
      {{ $key }} = "{{ $value }}"
{{- end }}
    }
{{- end }}
{{- range .KindConfig.Nodes }}

    node {
      role = "{{ .Role }}"
{{- if .Image }}
      image = "{{ .Image }}"
{{- end }}
{{- if .Labels }}

      labels = {
{{- range $key, $value := .Labels }}
        {{ $key }} = "{{ $value }}"
{{- end }}
      }
{{- end }}
    }
{{- end }}
{{- if .KindConfig.ContainerdConfigPatches }}
    containerd_config_patches = [
{{- range $index, $patch := .KindConfig.ContainerdConfigPatches }}{{ if $index }},{{ end }}
      <<-TOML
{{ $patch }}
      TOML
{{- end }}
    ]
{{- end }}
  }
{{- end }}
}
`

// clusterTpl is the parsed template for cluster configuration.
//
//nolint:gochecknoglobals // it's a testing helper, we're fine
var clusterTpl = template.Must(template.New("cluster").Parse(clusterConfigTemplate))

// ClusterConfig represents the configuration for a kind cluster in tests.
type ClusterConfig struct {
	KindConfig     *KindConfig
	Name           string
	NodeImage      string
	KubeconfigPath string
	WaitForReady   bool
}

// KindConfig represents the kind-specific configuration in tests.
type KindConfig struct {
	Nodes                   []Node
	Networking              *Networking
	RuntimeConfig           map[string]string
	ContainerdConfigPatches []string
}

// Node represents a node configuration in tests.
type Node struct {
	Labels map[string]string
	Role   string
	Image  string
}

// Networking represents networking configuration in tests.
type Networking struct {
	APIServerAddress string
	KubeProxyMode    string
	APIServerPort    int
}

// renderClusterConfig renders a ClusterConfig into a Terraform configuration string.
func renderClusterConfig(cfg ClusterConfig) string {
	var buf bytes.Buffer

	err := clusterTpl.Execute(&buf, cfg)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
