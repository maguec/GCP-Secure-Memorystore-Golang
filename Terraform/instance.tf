resource "google_compute_instance" "vm" {
  project      = var.gcp_project_id
  name         = "vm-${random_id.suffix.hex}"
  machine_type = "n1-standard-2"
  zone         = "${var.gcp_region}-a"
  tags         = ["ssh-access-${random_id.suffix.hex}"]

  metadata_startup_script = templatefile(
    "startup_script.sh",
    {
      projectid : var.gcp_project_id,
      region : var.gcp_region
      memorystore : google_redis_instance.cache.name
    },
  )


  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
    }
  }

  network_interface {
    network = google_compute_network.vpc.name
    access_config {
      # This will auto generated an external IP
    }
  }

  shielded_instance_config {
    enable_secure_boot = true
  }


  # Stop updating if the boot disk changes
  lifecycle {
    ignore_changes = [boot_disk]
  }
}

