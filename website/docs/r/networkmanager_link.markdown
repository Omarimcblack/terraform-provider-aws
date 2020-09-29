---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_link"
description: |-
  Creates a link for a specified site in a global network.
---

# Resource: aws_networkmanager_link

Creates a link for a specified site in a global network.

## Example Usage

```hcl
resource "aws_networkmanager_global_network" "example" {
}

resource "aws_networkmanager_site" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}

resource "aws_networkmanager_link" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  site_id           = aws_networkmanager_site.example.id

  bandwidth {
    download_speed  = 10
    upload_speed    = 20
  }
}
```

## Argument Reference

The following arguments are supported:
* `global_network_id` - (Required) The ID of the Global Network to create the link in.
* `site_id` - (Required) The ID of the site to create the link for.
* `description` - (Optional) Description of the link.
* `bandwidth` - (Required) The link bandwidth as documented below.
* `service_provider` - (Optional) The provider of the link.
* `type` - (Optional) The type of link.
* `tags` - (Optional) Key-value tags for the link.

The `bandwidth` object supports the following:

* `download_speed` - (Required) Address of the location.
* `upload_speed` - (Required) Latitude of the location.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - link Amazon Resource Name (ARN)

## Import

`aws_networkmanager_link` can be imported using the link ARN, e.g.

```
$ terraform import aws_networkmanager_link.example arn:aws:networkmanager::123456789012:link/global-network-0d47f6t230mz46dy4/link-11112222aaaabbbb1
```
