resource "yandex_vpc_network" "app_network" {
  name = "app-network"
}

resource "yandex_vpc_address" "admin_static_ip" {
  name = "admin-static-ip"
  external_ipv4_address {
    zone_id = local.zones[0]
  }
}

resource "yandex_vpc_gateway" "nat_gateway" {
  name = "nat-gateway"
  shared_egress_gateway {}
}

resource "yandex_vpc_subnet" "app_subnet" {
  for_each = local.zones_subnets
  name       = "app-subnet-${replace(each.key, "ru-central1-", "")}"
  zone       = each.key
  network_id = yandex_vpc_network.app_network.id
  v4_cidr_blocks = [each.value]
  route_table_id = yandex_vpc_route_table.nat_route_table.id
}


resource "yandex_vpc_route_table" "nat_route_table" {
  name       = "nat-route-table"
  network_id = yandex_vpc_network.app_network.id
  static_route {
    destination_prefix = "0.0.0.0/0"
    gateway_id         = yandex_vpc_gateway.nat_gateway.id
  }
}


resource "yandex_vpc_security_group" "admin_sg" {
  name       = "admin-sg"
  network_id = yandex_vpc_network.app_network.id

  ingress {
    protocol       = "TCP"
    port           = 22
    v4_cidr_blocks = ["0.0.0.0/0"]
    description    = "SSH from Internet to admin"
  }

  egress {
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}


resource "yandex_vpc_security_group" "app_sg" {
  name       = "app-sg"
  network_id = yandex_vpc_network.app_network.id

  ingress {
    protocol       = "TCP"
    port           = 80
    v4_cidr_blocks = ["0.0.0.0/0"]
    description    = "HTTP from Internet (via NLB)"
  }

  ingress {
    protocol       = "ANY"
    v4_cidr_blocks = [local.vpc_cidr]
    description    = "internal allow"
  }

  egress {
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

# DB SG — postgres only from internal VPC
resource "yandex_vpc_security_group" "db_sg" {
  name       = "db-sg"
  network_id = yandex_vpc_network.app_network.id

  ingress {
    protocol       = "TCP"
    port           = 5432
    v4_cidr_blocks = [local.vpc_cidr]
    description    = "Postgres access only from internal VPC"
  }

  ingress {
    protocol       = "ANY"
    v4_cidr_blocks = [local.vpc_cidr]
    description    = "internal allow"
  }

  egress {
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "worker_sg" {
  name       = "worker-sg"
  network_id = yandex_vpc_network.app_network.id

  ingress {
    protocol       = "ANY"
    v4_cidr_blocks = [local.vpc_cidr]
    description    = "internal allow"
  }

  egress {
    protocol       = "ANY"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
