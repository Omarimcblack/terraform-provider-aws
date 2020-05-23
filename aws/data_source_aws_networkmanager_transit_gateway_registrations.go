package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsNetworkManagerTransitGatewayRegistrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNetworkManagerTransitGatewayRegistrationsRead,

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsNetworkManagerTransitGatewayRegistrationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	log.Printf("[DEBUG] Reading Network Manager Transit Gateway Registrations: %s", input)
	output, err := conn.GetTransitGatewayRegistrations(input)

	if err != nil {
		return fmt.Errorf("error reading Network Manager Transit Gateway Registrations: %s", err)
	}

	if output == nil {
		return errors.New("error reading Network Manager Transit Gateway Registrations: no results found")
	}

	transit_gateway_arns := make([]string, 0)

	for _, arn := range output.TransitGatewayRegistrations {
		transit_gateway_arns = append(transit_gateway_arns, *arn.TransitGatewayArn)
	}

	d.Set("arns", transit_gateway_arns)

	return nil
}
