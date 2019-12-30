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
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.first"
	peerTransitGatewayResourceName := "aws_ec2_transit_gateway.second"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := fmt.Sprintf("tf-testacc-tgwpeeringattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", callerIdentityDatasourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", peerTransitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
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
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.first"
	peerTransitGatewayResourceName := "aws_ec2_transit_gateway.second"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := fmt.Sprintf("tf-testacc-tgwpeeringattach-%s", acctest.RandString(8))

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
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", callerIdentityDatasourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", peerTransitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", callerIdentityDatasourceName, "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", peerTransitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
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

//func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_TransitGatewayDefaultRouteTableAssociation(t *testing.T) {
//	var providers []*schema.Provider
//	var transitGateway ec2.TransitGateway
//	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
//	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
//	transitGatewayResourceName := "aws_ec2_transit_gateway.first"
//	rName := fmt.Sprintf("tf-testacc-tgwpeeringattach-%s", acctest.RandString(8))
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			testAccPreCheck(t)
//			testAccAlternateAccountPreCheck(t)
//			testAccPreCheckAWSEc2TransitGateway(t)
//		},
//		ProviderFactories: testAccProviderFactories(&providers),
//		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_defaultRouteTableAssociation(rName, false),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
//					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
//					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTablePeeringAttachmentNotAssociated(&transitGateway, &transitGatewayPeeringAttachment),
//					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
//				),
//			},
//			{
//				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_defaultRouteTableAssociation(rName, true),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
//					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
//					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTablePeeringAttachmentAssociated(&transitGateway, &transitGatewayPeeringAttachment),
//					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
//				),
//			},
//		},
//	})
//}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"
}
resource "aws_ec2_transit_gateway" "first" {
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-ConfigTags1"
  }
}
resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-second-ConfigTags1"
  }
}
// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[2]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  provider                      = "aws.alternate"
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`, rName, testAccGetAlternateRegion())
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

//
//func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_defaultRouteTableAssociation(rName string, association bool) string {
//	return testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
//resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
//  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
//
//  tags = {
//    Name = %[1]q
//    Side = "Accepter"
//  }
//
//  transit_gateway_default_route_table_association = %[2]t
//}
//`, rName, association)
//}
