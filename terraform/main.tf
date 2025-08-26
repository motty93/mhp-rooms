terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# -------------------- Variables --------------------
variable "project_id" {}
variable "region" { default = "asia-northeast1" }

# Container images
variable "image_web" {}
variable "image_sse" {}

# .env 相当（本番は Secret Manager を推奨）
variable "DATABASE_URL" {}
variable "SUPABASE_URL" {}
variable "SUPABASE_ANON_KEY" {}
variable "SUPABASE_JWT_SECRET" {}

variable "ENV" { default = "production" }
variable "PORT" { default = "8080" }
variable "LOG_LEVEL" { default = "info" }
variable "DEBUG_SQL_LOGS" { default = "false" }

# Billing / Budget
variable "billing_account_id" {}
variable "budget_currency" { default = "JPY" }
variable "budget_amount" { default = 1500 } # 月のしきい値（例: 1,500円）

# -------------------- Service Account --------------------
resource "google_service_account" "run_sa" {
  account_id   = "crun-app-sa"
  display_name = "Cloud Run App SA"
}

# ==================== Cloud Run: web（SSR/API） ====================
resource "google_cloud_run_v2_service" "web" {
  name     = "web"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.run_sa.email

    # 同時実行（SSR初期値）
    max_instance_request_concurrency = 8
    timeout                          = "3600s"

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    containers {
      image = var.image_web

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          memory = "512Mi"
          # cpu  = "1"
        }
      }

      env {
        name  = "PORT"
        value = var.PORT
      }
      env {
        name  = "ENV"
        value = var.ENV
      }
      env {
        name  = "LOG_LEVEL"
        value = var.LOG_LEVEL
      }
      env {
        name  = "DEBUG_SQL_LOGS"
        value = var.DEBUG_SQL_LOGS
      }
      env {
        name  = "DATABASE_URL"
        value = var.DATABASE_URL
      }
      env {
        name  = "SUPABASE_URL"
        value = var.SUPABASE_URL
      }
      env {
        name  = "SUPABASE_ANON_KEY"
        value = var.SUPABASE_ANON_KEY
      }
      env {
        name  = "SUPABASE_JWT_SECRET"
        value = var.SUPABASE_JWT_SECRET
      }
    }
  }
}

# ==================== Cloud Run: sse（SSE専用） ====================
resource "google_cloud_run_v2_service" "sse" {
  name     = "sse"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.run_sa.email

    # 長時間接続の安定化
    max_instance_request_concurrency = 1
    timeout                          = "3600s"

    scaling {
      min_instance_count = 0 # 必要に応じて 1 に
      max_instance_count = 50
    }

    containers {
      image = var.image_sse

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          memory = "512Mi"
          # cpu  = "1"
        }
      }

      env {
        name  = "PORT"
        value = var.PORT
      }
      env {
        name  = "ENV"
        value = var.ENV
      }
      env {
        name  = "LOG_LEVEL"
        value = var.LOG_LEVEL
      }
      env {
        name  = "DEBUG_SQL_LOGS"
        value = var.DEBUG_SQL_LOGS
      }
      env {
        name  = "DATABASE_URL"
        value = var.DATABASE_URL
      }
      env {
        name  = "SUPABASE_URL"
        value = var.SUPABASE_URL
      }
      env {
        name  = "SUPABASE_ANON_KEY"
        value = var.SUPABASE_ANON_KEY
      }
      env {
        name  = "SUPABASE_JWT_SECRET"
        value = var.SUPABASE_JWT_SECRET
      }
    }
  }
}

# ==================== Budget (通知のみ) ====================
data "google_billing_account" "account" {
  billing_account = var.billing_account_id
}

data "google_project" "this" {}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.account.id
  display_name    = "Budget for ${data.google_project.this.project_id}"

  budget_filter {
    projects               = ["projects/${data.google_project.this.number}"]
    credit_types_treatment = "EXCLUDE_ALL_CREDITS"
  }

  amount {
    specified_amount {
      currency_code = var.budget_currency
      units         = var.budget_amount
    }
  }

  threshold_rules { threshold_percent = 0.5 }
  threshold_rules { threshold_percent = 0.9 }
  threshold_rules { threshold_percent = 1.0 }
}

# -------------------- Outputs --------------------
output "web_url" {
  value = google_cloud_run_v2_service.web.uri
}

output "sse_url" {
  value = google_cloud_run_v2_service.sse.uri
}
