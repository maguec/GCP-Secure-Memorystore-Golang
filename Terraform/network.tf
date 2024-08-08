resource "google_compute_network" "vpc" {
  project = var.gcp_project_id
  name    = "vpc-${random_id.suffix.hex}"
}

resource "google_compute_firewall" "ssh-access" {
  project = var.gcp_project_id
  name = "firewall-${random_id.suffix.hex}"
  network = google_compute_network.vpc.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  source_tags   = ["ssh-access-${random_id.suffix.hex}"]

}
