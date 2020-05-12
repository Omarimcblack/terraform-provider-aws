// package aws

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/service/networkmanager"
// 	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
// 	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
// )

// func resourceAwsNetworkManagerLink() *schema.Resource {
// 	return &schema.Resource{
// 		Create: resourceAwsNetworkManagerLinkCreate,
// 		Read:   resourceAwsNetworkManagerLinkRead,
// 		Update: resourceAwsNetworkManagerLinkUpdate,
// 		Delete: resourceAwsNetworkManagerLinkDelete,
// 		Importer: &schema.ResourceImporter{
// 			State: schema.ImportStatePassthrough,
// 		},

// 		Schema: map[string]*schema.Schema{
// 			"arn": {
// 				Type:     schema.TypeString,
// 				Computed: true,
// 			},
// 			"description": {
// 				Type:     schema.TypeString,
// 				Optional: true,
// 				ForceNew: true,
// 			},
// 			"tags": tagsSchema(),
// 		},
// 	}
// }

// func resourceAwsNetworkManagerLinkCreate(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).networkmanagerconn

// 	input := &networkmanager.CreateLinkInput{
// 		Description: aws.String(d.Get("description").(string)),
// 		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
// 	}

// 	if v, ok := d.GetOk("description"); ok {
// 		input.Description = aws.String(v.(string))
// 	}

// 	log.Printf("[DEBUG] Creating Network Manager Global Network: %s", input)
// 	output, err := conn.CreateLink(input)
// 	if err != nil {
// 		return fmt.Errorf("error creating Network Manager Global Network: %s", err)
// 	}

// 	d.SetId(aws.StringValue(output.Link.LinkId))

// 	stateConf := &resource.StateChangeConf{
// 		Pending: []string{networkmanager.LinkStatePending},
// 		Target:  []string{networkmanager.CustomerGatewayAssociationStateAvailable},
// 		Refresh: networkmanagerLinkRefreshFunc(conn, aws.StringValue(output.Link.LinkId)),
// 		Timeout: 10 * time.Minute,
// 	}

// 	_, err = stateConf.WaitForState()
// 	if err != nil {
// 		return fmt.Errorf("error waiting for networkmanager Global Network (%s) availability: %s", d.Id(), err)
// 	}

// 	return resourceAwsNetworkManagerLinkRead(d, meta)
// }

// func resourceAwsNetworkManagerLinkRead(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).networkmanagerconn
// 	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

// 	link, err := networkmanagerDescribeLink(conn, d.Id())

// 	if isAWSErr(err, "InvalidLinkID.NotFound", "") {
// 		log.Printf("[WARN] networkmanager Global Network (%s) not found, removing from state", d.Id())
// 		d.SetId("")
// 		return nil
// 	}

// 	if err != nil {
// 		return fmt.Errorf("error reading networkmanager Global Network: %s", err)
// 	}

// 	if link == nil {
// 		log.Printf("[WARN] networkmanager Global Network (%s) not found, removing from state", d.Id())
// 		d.SetId("")
// 		return nil
// 	}

// 	if aws.StringValue(link.State) == networkmanager.LinkStateDeleting {
// 		log.Printf("[WARN] networkmanager Global Network (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(link.State))
// 		d.SetId("")
// 		return nil
// 	}

// 	d.Set("arn", link.LinkArn)
// 	d.Set("description", link.Description)

// 	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(link.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
// 		return fmt.Errorf("error setting tags: %s", err)
// 	}

// 	return nil
// }

// func resourceAwsNetworkManagerLinkUpdate(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).networkmanagerconn

// 	if d.HasChange("tags") {
// 		o, n := d.GetChange("tags")

// 		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Id(), o, n); err != nil {
// 			return fmt.Errorf("error updating networkmanager Global Network (%s) tags: %s", d.Id(), err)
// 		}
// 	}

// 	return nil
// }

// func resourceAwsNetworkManagerLinkDelete(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).networkmanagerconn

// 	input := &networkmanager.DeleteLinkInput{
// 		LinkId: aws.String(d.Id()),
// 	}

