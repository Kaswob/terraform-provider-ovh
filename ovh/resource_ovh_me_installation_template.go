package ovh

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/ovh/go-ovh/ovh"
)

func resourceMeInstallationTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceMeInstallationTemplateCreate,
		Read:   resourceMeInstallationTemplateRead,
		Update: resourceMeInstallationTemplateUpdate,
		Delete: resourceMeInstallationTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMeInstallationTemplateImportState,
		},

		Schema: map[string]*schema.Schema{
			"base_template_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "OVH template name yours will be based on, choose one among the list given by compatibleTemplates function",
			},
			"default_language": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The default language of this template",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					err := validateLanguageCode(v.(string))
					if err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
			"template_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "This template name",
			},

			"remove_default_partition_schemes": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Remove default partition schemes at creation",
			},

			"customization": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"change_log": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Template change log details",
						},
						"custom_hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Set up the server using the provided hostname instead of the default hostname",
						},
						"post_installation_script_link": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Indicate the URL where your postinstall customisation script is located",
						},
						"post_installation_script_return": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "indicate the string returned by your postinstall customisation script on successful execution. Advice: your script should return a unique validation string in case of succes. A good example is 'loh1Xee7eo OK OK OK UGh8Ang1Gu'",
						},
						"rating": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ssh_key_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the ssh key that should be installed. Password login will be disabled",
						},
						"use_distribution_kernel": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Use the distribution's native kernel instead of the recommended OVH Kernel",
						},
					},
				},
			},

			"available_languages": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all language available for this template",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"beta": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution is new and, although tested and functional, may still display odd behaviour",
			},
			"bit_format": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "This template bit format (32 or 64)",
			},
			"category": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Category of this template (informative only). (basic, customer, hosting, other, readyToUse, virtualisation)",
			},
			"deprecated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "is this distribution deprecated",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "information about this template",
			},
			"distribution": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the distribution this template is based on",
			},
			"family": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "this template family type (bsd,linux,solaris,windows)",
			},
			"hard_raid_configuration": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution supports hardware raid configuration through the OVH API",
			},
			"filesystems": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Filesystems available (btrfs,ext3,ext4,ntfs,reiserfs,swap,ufs,xfs,zfs)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"last_modification": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of last modification of the base image",
			},
			"supports_distribution_kernel": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution supports installation using the distribution's native kernel instead of the recommended OVH kernel",
			},
			"supports_gpt_label": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution supports the GUID Partition Table (GPT), providing up to 128 partitions that can have more than 2TB",
			},
			"supports_rtm": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution supports RTM software",
			},
			"supports_sql_server": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This distribution supports the microsoft SQL server",
			},
			"supports_uefi": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This distribution supports UEFI setup (no,only,yes)",
			},
		},
	}
}

func resourceMeInstallationTemplateImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	givenId := d.Id()
	splitId := strings.SplitN(givenId, "/", 2)
	if len(splitId) != 2 {
		return nil, fmt.Errorf("Import Id is not base_template_name/template_name formatted")
	}
	baseTemplateName := splitId[0]
	templateName := splitId[1]
	d.SetId(templateName)
	d.Set("base_template_name", baseTemplateName)
	d.Set("template_name", templateName)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceMeInstallationTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	opts := (&InstallationTemplateCreateOpts{}).FromResource(d)

	endpoint := "/me/installationTemplate"

	// the resource is created via the POST endpoint, then updated
	// via the PUT endpoint to apply customizations.
	// Thus we need to enable the Partial mode
	d.Partial(true)
	if err := config.OVHClient.Post(endpoint, opts, nil); err != nil {
		return fmt.Errorf("Error calling POST %s with opts %v:\n\t %q", endpoint, opts, err)
	}

	d.SetPartial("template_name")
	d.SetPartial("base_template_name")
	d.SetPartial("default_language")

	d.SetId(d.Get("template_name").(string))

	// We call the update method to put customization opts
	updateOpts := (&InstallationTemplateUpdateOpts{}).FromResource(d)
	endpoint = fmt.Sprintf(
		"/me/installationTemplate/%s",
		url.PathEscape(d.Id()),
	)

	if err := config.OVHClient.Put(endpoint, updateOpts, nil); err != nil {
		return fmt.Errorf("Error calling PUT %s with opts %v:\n\t %q", endpoint, opts, err)
	}
	d.SetPartial("customization")

	// handle remove_default_partitions option
	removeDefaultPartitions := false

	if v, ok := d.GetOk("remove_default_partition_schemes"); ok {
		removeDefaultPartitions = v.(bool)
	}

	d.Set("remove_default_partitions", removeDefaultPartitions)

	if removeDefaultPartitions {
		templateName := d.Get("template_name").(string)
		defaultSchemes, err := getPartitionSchemeIds(templateName, config.OVHClient)
		if err != nil {
			return err
		}

		for _, scheme := range defaultSchemes {
			endpoint := fmt.Sprintf(
				"/me/installationTemplate/%s/partitionScheme/%s",
				url.PathEscape(templateName),
				url.PathEscape(scheme),
			)

			if err := config.OVHClient.Delete(endpoint, nil); err != nil {
				return fmt.Errorf("Error calling DELETE %s: %s \n", endpoint, err.Error())
			}
		}
	}

	d.Partial(false)

	return resourceMeInstallationTemplateRead(d, meta)
}

func resourceMeInstallationTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	opts := (&InstallationTemplateUpdateOpts{}).FromResource(d)

	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s",
		url.PathEscape(d.Id()),
	)

	if err := config.OVHClient.Put(endpoint, opts, nil); err != nil {
		return fmt.Errorf("Error calling PUT %s with opts %v:\n\t %q", endpoint, opts, err)
	}

	return resourceMeInstallationTemplateRead(d, meta)
}

func resourceMeInstallationTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	name := d.Get("template_name").(string)

	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s",
		url.PathEscape(name),
	)

	if err := config.OVHClient.Delete(endpoint, nil); err != nil {
		return fmt.Errorf("Error calling DELETE %s: %s \n", endpoint, err.Error())
	}

	return nil
}

func resourceMeInstallationTemplateRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	name := d.Id()

	template, err := getInstallationTemplate(name, config.OVHClient)
	if err != nil {
		return err
	}

	// set attributes
	for k, v := range template.ToMap() {
		d.Set(k, v)
	}

	d.SetId(name)

	return nil
}

func getPartitionSchemePartitions(template, scheme string, client *ovh.Client) ([]*Partition, error) {
	mountPoints := []string{}
	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s/partitionScheme/%s/partition",
		url.PathEscape(template),
		url.PathEscape(scheme),
	)
	err := client.Get(endpoint, &mountPoints)

	if err != nil {
		return nil, fmt.Errorf("Error calling GET %s: %s \n", endpoint, err.Error())
	}

	partitions := []*Partition{}
	for _, mountPoint := range mountPoints {
		partition, err := getPartitionSchemePartition(template, scheme, mountPoint, client)
		if err != nil {
			return nil, err
		}

		partitions = append(partitions, partition)
	}

	return partitions, nil
}

func getPartitionSchemeHardwareRaids(template, scheme string, client *ovh.Client) ([]*HardwareRaid, error) {
	names := []string{}
	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s/partitionScheme/%s/hardwareRaid",
		url.PathEscape(template),
		url.PathEscape(scheme),
	)
	err := client.Get(endpoint, &names)

	if err != nil {
		return nil, fmt.Errorf("Error calling GET %s: %s \n", endpoint, err.Error())
	}

	hardwareRaids := []*HardwareRaid{}
	for _, name := range names {
		hardwareRaid, err := getPartitionSchemeHardwareRaid(template, scheme, name, client)
		if err != nil {
			return nil, err
		}

		hardwareRaids = append(hardwareRaids, hardwareRaid)
	}

	return hardwareRaids, nil
}

func getPartitionSchemes(template string, client *ovh.Client) ([]*PartitionScheme, error) {
	schemes, err := getPartitionSchemeIds(template, client)
	if err != nil {
		return nil, err
	}

	partitionSchemes := []*PartitionScheme{}
	for _, scheme := range schemes {
		partitionScheme, err := getPartitionScheme(template, scheme, client)
		if err != nil {
			return nil, err
		}

		partitionSchemes = append(partitionSchemes, partitionScheme)
	}

	return partitionSchemes, nil
}

func getPartitionSchemeIds(template string, client *ovh.Client) ([]string, error) {
	schemes := []string{}
	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s/partitionScheme",
		url.PathEscape(template),
	)
	err := client.Get(endpoint, &schemes)

	if err != nil {
		return nil, fmt.Errorf("Error calling GET %s: %s \n", endpoint, err.Error())
	}
	return schemes, nil
}

func getInstallationTemplate(template string, client *ovh.Client) (*InstallationTemplate, error) {
	r := &InstallationTemplate{}

	endpoint := fmt.Sprintf(
		"/me/installationTemplate/%s",
		url.PathEscape(template),
	)

	err := client.Get(endpoint, &r)

	if err != nil {
		return nil, fmt.Errorf("Error calling GET %s: %s \n", endpoint, err.Error())
	}

	return r, nil
}
