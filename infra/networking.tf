# VPC — the private network that contains all Moraine infrastructure.
# 10.0.0.0/16 gives 65,536 addresses, far more than needed but standard
# for a VPC since address space is free.
resource "aws_vpc" "moraine" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name    = "${var.project}-vpc"
    Project = var.project
  }
}

# Two availability zones — minimum required by both EKS and MSK.
# If one AZ goes down, the cluster can still operate in the other.
data "aws_availability_zones" "available" {
  state = "available"
}

# Public subnets — for load balancers that accept external traffic.
# EKS nodes and MSK brokers do NOT go here.
resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.moraine.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true

  tags = {
    Name                     = "${var.project}-public-${count.index + 1}"
    Project                  = var.project
    "kubernetes.io/role/elb" = "1"
  }
}

# Private subnets — EKS nodes and MSK brokers live here.
# Not directly reachable from the internet.
resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.moraine.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name                              = "${var.project}-private-${count.index + 1}"
    Project                           = var.project
    "kubernetes.io/role/internal-elb" = "1"
  }
}

# Internet Gateway — allows the public subnets to reach the internet.
resource "aws_internet_gateway" "moraine" {
  vpc_id = aws_vpc.moraine.id

  tags = {
    Name    = "${var.project}-igw"
    Project = var.project
  }
}

# Elastic IP for the NAT Gateway.
resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name    = "${var.project}-nat-eip"
    Project = var.project
  }
}

# NAT Gateway — allows private subnet resources (EKS nodes, MSK brokers)
# to reach the internet for pulling images and updates, without being
# directly reachable from the internet themselves.
resource "aws_nat_gateway" "moraine" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name    = "${var.project}-nat"
    Project = var.project
  }
}

# Route table for public subnets — sends all traffic to the IGW.
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.moraine.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.moraine.id
  }

  tags = {
    Name    = "${var.project}-public-rt"
    Project = var.project
  }
}

# Route table for private subnets — sends internet traffic through the NAT.
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.moraine.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.moraine.id
  }

  tags = {
    Name    = "${var.project}-private-rt"
    Project = var.project
  }
}

# Associate route tables with their subnets.
resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
