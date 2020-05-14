package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsNetworkManagerLocation(d *schema.ResourceData) *networkmanager.Location {
	count := d.Get("location.#").(int)
	if count == 0 {
		return nil
	}

	return &networkmanager.Location{
		Address:   aws.String(d.Get("location.0.address").(string)),
		Latitude:  aws.String(d.Get("location.0.latitude").(string)),
		Longitude: aws.String(d.Get("location.0.longitude").(string)),
	}
}
