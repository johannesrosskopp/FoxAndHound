package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/sql/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an Azure Resource Group
		resourceGroup, err := resources.NewResourceGroup(ctx, "foxnhound-rg", nil)
		if err != nil {
			return err
		}

		// Create a Virtual Network
		vnet, err := network.NewVirtualNetwork(ctx, "foxnhound-vnet", &network.VirtualNetworkArgs{
			ResourceGroupName: resourceGroup.Name,
			AddressSpace: &network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String("10.0.0.0/16"),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create a Subnet
		subnet, err := network.NewSubnet(ctx, "foxnhound-subnet", &network.SubnetArgs{
			ResourceGroupName:  resourceGroup.Name,
			VirtualNetworkName: vnet.Name,
			AddressPrefix:      pulumi.String("10.0.1.0/24"),
		})
		if err != nil {
			return err
		}

		dbserver, err := sql.NewServer(ctx, "foxnhound-db-server", &sql.ServerArgs{
			Administrators: &sql.ServerExternalAdministratorArgs{
				AzureADOnlyAuthentication: pulumi.Bool(true),
				Login:                     pulumi.String("j.rosskopp@gmx.de"),
				PrincipalType:             pulumi.String(sql.PrincipalTypeUser),
				Sid:                       pulumi.String("1685f20c-0f20-4999-aaa0-550994bcc380"),
				TenantId:                  pulumi.String("e467b6d8-cf62-4e59-9a87-758ef858aeb6"),
			},
			PublicNetworkAccess:           pulumi.String(sql.ServerNetworkAccessFlagDisabled),
			ResourceGroupName:             resourceGroup.Name,
			RestrictOutboundNetworkAccess: pulumi.String(sql.ServerNetworkAccessFlagDisabled),
		})
		if err != nil {
			return err
		}

		_, err = sql.NewDatabase(ctx, "foxnhound-db", &sql.DatabaseArgs{
			ResourceGroupName: resourceGroup.Name,
			ServerName:        dbserver.Name,
			Sku: &sql.SkuArgs{
				Name:     pulumi.String("Basic"),
				Tier:     pulumi.String("Basic"),
			},
		})
		if err != nil {
			return err
		}

		// Create a Private Endpoint for the Azure SQL Database
		privateEndpoint, err := network.NewPrivateEndpoint(ctx, "foxnhound-private-endpoint", &network.PrivateEndpointArgs{
			ResourceGroupName: resourceGroup.Name,
			Subnet: &network.SubnetTypeArgs{
				Id: subnet.ID(),
			},
			PrivateLinkServiceConnections: network.PrivateLinkServiceConnectionArray{
				&network.PrivateLinkServiceConnectionArgs{
					Name:                 pulumi.String("sqlPrivateLink"),
					PrivateLinkServiceId: dbserver.ID(),
					GroupIds: pulumi.StringArray{
						pulumi.String("sqlServer"),
					},
					RequestMessage: pulumi.String("Please approve my connection"),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create a Private DNS Zone for the SQL Server
		privateDnsZone, err := network.NewPrivateZone(ctx, "foxnhound-private-dns-zone", &network.PrivateZoneArgs{
			ResourceGroupName: resourceGroup.Name,
			PrivateZoneName:  pulumi.String("foxnhound.privatelink.database.windows.net"),
			Location:          pulumi.String("Global"),
		})
		if err != nil {
			return err
		}

		// Create a DNS Zone Group for the Private Endpoint
		_, err = network.NewPrivateDnsZoneGroup(ctx, "foxnhound-dns-zone-group", &network.PrivateDnsZoneGroupArgs{
			ResourceGroupName:   resourceGroup.Name,
			PrivateEndpointName: privateEndpoint.Name,
			PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
				&network.PrivateDnsZoneConfigArgs{
					Name:             pulumi.String("sqlDnsZoneConfig"),
					PrivateDnsZoneId: privateDnsZone.ID(),
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
