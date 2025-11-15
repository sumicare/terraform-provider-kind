terraform {
  required_providers {
    kind = {
      source  = "sumicare/kind"
      version = "~> 1.1.0"
    }
  }
}

provider "kind" {}

# Basic cluster with default settings
resource "kind_cluster" "default" {
  name           = "test-cluster"
  wait_for_ready = true

  kind_config {
    kind = "Cluster"
    # api_version defaults to "kind.x-k8s.io/v1alpha4" if not specified

    node {
      role = "control-plane"
    }

    node {
      role = "worker"
    }
  }
}

# Output cluster connection details
output "cluster_endpoint" {
  value       = kind_cluster.default.endpoint
  description = "Kubernetes API server endpoint"
}

output "kubeconfig_path" {
  value       = kind_cluster.default.kubeconfig_path
  description = "Path to the kubeconfig file"
}

output "cluster_name" {
  value       = kind_cluster.default.name
  description = "Name of the Kind cluster"
}
