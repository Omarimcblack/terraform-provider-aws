package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNetworkManagerLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkManagerLinkCreate,
		Read:   resourceAwsNetworkManagerLinkRead,
		Update: resourceAwsNetworkManagerLinkUpdate,
		Delete: resourceAwsNetworkManagerLinkDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("arn", d.Id())

				idErr := fmt.Errorf("Expected ID in format of arn:aws:networkmanager::ACCOUNTID:link/GLOBALNETWORKID/LINKID and provided: %s", d.Id())

				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}

				identifiers := strings.TrimPrefix(resARN.Resource, "link/")
				identifierParts := strings.Split(identifiers, "/")
				if len(identifierParts) != 2 {
					return nil, idErr
				}
				d.SetId(identifierParts[1])
				d.Set("global_network_id", identifierParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"download_speed": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"upload_speed": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_provider": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsNetworkManagerLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.CreateLinkInput{
		Bandwidth:       resourceAwsNetworkManagerLinkBandwidth(d),
		Description:     aws.String(d.Get("description").(string)),
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		Provider:        aws.String(d.Get("service_provider").(string)),
		SiteId:          aws.String(d.Get("site_id").(string)),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
		Type:            aws.String(d.Get("type").(string)),
	}

	log.Printf("[DEBUG] Creating Network Manager Link: %s", input)
	output, err := conn.CreateLink(input)
	if err != nil {
		return fmt.Errorf("error creating Network Manager Link: %s", err)
	}

	d.SetId(aws.StringValue(output.Link.LinkId))
	// d.Set("global_network_id", aws.StringValue(output.Link.GlobalNetworkId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.LinkStatePending},
		Target:  []string{networkmanager.LinkStateAvailable},
		Refresh: networkmanagerLinkRefreshFunc(conn, aws.StringValue(output.Link.GlobalNetworkId), aws.StringValue(output.Link.LinkId)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Network Manager Link (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsNetworkManagerLinkRead(d, meta)
}

func resourceAwsNetworkManagerLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	link, err := networkmanagerDescribeLink(conn, d.Get("global_network_id").(string), d.Id())

	if isAWSErr(err, "InvalidLinkID.NotFound", "") {
		log.Printf("[WARN] Network Manager Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Network Manager Link: %s", err)
	}

	if link == nil {
		log.Printf("[WARN] Network Manager Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(link.State) == networkmanager.LinkStateDeleting {
		log.Printf("[WARN] Network Manager Link (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(link.State))
		d.SetId("")
		return nil
	}

	d.Set("arn", link.LinkArn)
	d.Set("bandwidth", link.Bandwidth)
	d.Set("description", link.Description)
	d.Set("service_provider", link.Provider)
	d.Set("site_id", link.SiteId)
	d.Set("type", link.Type)

	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(link.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsNetworkManagerLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Network Manager Link (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsNetworkManagerLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeleteLinkInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		LinkId:          aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Network Manager Link (%s): %s", d.Id(), input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteLink(input)

		if isAWSErr(err, "IncorrectState", "has non-deleted Link Associations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Device") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteLink(input)
	}

	if isAWSErr(err, "InvalidLinkID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Link: %s", err)
	}

	if err := waitForNetworkManagerLinkDeletion(conn, d.Get("global_network_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error waiting for Network Manager Link (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsNetworkManagerLinkBandwidth(d *schema.ResourceData) *networkmanager.Bandwidth {
	count := d.Get("bandwidth.#").(int)
	if count == 0 {
		return nil
	}

	return &networkmanager.Bandwidth{
		DownloadSpeed: aws.Int64(int64(d.Get("bandwidth.0.download_speed").(int))),
		UploadSpeed:   aws.Int64(int64(d.Get("bandwidth.0.upload_speed").(int))),
	}
}

func networkmanagerLinkRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID, linkID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		link, err := networkmanagerDescribeLink(conn, globalNetworkID, linkID)

		if isAWSErr(err, "InvalidLinkID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading Network Manager Link (%s): %s", linkID, err)
		}

		if link == nil {
			return nil, "DELETED", nil
		}

		return link, aws.StringValue(link.State), nil
	}
}

func networkmanagerDescribeLink(conn *networkmanager.NetworkManager, globalNetworkID, linkID string) (*networkmanager.Link, error) {
	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkIds:         []*string{aws.String(linkID)},
	}

	log.Printf("[DEBUG] Reading Network Manager Link (%s): %s", linkID, input)
	for {
		output, err := conn.GetLinks(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.Links) == 0 {
			return nil, nil
		}

		for _, link := range output.Links {
			if link == nil {
				continue
			}

			if aws.StringValue(link.LinkId) == linkID {
				return link, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerLinkDeletion(conn *networkmanager.NetworkManager, globalNetworkID, linkID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.LinkStateAvailable,
			networkmanager.LinkStateDeleting,
		},
		Target:         []string{""},
		Refresh:        networkmanagerLinkRefreshFunc(conn, globalNetworkID, linkID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Network Manager Link (%s) deletion", linkID)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}
