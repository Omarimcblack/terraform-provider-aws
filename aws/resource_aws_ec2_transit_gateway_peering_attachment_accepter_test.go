package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	PeeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", resourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", resourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", PeeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_Tags(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	PeeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", resourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", resourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", resourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", PeeringAttachmentName, "id"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", resourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", resourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", resourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", PeeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}
resource "aws_ram_resource_share" "test" {
  name = %[1]q
  tags = {
    Name = %[1]q
  }
}
resource "aws_ram_resource_association" "test" {
  resource_arn       = "${aws_ec2_transit_gateway.test.arn}"
  resource_share_arn = "${aws_ram_resource_share.test.id}"
}
resource "aws_ram_principal_association" "test" {
  principal          = "${data.aws_caller_identity.requester.account_id}"
  resource_share_arn = "${aws_ram_resource_share.test.id}"
}
# Peering attachment requester.
data "aws_caller_identity" "requester" {
  provider = "aws.alternate"
}
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "aws.alternate"
  depends_on = ["aws_ram_principal_association.test", "aws_ram_resource_association.test"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  tags = {
    Name = %[1]q
    Side = "requester"
  }
}
`, rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tags(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key3 = "Value3"
    Key2 = "Value2b"
  }
}
`, rName)
}
