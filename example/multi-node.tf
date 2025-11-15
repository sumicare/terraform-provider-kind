# Multi-node cluster for high availability testing
resource "kind_cluster" "ha_cluster" {
  name           = "ha-cluster"
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    # Multiple control plane nodes for HA
    node {
      role = "control-plane"
    }

    node {
      role = "control-plane"
    }

    node {
      role = "control-plane"
    }

    # Worker nodes
    node {
      role = "worker"
    }

    node {
      role = "worker"
    }

    node {
      role = "worker"
    }

    networking {
      api_server_address = "127.0.0.1"
      api_server_port    = 6444
    }
  }
}

output "ha_cluster_endpoint" {
  value       = kind_cluster.ha_cluster.endpoint
  description = "HA cluster API endpoint"
}
