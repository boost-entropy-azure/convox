locals {
  tags = merge(var.tags, {
    System = "convox"
    Rack   = var.name
  })
}

data "aws_region" "current" {
}

module "nginx" {
  source = "../nginx"

  providers = {
    kubernetes = kubernetes
  }

  cloud_provider            = "aws"
  docker_hub_authentication = var.docker_hub_authentication
  internal_router           = var.internal_router
  namespace                 = var.namespace
  nginx_image               = var.nginx_image
  proxy_protocol            = var.proxy_protocol
  rack                      = var.name
  replicas_max              = var.high_availability ? 10 : 1
  replicas_min              = var.high_availability ? 2 : 1
  ssl_ciphers               = var.ssl_ciphers
  ssl_protocols             = var.ssl_protocols
}

resource "kubernetes_config_map" "nginx-configuration" {
  metadata {
    namespace = var.namespace
    name      = "nginx-configuration"
  }

  data = {
    "proxy-body-size"           = "0"
    "use-proxy-protocol"        = var.proxy_protocol ? "true" : "false"
    "log-format-upstream"       = file("${path.module}/log-format.txt")
    "ssl-ciphers"               = var.ssl_ciphers == "" ? null : var.ssl_ciphers
    "ssl-protocols"             = var.ssl_protocols == "" ? null : var.ssl_protocols
    "allow-snippet-annotations" = "true"
    "annotations-risk-level"    = "Critical"
  }

  depends_on = [
    null_resource.set_proxy_protocol
  ]
}

resource "kubernetes_config_map" "nginx-internal-configuration" {
  metadata {
    namespace = var.namespace
    name      = "nginx-internal-configuration"
  }

  data = {
    "proxy-body-size"           = "0"
    "ssl-ciphers"               = var.ssl_ciphers == "" ? null : var.ssl_ciphers
    "ssl-protocols"             = var.ssl_protocols == "" ? null : var.ssl_protocols
    "allow-snippet-annotations" = "true"
    "annotations-risk-level"    = "Critical"
  }
}

resource "null_resource" "set_proxy_protocol" {

  count = kubernetes_service.router.spec[0].load_balancer_class == "service.k8s.aws/nlb" ? 0 : 1

  triggers = {
    proxy_protocol = var.proxy_protocol
  }

  provisioner "local-exec" {
    command = "sh ${path.module}/proxy-protocol.sh ${var.name} ${var.proxy_protocol} ${data.aws_region.current.name}"
  }

  depends_on = [
    kubernetes_service.router
  ]
}

resource "null_resource" "set_preserve_client_ip_false" {
  count = var.internal_router ? (kubernetes_service.router-internal[0].spec[0].load_balancer_class == "service.k8s.aws/nlb" ? 0 : 1) : 0

  provisioner "local-exec" {
    command = "sh ${path.module}/preserve-client-ip.sh ${var.name} ${data.aws_region.current.name}"
  }

  depends_on = [
    kubernetes_service.router-internal
  ]
}

resource "kubernetes_service" "router_extra" {
  count = var.deploy_extra_nlb ? 1 : 0

  metadata {
    namespace = var.namespace
    name      = "router-extra"

    annotations = {
      "service.beta.kubernetes.io/aws-load-balancer-name"                                = "router-extra-${var.name}"
      "service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout"             = "${var.idle_timeout}"
      "service.beta.kubernetes.io/aws-load-balancer-type"                                = "nlb"
      "service.beta.kubernetes.io/aws-load-balancer-additional-resource-tags"            = join(",", [for key, value in local.tags : "${key}=${value}"])
      "service.beta.kubernetes.io/aws-load-balancer-scheme"                              = "internet-facing"
      "service.beta.kubernetes.io/aws-load-balancer-security-groups"                     = var.nlb_security_group
      "service.beta.kubernetes.io/aws-load-balancer-manage-backend-security-group-rules" = "true"
      "service.beta.kubernetes.io/aws-load-balancer-target-group-attributes"             = var.proxy_protocol ? "proxy_protocol_v2.enabled=true" : "proxy_protocol_v2.enabled=false"
      "convox.io/dependency"                                                             = var.lbc_helm_id
    }
  }

  spec {
    external_traffic_policy = "Cluster"
    type                    = "LoadBalancer"

    load_balancer_source_ranges = var.whitelist

    port {
      name        = "http"
      port        = 80
      protocol    = "TCP"
      target_port = 80
    }

    port {
      name        = "https"
      port        = 443
      protocol    = "TCP"
      target_port = 443
    }

    selector = module.nginx.selector

    load_balancer_class = "service.k8s.aws/nlb"
  }

  lifecycle {
    ignore_changes = [spec[0].load_balancer_class]
  }
}


