package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/dbformysql/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/sql/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/web/v2"
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

		// Create a SQL Server
		// _, err = createSqlServer(ctx, resourceGroup, subnet)
		// if err != nil {
		// 	return err
		// }

		// Create MySQL Server
		returnArgs := createMySqlServer(MySqlServerArgs{
			ctx:           ctx,
			resourceGroup: resourceGroup,
			vnet:          vnet})
		if returnArgs.err != nil {
			return returnArgs.err
		}

		backendArgs := createBackend(WebAppArgs{
			ctx:            ctx,
			resourceGroup:  resourceGroup,
			vnet:           vnet,
			appServicePlan: nil,
		})
		if backendArgs.err != nil || backendArgs.appServicePlan == nil {
			return backendArgs.err
		}

		webAppArgs := createFrontend(WebAppArgs{
			ctx:            ctx,
			resourceGroup:  resourceGroup,
			vnet:           nil,
			appServicePlan: backendArgs.appServicePlan,
		})
		if webAppArgs.err != nil {
			return webAppArgs.err
		}

		return nil
	})

}

type WebAppArgs struct {
	ctx            *pulumi.Context
	resourceGroup  *resources.ResourceGroup
	vnet           *network.VirtualNetwork
	appServicePlan *web.AppServicePlan
}

type WebAppReturn struct {
	webapp         *web.WebApp
	appServicePlan *web.AppServicePlan
	err            error
}

func createBackend(args WebAppArgs) WebAppReturn {

	// Cretae a new delegated subnet
	subnet, err := network.NewSubnet(args.ctx, "foxnhound-sn-backend", &network.SubnetArgs{
		Delegations: network.DelegationArray{
			&network.DelegationArgs{
				Name:        pulumi.String("backendDelegation"),
				ServiceName: pulumi.String("Microsoft.Web/serverFarms"),
			},
		},
		ResourceGroupName:  args.resourceGroup.Name,
		VirtualNetworkName: args.vnet.Name,
		AddressPrefix:      pulumi.String("10.0.2.0/24"),
	})
	if err != nil {
		return WebAppReturn{err: err}
	}

	if args.appServicePlan == nil {
		// Create an App Service Plan if wen need to
		args.appServicePlan, err = web.NewAppServicePlan(args.ctx, "appServicePlan", &web.AppServicePlanArgs{
			ResourceGroupName: args.resourceGroup.Name,
			Location:          args.resourceGroup.Location,
			Sku: &web.SkuDescriptionArgs{
				Name:     pulumi.String("B1"),
				Tier:     pulumi.String("Basic"),
				Capacity: pulumi.Int(1),
			},
			Reserved: pulumi.Bool(true), // Reserved indicates Linux
		})
		if err != nil {
			return WebAppReturn{err: err}
		}
	}

	// Create a Web App running a container on Linux
	webApp, err := web.NewWebApp(args.ctx, "foxnhound-backend", &web.WebAppArgs{
		ResourceGroupName: args.resourceGroup.Name,
		Location:          args.resourceGroup.Location,
		ServerFarmId:      args.appServicePlan.ID(),
		SiteConfig: &web.SiteConfigArgs{
			AlwaysOn:       pulumi.Bool(true),
			LinuxFxVersion: pulumi.String("DOCKER|nginx:latest"),
		},
		VirtualNetworkSubnetId: subnet.ID(),
	})
	if err != nil {
		return WebAppReturn{err: err}
	}

	return WebAppReturn{webapp: webApp, appServicePlan: args.appServicePlan}
}

func createFrontend(args WebAppArgs) WebAppReturn {
	// Create a Web App running a container on Linux
	webApp, err := web.NewWebApp(args.ctx, "foxnhound-webapp", &web.WebAppArgs{
		ResourceGroupName: args.resourceGroup.Name,
		Location:          args.resourceGroup.Location,
		ServerFarmId:      args.appServicePlan.ID(),
		SiteConfig: &web.SiteConfigArgs{
			AlwaysOn:       pulumi.Bool(true),
			LinuxFxVersion: pulumi.String("DOCKER|nginx:latest"),
		},
	})
	if err != nil {
		return WebAppReturn{err: err}
	}

	return WebAppReturn{webapp: webApp}
}

type MySqlServerArgs struct {
	ctx           *pulumi.Context
	resourceGroup *resources.ResourceGroup
	vnet          *network.VirtualNetwork
}

type MySqlServerReturn struct {
	dbserver *dbformysql.Server
	err      error
}

