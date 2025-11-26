data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2004-lts"
}

resource "yandex_compute_instance" "vms" {
  for_each = local.vms

  name = each.key
  zone = each.value.zone
  platform_id = "standard-v3" 
  resources {
    cores         = 2
    memory        = 2
    core_fraction = 20
  }

  scheduling_policy {
    preemptible = true
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
      size     = 20
      type     = "network-hdd"
    }
  }

  network_interface {
    subnet_id = each.key == "bastion-1" ? yandex_vpc_subnet.bastion_subnet.id : yandex_vpc_subnet.app_subnet[each.value.zone].id
    nat                = each.value.is_public
    nat_ip_address = each.key == "bastion-1" ? yandex_vpc_address.admin_static_ip.external_ipv4_address[0].address : null
    ip_address = each.key == "bastion-1" ? "10.10.99.10" : null
    
    security_group_ids = [
        each.value.is_public ? yandex_vpc_security_group.admin_sg.id : yandex_vpc_security_group.worker_sg.id
    ]
  }


  metadata = {
    user-data = templatefile("D:/MEPHI/7_semestr/SRP/GORGON/infra/tf/cloud-init.yml", {
      ssh_public_key = chomp(local.ssh_key)
    })
    #ssh-keys = "root:${local.ssh_key}"
  }
}