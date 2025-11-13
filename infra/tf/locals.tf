locals {
  zones = ["ru-central1-a", "ru-central1-b", "ru-central1-d" ]
  
  zones_subnets = {
    "ru-central1-a" = "10.10.10.0/24"
    "ru-central1-b" = "10.10.20.0/24"
    "ru-central1-d" = "10.10.30.0/24"
  }

  vpc_cidr = "10.10.0.0/16"

  worker_count = 8
  workers = {
    for i in range (local.worker_count):
      "worker-${i+1}" => {
        is_public = false
        zone      = local.zones[i % length(local.zones)]
        role      = "worker"
      }
  }
  bastion = {
    "bastion-1" = {
      is_public = true
      zone      = local.zones[0]
      role      = "bastion"
    }
  }

  vms = merge(local.workers, local.bastion)

  ssh_key = trimspace(file(var.ssh_key_public_path))

}