func createMySqlServer(args MySqlServerArgs) MySqlServerReturn {

	// Cretae a new delegated subnet
	subnet, err := network.NewSubnet(args.ctx, "foxnhound-sn-db", &network.SubnetArgs{
		Delegations: network.DelegationArray{
			&network.DelegationArgs{
				ServiceName: pulumi.String("Microsoft.DBforMySQL/flexibleServers"),
				Name:        pulumi.String("mysqlDelegation"),
			},
		},
		ResourceGroupName:  args.resourceGroup.Name,
		VirtualNetworkName: args.vnet.Name,
		AddressPrefix:      pulumi.String("10.0.1.0/24"),
	})
	if err != nil {
		return MySqlServerReturn{err: err}
	}

	dnszone, err := network.NewPrivateZone(args.ctx, "foxnhound-private-dns-zone", &network.PrivateZoneArgs{
		ResourceGroupName: args.resourceGroup.Name,
		PrivateZoneName:   pulumi.String("foxnhound.mysql.database.azure.com"),
		Location:          pulumi.String("Global"),
	})
	if err != nil {
		return MySqlServerReturn{err: err}
	}

	networklink, err := network.NewVirtualNetworkLink(args.ctx, "foxnhound-db-nwlink", &network.VirtualNetworkLinkArgs{
		ResourceGroupName:   args.resourceGroup.Name,
		PrivateZoneName:     dnszone.Name,
		Location:            pulumi.String("Global"),
		RegistrationEnabled: pulumi.Bool(false),
		VirtualNetwork: &network.SubResourceArgs{
			Id: args.vnet.ID(),
		},
	})
	if err != nil {
		return MySqlServerReturn{err: err}
	}

	dbserver, err := dbformysql.NewServer(args.ctx, "foxnhound-mysql-server", &dbformysql.ServerArgs{
		AdministratorLogin:         pulumi.String("sqladmin_jH5JKsj_54KJH"),
		AdministratorLoginPassword: pulumi.String("jHGJ7JKsd(sjd)jkh%"),
		ResourceGroupName:          args.resourceGroup.Name,
		Sku: &dbformysql.SkuArgs{
			Name: pulumi.String("Standard_B1ms"),
			Tier: pulumi.String(dbformysql.SkuTierBurstable),
		},
		Version: pulumi.String(dbformysql.ServerVersion_8_0_21),
		Network: &dbformysql.NetworkArgs{
			DelegatedSubnetResourceId: subnet.ID(),
			PrivateDnsZoneResourceId:  dnszone.ID(),
		},
	},
		pulumi.DependsOn([]pulumi.Resource{subnet, dnszone, networklink}))
	if err != nil {
		return MySqlServerReturn{err: err}
	}

	_, err = dbformysql.NewDatabase(args.ctx, "foxnhound-db", &dbformysql.DatabaseArgs{
		Charset:           pulumi.String("utf8"),
		Collation:         pulumi.String("utf8_general_ci"),
		ServerName:        dbserver.Name,
		DatabaseName:      dbserver.Name,
		ResourceGroupName: args.resourceGroup.Name,
	})
	if err != nil {
		return MySqlServerReturn{err: err}
	}

	return MySqlServerReturn{dbserver: dbserver}

}

type SqlServerArgs struct {
	ctx           *pulumi.Context
	resourceGroup *resources.ResourceGroup
	vnet          *network.VirtualNetwork
}

type SqlServerReturn struct {
	dbserver *sql.Server
	err      error
}

func createSqlServer(args SqlServerArgs) SqlServerReturn {
	subnet, err := network.NewSubnet(args.ctx, "foxnhound-subnet", &network.SubnetArgs{
		ResourceGroupName:  args.resourceGroup.Name,
		VirtualNetworkName: args.vnet.Name,
		AddressPrefix:      pulumi.String("10.0.1.0/24"),
	})
	if err != nil {
		return SqlServerReturn{err: err}
	}

	dbserver, err := sql.NewServer(args.ctx, "foxnhound-db-server", &sql.ServerArgs{
		Administrators: &sql.ServerExternalAdministratorArgs{
			AzureADOnlyAuthentication: pulumi.Bool(true),
			Login:                     pulumi.String("j.rosskopp@gmx.de"),
			PrincipalType:             pulumi.String(sql.PrincipalTypeUser),
			Sid:                       pulumi.String("1685f20c-0f20-4999-aaa0-550994bcc380"),
			TenantId:                  pulumi.String("e467b6d8-cf62-4e59-9a87-758ef858aeb6"),
		},
		PublicNetworkAccess:           pulumi.String(sql.ServerNetworkAccessFlagDisabled),
		ResourceGroupName:             args.resourceGroup.Name,
		RestrictOutboundNetworkAccess: pulumi.String(sql.ServerNetworkAccessFlagDisabled),
	})
	if err != nil {
		return SqlServerReturn{err: err}
	}

	_, err = sql.NewDatabase(args.ctx, "foxnhound-db", &sql.DatabaseArgs{
		ResourceGroupName: args.resourceGroup.Name,
		ServerName:        dbserver.Name,
		Sku: &sql.SkuArgs{
			Name: pulumi.String("Basic"),
			Tier: pulumi.String("Basic"),
		},
	})
	if err != nil {
		return SqlServerReturn{err: err}
	}

	// Create a Private Endpoint for the Azure SQL Database
	privateEndpoint, err := network.NewPrivateEndpoint(args.ctx, "foxnhound-private-endpoint", &network.PrivateEndpointArgs{
		ResourceGroupName: args.resourceGroup.Name,
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
		return SqlServerReturn{err: err}
	}

	// Create a Private DNS Zone for the SQL Server
	privateDnsZone, err := network.NewPrivateZone(args.ctx, "foxnhound-private-dns-zone", &network.PrivateZoneArgs{
		ResourceGroupName: args.resourceGroup.Name,
		PrivateZoneName:   pulumi.String("foxnhound.privatelink.database.windows.net"),
		Location:          pulumi.String("Global"),
	})
	if err != nil {
		return SqlServerReturn{err: err}
	}

	// Create a DNS Zone Group for the Private Endpoint
	_, err = network.NewPrivateDnsZoneGroup(args.ctx, "foxnhound-dns-zone-group", &network.PrivateDnsZoneGroupArgs{
		ResourceGroupName:   args.resourceGroup.Name,
		PrivateEndpointName: privateEndpoint.Name,
		PrivateDnsZoneConfigs: network.PrivateDnsZoneConfigArray{
			&network.PrivateDnsZoneConfigArgs{
				Name:             pulumi.String("sqlDnsZoneConfig"),
				PrivateDnsZoneId: privateDnsZone.ID(),
			},
		},
	})
	if err != nil {
		return SqlServerReturn{err: err}
	}

	return SqlServerReturn{dbserver: dbserver}
}
