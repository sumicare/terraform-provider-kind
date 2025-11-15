# IPv6 and dual-stack networking examples
resource "kind_cluster" "ipv6_cluster" {
  name           = "ipv6-cluster"
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"
    }

    node {
      role = "worker"
    }

    networking {
      ip_family      = "ipv6"
      pod_subnet     = "fd00:10:244::/56"
      service_subnet = "fd00:10:96::/112"
    }
  }
}

# Dual-stack cluster
resource "kind_cluster" "dual_stack_cluster" {
  name           = "dual-stack-cluster"
  wait_for_ready = true

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    node {
      role = "control-plane"
    }

    node {
      role = "worker"
    }

    networking {
      ip_family      = "dual"
      pod_subnet     = "10.244.0.0/16,fd00:10:244::/56"
      service_subnet = "10.96.0.0/12,fd00:10:96::/112"
    }
  }
}

output "ipv6_cluster_endpoint" {
  value = kind_cluster.ipv6_cluster.endpoint
}

output "dual_stack_cluster_endpoint" {
  value = kind_cluster.dual_stack_cluster.endpoint
}
