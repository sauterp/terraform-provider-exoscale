package exoscale

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/terraform-provider-exoscale/pkg/config"
	"github.com/exoscale/terraform-provider-exoscale/pkg/general"
)

const (
	resAffinityDeprecationMessage = `**WARNING:** This resource is **DEPRECATED** and will be removed in the next major version. Please use [exoscale_anti_affinity_group](./anti_affinity_group.md) instead.`
)

func resourceAffinityIDString(d general.ResourceIDStringer) string {
	return general.ResourceIDString(d, "exoscale_affinity")
}

func resourceAffinity() *schema.Resource {
	return &schema.Resource{
		Description:        "Manage Exoscale Anti-Affinity Groups.",
		DeprecationMessage: resAffinityDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The anti-affinity group name.",
			},
			"description": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "A free-form text describing the group.",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "host anti-affinity",
				Description: "The type of the group (`host anti-affinity` is the only supported value).",
			},
			"virtual_machine_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The compute instances (IDs) members of the group.",
			},
		},

		Create: resourceAffinityCreate,
		Read:   resourceAffinityRead,
		Delete: resourceAffinityDelete,
		Exists: resourceAffinityExists,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(config.DefaultTimeout),
			Read:   schema.DefaultTimeout(config.DefaultTimeout),
			Delete: schema.DefaultTimeout(config.DefaultTimeout),
		},
	}
}

func resourceAffinityCreate(d *schema.ResourceData, meta interface{}) error {
	tflog.Debug(context.Background(), "beginning create", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	client := GetComputeClient(meta)

	req := &egoscale.CreateAffinityGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	resp, err := client.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}

	ag := resp.(*egoscale.AffinityGroup)
	d.SetId(ag.ID.String())

	tflog.Debug(context.Background(), "create finished successfully", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	return resourceAffinityRead(d, meta)
}

func resourceAffinityExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	client := GetComputeClient(meta)

	id, err := egoscale.ParseUUID(d.Id())
	if err != nil {
		return false, err
	}

	ag := &egoscale.AffinityGroup{ID: id}
	_, err = client.GetWithContext(ctx, ag)
	if err != nil {
		e := handleNotFound(d, err)
		return d.Id() != "", e
	}

	return true, nil
}

func resourceAffinityRead(d *schema.ResourceData, meta interface{}) error {
	tflog.Debug(context.Background(), "beginning read", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	client := GetComputeClient(meta)

	id, err := egoscale.ParseUUID(d.Id())
	if err != nil {
		return err
	}

	ag := &egoscale.AffinityGroup{ID: id}

	resp, err := client.GetWithContext(ctx, ag)
	if err != nil {
		return handleNotFound(d, err)
	}

	tflog.Debug(context.Background(), "read finished successfully", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	return resourceAffinityApply(d, resp.(*egoscale.AffinityGroup))
}

func resourceAffinityDelete(d *schema.ResourceData, meta interface{}) error {
	tflog.Debug(context.Background(), "beginning delete", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	client := GetComputeClient(meta)

	id, err := egoscale.ParseUUID(d.Id())
	if err != nil {
		return err
	}

	ag := &egoscale.AffinityGroup{ID: id}

	if err := client.DeleteWithContext(ctx, ag); err != nil {
		return err
	}

	tflog.Debug(context.Background(), "delete finished successfully", map[string]interface{}{
		"id": resourceAffinityIDString(d),
	})

	return nil
}

func resourceAffinityApply(d *schema.ResourceData, affinity *egoscale.AffinityGroup) error {
	if err := d.Set("name", affinity.Name); err != nil {
		return err
	}
	if err := d.Set("description", affinity.Description); err != nil {
		return err
	}
	if err := d.Set("type", affinity.Type); err != nil {
		return err
	}
	ids := make([]string, len(affinity.VirtualMachineIDs))
	for i, id := range affinity.VirtualMachineIDs {
		ids[i] = id.String()
	}
	if err := d.Set("virtual_machine_ids", ids); err != nil {
		return err
	}

	return nil
}