resource "kubernetes_service" "router" {
  metadata {
    namespace = var.namespace
    name      = "router"

    annotations = {
      "service.beta.kubernetes.io/aws-load-balancer-name"                                = "router-${var.name}"
      "service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout"             = "${var.idle_timeout}"
      "service.beta.kubernetes.io/aws-load-balancer-type"                                = "nlb"
      "service.beta.kubernetes.io/aws-load-balancer-additional-resource-tags"            = join(",", [for key, value in local.tags : "${key}=${value}"])
      "service.beta.kubernetes.io/aws-load-balancer-scheme"                              = "internet-facing"
      "service.beta.kubernetes.io/aws-load-balancer-security-groups"                     = var.nlb_security_group
      "service.beta.kubernetes.io/aws-load-balancer-manage-backend-security-group-rules" = "true"
      "service.beta.kubernetes.io/aws-load-balancer-target-group-attributes"             = var.proxy_protocol ? "proxy_protocol_v2.enabled=true" : "proxy_protocol_v2.enabled=false"
      "convox.io/dependency"                                                             = var.lbc_helm_id
    }
  }

  spec {
    external_traffic_policy = "Cluster"
    type                    = "LoadBalancer"

    load_balancer_source_ranges = var.whitelist

    port {
      name        = "http"
      port        = 80
      protocol    = "TCP"
      target_port = 80
    }

    port {
      name        = "https"
      port        = 443
      protocol    = "TCP"
      target_port = 443
    }

    selector = module.nginx.selector

    load_balancer_class = "service.k8s.aws/nlb"
  }

  lifecycle {
    ignore_changes = [spec[0].load_balancer_class]
  }
}

data "http" "alias" {
  url = "https://alias.convox.com/alias/${length(kubernetes_service.router.status.0.load_balancer.0.ingress) > 0 ? kubernetes_service.router.status.0.load_balancer.0.ingress.0.hostname : ""}"
}

resource "kubernetes_service" "router-internal" {
  count = var.internal_router ? 1 : 0
  metadata {
    namespace = var.namespace
    name      = "router-internal"

    annotations = {
      "service.beta.kubernetes.io/aws-load-balancer-name"                                = "router-internal-${var.name}"
      "service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout"             = "${var.idle_timeout}"
      "service.beta.kubernetes.io/aws-load-balancer-type"                                = "nlb"
      "service.beta.kubernetes.io/aws-load-balancer-additional-resource-tags"            = join(",", [for key, value in local.tags : "${key}=${value}"])
      "service.beta.kubernetes.io/aws-load-balancer-internal"                            = "true"
      "service.beta.kubernetes.io/aws-load-balancer-scheme"                              = "internal"
      "service.beta.kubernetes.io/aws-load-balancer-manage-backend-security-group-rules" = "true"
      "service.beta.kubernetes.io/aws-load-balancer-target-group-attributes"             = "preserve_client_ip.enabled=false"
      "convox.io/dependency"                                                             = var.lbc_helm_id
    }
  }

  spec {
    external_traffic_policy = "Cluster"
    type                    = "LoadBalancer"

    port {
      name        = "http"
      port        = 80
      protocol    = "TCP"
      target_port = 80
    }

    port {
      name        = "https"
      port        = 443
      protocol    = "TCP"
      target_port = 443
    }

    selector = module.nginx.selector-internal

    load_balancer_class = "service.k8s.aws/nlb"
  }

  lifecycle {
    ignore_changes = [spec[0].load_balancer_class]
  }
}

data "http" "alias-internal" {
  count = var.internal_router ? 1 : 0
  url   = "https://alias.convox.com/alias/${length(kubernetes_service.router-internal[0].status.0.load_balancer.0.ingress) > 0 ? kubernetes_service.router-internal[0].status.0.load_balancer.0.ingress.0.hostname : ""}"
}
