
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.64.0"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

locals {
  enable_services = toset([
    "cloudbuild.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "run.googleapis.com",
    "pubsub.googleapis.com",
    "firestore.googleapis.com",
    "storage.googleapis.com"
  ])
}

resource "google_project_service" "compute_service" {
  service = "compute.googleapis.com"
}

resource "google_project_service" "service" {
  for_each = local.enable_services
  project  = var.project
  service  = each.value
  timeouts {
    create = "60m"
    update = "120m"
  }
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_pubsub_topic" "test" {
  name = "test"

  labels = {
    foo = "test"
  }

  message_retention_duration = "86600s"
}

resource "google_pubsub_subscription" "test" {
  name  = "test"
  topic = google_pubsub_topic.test.name

  labels = {
    foo = "test"
  }

  # 20 minutes
  message_retention_duration = "1200s"
  retain_acked_messages      = true

  retry_policy {
    minimum_backoff = "10s"
  }

  enable_message_ordering    = false
}

resource "google_storage_bucket" "test" {
  name          = var.project
  location      = "US"
  force_destroy = true

  lifecycle_rule {
    condition {
      age = 14
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_compute_instance_template" "test" {
  name        = "cutting"
  description = "cutting"

  tags = ["foo", "bar"]

  labels = {
    environment = "cutting"
  }

  instance_description = "description assigned to instances"
  machine_type         = "e2-standard-2"
  can_ip_forward       = false

  // Create a new boot disk from an image
  disk {
    source_image      = "debian-cloud/debian-11"
    auto_delete       = true
    boot              = true
    // backup the disk every day
    // resource_policies = [google_compute_resource_policy.daily_backup.id]
  }

  // Use an existing disk resource
  // disk {
  //   // Instance Templates reference disks by name, not self link
  //   source      = google_compute_disk.foobar.name
  //   auto_delete = false
  //   boot        = false
  // }

  network_interface {
    network = "default"
  }

  metadata = {
    foo = "bar"
  }

  service_account {
    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    email  = google_service_account.run_sa.email
    scopes = ["cloud-platform"]
  }
}

resource "google_compute_region_instance_group_manager" "test" {
  name = "cutting-machines"

  base_instance_name         = "cutting"
  region                     = "us-central1"
  distribution_policy_zones  = ["us-central1-a", "us-central1-f"]

  version {
    instance_template = google_compute_instance_template.test.self_link
  }

  // all_instances_config {
  //   metadata = {
  //     metadata_key = "metadata_value"
  //   }
  //   labels = {
  //     label_key = "label_value"
  //   }
  // }

  //target_pools = [google_compute_target_pool.test.id]
  target_size  = 2

  // auto_healing_policies {
  //   health_check      = google_compute_health_check.autohealing.id
  //   initial_delay_sec = 300
  // }
}

data "google_compute_image" "my_image" {
  family  = "debian-11"
  project = "debian-cloud"
}

resource "google_cloud_run_service" "game_api" {
  name     = "game-api"
  provider = google-beta
  location = var.region

  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
        resources {
          limits = {
            cpu    = "1000m"
            memory = "1028M"
          }
        }
      }
      service_account_name = google_service_account.run_sa.email
    }
  }
  autogenerate_revision_name = true
  depends_on                 = [google_project_service.service]
}

resource "google_cloud_run_service_iam_binding" "run_iam_binding" {
  location = google_cloud_run_service.game_api.location
  project  = google_cloud_run_service.game_api.project
  service  = google_cloud_run_service.game_api.name
  role     = "roles/run.invoker"
  members = [
    "allUsers",
  ]
}

resource "google_service_account" "run_sa" {
  account_id = "game-api"
}

resource "google_project_iam_member" "binding_run_sa" {
  role    = "roles/spanner.databaseUser"
  member  = "serviceAccount:${google_service_account.run_sa.email}"
  project = var.project
}

resource "google_compute_region_network_endpoint_group" "run_neg" {
  name                  = "run-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region
  cloud_run {
    service = google_cloud_run_service.game_api.name
  }
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_global_address" "reserved_ip" {
  name = "reserverd-ip"
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_managed_ssl_certificate" "managed_cert" {
  provider = google-beta

  name = "managed-cert"
  managed {
    domains = ["${var.domain}"]
  }
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_backend_service" "run_backend" {
  name = "run-backend"

  protocol    = "HTTP"
  port_name   = "http"
  timeout_sec = 30

  backend {
    group = google_compute_region_network_endpoint_group.run_neg.id
  }
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_url_map" "run_url_map" {
  name = "run-url-map"

  default_service = google_compute_backend_service.run_backend.id
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_target_https_proxy" "run_https_proxy" {
  name = "run-https-proxy"

  url_map = google_compute_url_map.run_url_map.id
  ssl_certificates = [
    google_compute_managed_ssl_certificate.managed_cert.id
  ]
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_compute_global_forwarding_rule" "run_lb" {
  name = "rub-lb"

  target     = google_compute_target_https_proxy.run_https_proxy.id
  port_range = "443"
  ip_address = google_compute_global_address.reserved_ip.address
  depends_on = [
    google_project_service.compute_service
  ]
}

resource "google_bigquery_dataset" "my_dataset" {
  dataset_id                  = "my_dataset"
  friendly_name               = "my_dataset"
  location                    = "US"
}

resource "google_logging_project_sink" "logging_to_bq" {
  name = "logging-to-bq"

  destination = "bigquery.googleapis.com/projects/${var.project}/datasets/${google_bigquery_dataset.my_dataset.dataset_id}"

  filter = "resource.type=\"cloud_run_revision\" AND resource.labels.configuration_name=\"game-api\" AND jsonPayload.message!=\"\""

  unique_writer_identity = true
}

resource "google_project_iam_binding" "log_writer" {
  project = var.project
  role    = "roles/bigquery.dataEditor"

  members = [
    google_logging_project_sink.logging_to_bq.writer_identity,
  ]
}

output "external_ip_attached_to_gclb" {
  value = google_compute_global_address.reserved_ip.address
}

output "cloud_run_embeded_url" {
  value = google_cloud_run_service.game_api.status[0].url
}