// 	log.Printf("[DEBUG] Deleting networkmanager Global Network (%s): %s", d.Id(), input)
// 	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
// 		_, err := conn.DeleteLink(input)

// 		if isAWSErr(err, "IncorrectState", "has non-deleted Transit Gateway Registrations") {
// 			return resource.RetryableError(err)
// 		}

// 		if isAWSErr(err, "IncorrectState", "has non-deleted Customer Gateway Associations") {
// 			return resource.RetryableError(err)
// 		}

// 		if isAWSErr(err, "IncorrectState", "has non-deleted Device") {
// 			return resource.RetryableError(err)
// 		}

// 		if isAWSErr(err, "IncorrectState", "has non-deleted Link") {
// 			return resource.RetryableError(err)
// 		}

// 		if isAWSErr(err, "IncorrectState", "has non-deleted Site") {
// 			return resource.RetryableError(err)
// 		}

// 		if err != nil {
// 			return resource.NonRetryableError(err)
// 		}

// 		return nil
// 	})

// 	if isResourceTimeoutError(err) {
// 		_, err = conn.DeleteLink(input)
// 	}

// 	if isAWSErr(err, "InvalidLinkID.NotFound", "") {
// 		return nil
// 	}

// 	if err != nil {
// 		return fmt.Errorf("error deleting networkmanager Global Network: %s", err)
// 	}

// 	if err := waitForNetworkManagerLinkDeletion(conn, d.Id()); err != nil {
// 		return fmt.Errorf("error waiting for networkmanager Global Network (%s) deletion: %s", d.Id(), err)
// 	}

// 	return nil
// }

// func networkmanagerLinkRefreshFunc(conn *networkmanager.NetworkManager, linkID string) resource.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		link, err := networkmanagerDescribeLink(conn, linkID)

// 		if isAWSErr(err, "InvalidLinkID.NotFound", "") {
// 			return nil, "DELETED", nil
// 		}

// 		if err != nil {
// 			return nil, "", fmt.Errorf("error reading NetworkManager Global Network (%s): %s", linkID, err)
// 		}

// 		if link == nil {
// 			return nil, "DELETED", nil
// 		}

// 		return link, aws.StringValue(link.State), nil
// 	}
// }

// func networkmanagerDescribeLink(conn *networkmanager.NetworkManager, linkID string) (*networkmanager.Link, error) {
// 	input := &networkmanager.DescribeLinksInput{
// 		LinkIds: []*string{aws.String(linkID)},
// 	}

// 	log.Printf("[DEBUG] Reading NetworkManager Global Network (%s): %s", linkID, input)
// 	for {
// 		output, err := conn.DescribeLinks(input)

// 		if err != nil {
// 			return nil, err
// 		}

// 		if output == nil || len(output.Links) == 0 {
// 			return nil, nil
// 		}

// 		for _, link := range output.Links {
// 			if link == nil {
// 				continue
// 			}

// 			if aws.StringValue(link.LinkId) == linkID {
// 				return link, nil
// 			}
// 		}

// 		if aws.StringValue(output.NextToken) == "" {
// 			break
// 		}

// 		input.NextToken = output.NextToken
// 	}

// 	return nil, nil
// }

// func waitForNetworkManagerLinkDeletion(conn *networkmanager.NetworkManager, linkID string) error {
// 	stateConf := &resource.StateChangeConf{
// 		Pending: []string{
// 			networkmanager.LinkStateAvailable,
// 			networkmanager.LinkStateDeleting,
// 		},
// 		Target:         []string{""},
// 		Refresh:        networkmanagerLinkRefreshFunc(conn, linkID),
// 		Timeout:        10 * time.Minute,
// 		NotFoundChecks: 1,
// 	}

// 	log.Printf("[DEBUG] Waiting for NetworkManager Transit Gateway (%s) deletion", linkID)
// 	_, err := stateConf.WaitForState()

// 	if isResourceNotFoundError(err) {
// 		return nil
// 	}

// 	return err
// }
