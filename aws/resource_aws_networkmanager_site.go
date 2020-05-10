package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNetworkManagerSite() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkManagerSiteCreate,
		Read:   resourceAwsNetworkManagerSiteRead,
		Update: resourceAwsNetworkManagerSiteUpdate,
		Delete: resourceAwsNetworkManagerSiteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNetworkManagerSiteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.CreateSiteInput{
		Description:     aws.String(d.Get("description").(string)),
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		Location:        resourceAwsAwsNetworkManagerSiteLocation(d),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
	}

	log.Printf("[DEBUG] Creating Network Manager Site: %s", input)
	output, err := conn.CreateSite(input)
	if err != nil {
		return fmt.Errorf("error creating Network Manager Site: %s", err)
	}

	d.SetId(aws.StringValue(output.Site.SiteId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.SiteStatePending},
		Target:  []string{networkmanager.CustomerGatewayAssociationStateAvailable},
		Refresh: networkmanagerSiteRefreshFunc(conn, aws.StringValue(output.Site.GlobalNetworkId), aws.StringValue(output.Site.SiteId)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for networkmanager Site (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsNetworkManagerSiteRead(d, meta)
}

func resourceAwsNetworkManagerSiteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	site, err := networkmanagerDescribeSite(conn, d.Get("global_network_id").(string), d.Id())

	if isAWSErr(err, "InvalidSiteID.NotFound", "") {
		log.Printf("[WARN] networkmanager Site (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading networkmanager Site: %s", err)
	}

	if site == nil {
		log.Printf("[WARN] networkmanager Site (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(site.State) == networkmanager.SiteStateDeleting {
		log.Printf("[WARN] networkmanager Site (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(site.State))
		d.SetId("")
		return nil
	}

	d.Set("arn", site.SiteArn)
	d.Set("description", site.Description)

	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(site.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsNetworkManagerSiteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating networkmanager Site (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsNetworkManagerSiteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeleteSiteInput{
		GlobalNetworkId: aws.String( d.Get("global_network_id").(string)),
		SiteId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting networkmanager Site (%s): %s", d.Id(), input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteSite(input)

		if isAWSErr(err, "IncorrectState", "has non-deleted Transit Gateway Registrations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Customer Gateway Associations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Device") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Link") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Site") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteSite(input)
	}

	if isAWSErr(err, "InvalidSiteID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting networkmanager Site: %s", err)
	}

	if err := waitForNetworkManagerSiteDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for networkmanager Site (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAwsNetworkManagerSiteLocation(d *schema.ResourceData) *networkmanager.Location {
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

func networkmanagerSiteRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID, siteID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		site, err := networkmanagerDescribeSite(conn, globalNetworkID, siteID)

		if isAWSErr(err, "InvalidSiteID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading NetworkManager Site (%s): %s", siteID, err)
		}

		if site == nil {
			return nil, "DELETED", nil
		}

		return site, aws.StringValue(site.State), nil
	}
}

func networkmanagerDescribeSite(conn *networkmanager.NetworkManager, globalNetworkID, siteID string) (*networkmanager.Site, error) {
	input := &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		SiteIds:         []*string{aws.String(siteID)},
	}

	log.Printf("[DEBUG] Reading NetworkManager Site (%s): %s", siteID, input)
	for {
		output, err := conn.DescribeSites(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.Sites) == 0 {
			return nil, nil
		}

		for _, site := range output.Sites {
			if site == nil {
				continue
			}

			if aws.StringValue(site.SiteId) == siteID {
				return site, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerSiteDeletion(conn *networkmanager.NetworkManager, siteID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.SiteStateAvailable,
			networkmanager.SiteStateDeleting,
		},
		Target:         []string{""},
		Refresh:        networkmanagerSiteRefreshFunc(conn, siteID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for NetworkManager Transit Gateway (%s) deletion", siteID)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}
