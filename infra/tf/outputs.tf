output "admin_public_ip" {
  description = "Public IP address of the master node"
  value = yandex_vpc_address.admin_static_ip.external_ipv4_address[0].address
  #value       = yandex_compute_instance.vms["master"].network_interface.0.nat_ip_address
}

output "instances_internal_ips" {
  description = "Internal IP addresses of the worker nodes"
  value = {
    for vm_name, vm in yandex_compute_instance.vms :
    vm_name => vm.network_interface.0.ip_address
    if !vm.network_interface.0.nat
  }
}

output "db_credentials_secret_id" {
  description = "ID of the Lockbox secret with DB credentials"
  value       = yandex_lockbox_secret.db_credentials.id
}