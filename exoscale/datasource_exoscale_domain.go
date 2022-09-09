package exoscale

import (
	"context"

	exo "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the Domain",
				Required:    true,
			},
		},
		ReadContext: dataSourceDomainRead,
	}
}

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "beginning read", map[string]interface{}{
		"id": resourceIDString(d, "exoscale_domain"),
	})

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutRead))
	ctx = exoapi.WithEndpoint(ctx, exoapi.NewReqEndpoint(getEnvironment(meta), defaultZone))
	defer cancel()

	client := GetComputeClient(meta)

	domainName := d.Get("name")
	var domain *exo.DNSDomain

	domains, err := client.ListDNSDomains(ctx, defaultZone)
	if err != nil {
		return diag.Errorf("error retrieving domain list: %s", err)
	}

	for _, item := range domains {
		if *item.UnicodeName == domainName {
			t := item
			domain = &t
			break
		}
	}

	if domain == nil {
		return diag.Errorf("domain %q not found", domainName)
	}

	d.SetId(*domain.ID)

	err = d.Set("name", domain.UnicodeName)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "read finished successfully", map[string]interface{}{
		"id": resourceIDString(d, "exoscale_domain"),
	})

	return nil
}
