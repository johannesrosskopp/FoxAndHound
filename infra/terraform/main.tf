terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.0.0"
    }
  }
}

provider "azurerm" {
  features {}

  subscription_id = var.subscription_id
  client_id       = var.client_id
  client_secret   = var.client_secret
  tenant_id       = var.tenant_id
}

resource "azurerm_resource_group" "foxnhound_rg" {
  name     = "foxnhound-rg"
  # location = "West Europe"
  location = "eastus"
}

resource "azurerm_virtual_network" "backend-network" {
  name                = "foxnhound-vnet"
  resource_group_name = azurerm_resource_group.foxnhound_rg.name
  location            = azurerm_resource_group.foxnhound_rg.location
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "database-subnet" {
  name                 = "foxnhound-db-sn"
  resource_group_name  = azurerm_resource_group.foxnhound_rg.name
  virtual_network_name = azurerm_virtual_network.backend-network.name
  address_prefixes     = ["10.0.2.0/24"]
  service_endpoints    = ["Microsoft.Storage"]
  delegation {
    name = "fs"
    service_delegation {
      name = "Microsoft.DBforMySQL/flexibleServers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
      ]
    }
  }
}

resource "azurerm_private_dns_zone" "db-dnsz" {
  name                = "foxnhound.mysql.database.azure.com"
  resource_group_name = azurerm_resource_group.foxnhound_rg.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "db-dnsprivli" {
  name                  = "foxnhoundVnetZone.com"
  private_dns_zone_name = azurerm_private_dns_zone.db-dnsz.name
  virtual_network_id    = azurerm_virtual_network.backend-network.id
  resource_group_name   = azurerm_resource_group.foxnhound_rg.name
}

resource "azurerm_mysql_flexible_server" "db-sv" {
  name                   = "foxnhound-fs"
  resource_group_name    = azurerm_resource_group.foxnhound_rg.name
  location               = azurerm_resource_group.foxnhound_rg.location
  administrator_login    = "sqladminj64_2l54aHgts"
  administrator_password = "Hjhdf566+&*3hkjsd&98798"
  backup_retention_days  = 7
  delegated_subnet_id    = azurerm_subnet.database-subnet.id
  private_dns_zone_id    = azurerm_private_dns_zone.db-dnsz.id
  sku_name               = "B_Standard_B1ms"

  depends_on = [azurerm_private_dns_zone_virtual_network_link.db-dnsprivli]
}

resource "azurerm_mysql_flexible_database" "db" {
  name                = "foxnhound-db"
  resource_group_name = azurerm_resource_group.foxnhound_rg.name
  server_name         = azurerm_mysql_flexible_server.db-sv.name
  charset             = "utf8"
  collation           = "utf8_unicode_ci"
}

# App Service

resource "azurerm_subnet" "appserv-sn" {
  name                 = "foxnhound-as-sn"
  resource_group_name  = azurerm_resource_group.foxnhound_rg.name
  virtual_network_name = azurerm_virtual_network.backend-network.name
  address_prefixes     = ["10.0.3.0/24"]

  delegation {
    name = "delegation"

    service_delegation {
        actions = [
            "Microsoft.Network/virtualNetworks/subnets/action",
            "Microsoft.Network/virtualNetworks/subnets/join/action"
          ]
        name    = "Microsoft.Web/serverFarms"
      }
  }
}

resource "azurerm_service_plan" "backend-asp" {
  name                = "foxnhound-asp"
  resource_group_name = azurerm_resource_group.foxnhound_rg.name
  location            = azurerm_resource_group.foxnhound_rg.location
  os_type             = "Linux"
  sku_name            = "B1"
  
}

resource "azurerm_linux_web_app" "backend-wa" {
  name                = "foxnhound-wa"
  resource_group_name = azurerm_resource_group.foxnhound_rg.name
  location            = azurerm_service_plan.backend-asp.location
  service_plan_id     = azurerm_service_plan.backend-asp.id

  site_config {

  }

  virtual_network_subnet_id = azurerm_subnet.appserv-sn.id
}