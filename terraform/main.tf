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

variable "project_id" {}
variable "region" { default = "asia-northeast1" }
variable "image_web" {}
variable "image_sse" {}

# 共通：サービスアカウント
resource "google_service_account" "run_sa" {
  account_id   = "crun-app-sa"
  display_name = "Cloud Run App SA"
}

# ============ 通常SSR用（web） ============
resource "google_cloud_run_v2_service" "web" {
  name     = "web"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"
  template {
    service_account                  = google_service_account.run_sa.email
    max_instance_request_concurrency = 8 # 同時実行: 8（SSR向けの初期値）
    timeout                          = "3600s"

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    containers {
      image = var.image_web
      ports { container_port = 8080 }

      env {
        name  = "DATABASE_URL"
        value = var.database_url
      }
      env {
        name  = "SUPABASE_URL"
        value = var.supabase_url
      }
      env {
        name  = "SUPABASE_ANON_KEY"
        value = var.supabase_anon_key
      }
    }
  }
}

# ============ SSE専用（sse） ============
resource "google_cloud_run_v2_service" "sse" {
  name     = "sse"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"
  template {
    service_account                  = google_service_account.run_sa.email
    max_instance_request_concurrency = 1 # SSEは1が安全
    timeout                          = "3600s"

    scaling {
      min_instance_count = 0 # 必要なら 1 に
      max_instance_count = 50
    }

    containers {
      image = var.image_sse
      ports { container_port = 8080 }

      env {
        name  = "DATABASE_URL"
        value = var.database_url
      }
      env {
        name  = "SUPABASE_URL"
        value = var.supabase_url
      }
      env {
        name  = "SUPABASE_ANON_KEY"
        value = var.supabase_anon_key
      }
      # SSE用: 念のため送信バッファやKeep-Alive制御をコード側で実装
    }
  }
}

# ------------- 変数 -------------
variable "database_url" {} # Neonの接続文字列(sslmode=require)
variable "supabase_url" {}
variable "supabase_anon_key" {}

output "web_url" { value = google_cloud_run_v2_service.web.uri }
output "sse_url" { value = google_cloud_run_v2_service.sse.uri }
