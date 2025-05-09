variable "cert_duration" {
  default = "2160h"
  type    = string
}

variable "docker_hub_username" {
  default = ""
}

variable "docker_hub_password" {
  default = ""
}

variable "image" {
  default = "convox/convox"
}

variable "k8s_version" {
  type    = string
  default = "1.31"
}

variable "name" {
  type = string
}

variable "rack_name" {
  default = ""
  type    = string
}

variable "node_type" {
  default = "Standard_D2_v3"
}

variable "region" {
  default = "eastus"
}

variable "release" {
  default = ""
}

variable "settings" {
  default = ""
}

variable "syslog" {
  default = ""
}

variable "telemetry" {
  type    = bool
  default = false
}

variable "whitelist" {
  default = "0.0.0.0/0"
}
