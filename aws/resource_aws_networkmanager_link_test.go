package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_link", &resource.Sweeper{
		Name: "aws_networkmanager_link",
		F:    testSweepNetworkManagerLink,
	})
}

func testSweepNetworkManagerLink(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkmanagerconn
	var sweeperErrs *multierror.Error

	err = conn.GetLinksPages(&networkmanager.GetLinksInput{},
		func(page *networkmanager.GetLinksOutput, lastPage bool) bool {
			for _, link := range page.Links {
				input := &networkmanager.DeleteLinkInput{
					GlobalNetworkId: link.GlobalNetworkId,
					LinkId:          link.LinkId,
				}
				id := aws.StringValue(link.LinkId)
				globalNetworkID := aws.StringValue(link.GlobalNetworkId)

				log.Printf("[INFO] Deleting Network Manager Link: %s", id)
				_, err := conn.DeleteLink(input)

				if isAWSErr(err, "InvalidLinkID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("failed to delete Network Manager Link %s: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := waitForNetworkManagerLinkDeletion(conn, globalNetworkID, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Network Manager Link (%s) deletion: %s", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Link sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Network Manager Links: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSNetworkManagerLink_basic(t *testing.T) {
	resourceName := "aws_networkmanager_link.test"
	siteResourceName := "aws_networkmanager_site.test"
	site2ResourceName := "aws_networkmanager_site.test"
	gloablNetworkResourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkManagerLinkConfig("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerLinkExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", siteResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.0.download_speed", "10"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.0.upload_speed", "20"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSNetworkManagerLinkImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkManagerLinkConfig_Update("test updated", "company", "broadband", 30, 40),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerLinkExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "test updated"),
					resource.TestCheckResourceAttrPair(resourceName, "site_id", site2ResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "broadband"),
					resource.TestCheckResourceAttr(resourceName, "provider", "company"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.0.download_speed", "30"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth.0.upload_speed", "40"),
				),
			},
		},
	})
}

func TestAccAWSNetworkManagerLink_tags(t *testing.T) {
	resourceName := "aws_networkmanager_link.test"
	description := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkManagerLinkConfigTags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSNetworkManagerLinkImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkManagerLinkConfigTags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNetworkManagerLinkConfigTags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerLinkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsNetworkManagerLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_link" {
			continue
		}

		link, err := networkmanagerDescribeLink(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)
		if err != nil {
			if isAWSErr(err, networkmanager.ErrCodeValidationException, "") {
				return nil
			}
			return err
		}

		if link == nil {
			continue
		}

		return fmt.Errorf("Expected Link to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsNetworkManagerLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn

		link, err := networkmanagerDescribeLink(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		if link == nil {
			return fmt.Errorf("Network Manager Link not found")
		}

		if aws.StringValue(link.State) != networkmanager.LinkStateAvailable && aws.StringValue(link.State) != networkmanager.LinkStatePending {
			return fmt.Errorf("Network Manager Link (%s) exists in (%s) state", rs.Primary.ID, aws.StringValue(link.State))
		}

		return err
	}
}

func testAccNetworkManagerLinkConfig(description string) string {
	fmt.Println("MYDEBUG:	testAccNetworkManagerLinkConfig()")
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = test
 global_network_id = "${aws_networkmanager_global_network.test.id}"
}

resource "aws_networkmanager_link" "test" {
 description       = %q
 global_network_id = "${aws_networkmanager_global_network.test.id}"
 site_id           = "${aws_networkmanager_site.test.id}"

 bandwidth {
  download_speed  = 10
  upload_speed    = 20
 }
}
`, description)
}

func testAccNetworkManagerLinkConfigTags1(description, tagKey1, tagValue1 string) string {
	fmt.Println("MYDEBUG:	testAccNetworkManagerLinkConfigTags1()")
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = test
 global_network_id = "${aws_networkmanager_global_network.test.id}"
}

resource "aws_networkmanager_link" "test" {
 description       = %q
 global_network_id = "${aws_networkmanager_global_network.test.id}"
 site_id           = "${aws_networkmanager_site.test.id}"

 bandwidth {
  download_speed  = 10
  upload_speed    = 20
 }

  tags = {
    %q = %q
  }
}
`, description, tagKey1, tagValue1)
}

func testAccNetworkManagerLinkConfigTags2(description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	fmt.Println("MYDEBUG:	testAccNetworkManagerLinkConfigTags2()")
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = test
 global_network_id = "${aws_networkmanager_global_network.test.id}"
}

resource "aws_networkmanager_link" "test" {
 description       = %q
 global_network_id = "${aws_networkmanager_global_network.test.id}"
 
  bandwidth {
   download_speed  = 10
   upload_speed    = 20
  }

  tags = {
   %q = %q
   %q = %q
  }
}
`, description, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccNetworkManagerLinkConfig_Update(description, service_provider, link_type string, download_speed, upload_speed int) string {
	fmt.Println("MYDEBUG:	testAccNetworkManagerLinkConfig_Update()")
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = test
 global_network_id = "${aws_networkmanager_global_network.test.id}"
}

resource "aws_networkmanager_site" "test2" {
 description       = test2
 global_network_id = "${aws_networkmanager_global_network.test.id}"
}

resource "aws_networkmanager_link" "test" {
 description       = %q
 global_network_id = "${aws_networkmanager_global_network.test.id}"
 site_id           = "${aws_networkmanager_site.test2.id}"
 service_provider  = %q
 type              = %q

 bandwidth {
  download_speed  = %q
  upload_speed    = %q
 }
}
`, service_provider, link_type, description, download_speed, upload_speed)
}

func testAccAWSNetworkManagerLinkImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["arn"], nil
	}
}
