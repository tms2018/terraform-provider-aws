// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package medialive

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{
		{
			Factory: newResourceMultiplexProgram,
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceChannel,
			TypeName: "aws_medialive_channel",
		},
		{
			Factory:  ResourceInput,
			TypeName: "aws_medialive_input",
		},
		{
			Factory:  ResourceInputSecurityGroup,
			TypeName: "aws_medialive_input_security_group",
		},
		{
			Factory:  ResourceMultiplex,
			TypeName: "aws_medialive_multiplex",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.MediaLive
}

var ServicePackage = &servicePackage{}
