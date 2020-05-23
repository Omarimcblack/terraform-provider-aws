package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAWSNetworkManagerTransitGatewayRegistrations(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_transit_gateway_registrations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerTransitGatewayRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkManagerTransitGatewayRegistrationsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "2"),
				),
			},
		},
	})
}

const testAccDataSourceAwsNetworkManagerTransitGatewayRegistrationsConfig = `
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_ec2_transit_gateway" "test" {
 count = 2
}

resource "aws_networkmanager_transit_gateway_registration" "test" {
 count               = 2
 global_network_id   = "${aws_networkmanager_global_network.test.id}"
 transit_gateway_arn = "${element(aws_ec2_transit_gateway.test.*.arn, count.index)}"
}

data "aws_networkmanager_transit_gateway_registrations" "test" {
 global_network_id   = "${aws_networkmanager_global_network.test.id}"

 depends_on = ["aws_networkmanager_transit_gateway_registration.test"]
}
`
