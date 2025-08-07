# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    omni = {
      version = "0.1.0"
      source  = "flpajany/omni"
    }
  }
}

#
# Set TF_VAR_service_account
#
variable "service_account" {
  type = string
}

#
# Set TF_VAR_omni_uri
#
variable "omni_uri" {
  type = string
}

provider "omni" {
  uri             = var.omni_uri
  service_account = var.service_account
}

#
# Change uuids of machines to make it works
#
resource "omni_cluster" "cluster" {
  template = <<EOT
kind: Cluster
name: omni-cluster-1
kubernetes:
  version: "v1.29.9"
talos:
  version: "v1.7.7"
patches:
  - name: with-cilium-without-proxy
    inline:
      cluster:
        network:
          dnsDomain: omni-cluster-1.mydomain
          cni:
            name: custom
            urls:
              - http://myserver/cilium-withoutproxy.yaml
        proxy:
          disabled: true
        discovery:
          enabled: true
          registries:
            service:
              disabled: true
      machine:
        features:
          kubePrism:
            enabled: true
            port: 7445
        sysctls:
          fs.aio-max-nr: "1048576"
          fs.inotify.max_user_instances: "8192"
          fs.inotify.max_user_watches: "1048576"
          kernel.domainname: mydomain
          net.bridge.bridge-nf-call-arptables: "1"
          net.bridge.bridge-nf-call-ip6tables: "1"
          net.bridge.bridge-nf-call-iptables: "1"
          net.ipv4.conf.all.route_localnet: "0"
          net.ipv4.conf.all.rp_filter: "0"
          net.ipv4.ip_forward: "1"
          net.ipv4.ip_local_reserved_ports: 30000-32767
          net.ipv4.neigh.default.gc_thresh1: "8192"
          net.ipv4.neigh.default.gc_thresh2: "32768"
          net.ipv4.neigh.default.gc_thresh3: "65536"
          net.ipv4.tcp_mem: 12367902 16490539 24735804
          net.ipv4.tcp_rmem: 4096 87380 6291456
          net.ipv4.tcp_sack: "1"
          net.ipv4.tcp_timestamps: "1"
          net.ipv4.tcp_tw_reuse: "1"
          net.ipv4.tcp_wmem: 4096 16384 4194304
          net.ipv4.vs.conn_reuse_mode: "0"
          vm.max_map_count: "262144"
        time:
          disabled: false
---
kind: Machine
name: "d7413242-47ce-2140-0eee-cefb3e72d13e"
patches:
  - name: network
    inline:
      machine:
        network:
          hostname: km01
          interfaces:
          - addresses:
            - 10.0.0.2/24
            deviceSelector:
              driver: vmxnet3
              hardwareAddr: <hardwareAddr>
            mtu: 9000
            routes:
            - gateway: 10.0.0.1
              metric: 1024
              network: 0.0.0.0/0
          nameservers:
          - 8.8.8.8
          - 8.8.4.4
---
kind: Machine
name: "954d3242-0d82-0cb3-f8de-7d3342c2051d"
patches:
  - name: network
    inline:
      machine:
        network:
          hostname: kw01
          interfaces:
          - addresses:
            - 10.0.0.3/24
            deviceSelector:
              driver: vmxnet3
              hardwareAddr: <hardwareAddr>
            mtu: 9000
            routes:
            - gateway: 10.0.0.1
              metric: 1024
              network: 0.0.0.0/0
          nameservers:
          - 8.8.8.8
          - 8.8.4.4
---
kind: ControlPlane
machines:
  - "d7413242-47ce-2140-0eee-cefb3e72d13e"
---
kind: Workers
machines:
  - "954d3242-0d82-0cb3-f8de-7d3342c2051d"
EOT
}