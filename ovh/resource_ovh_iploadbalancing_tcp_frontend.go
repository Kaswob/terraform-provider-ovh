package ovh

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceIpLoadbalancingTcpFrontend() *schema.Resource {
	return &schema.Resource{
		Create: resourceIpLoadbalancingTcpFrontendCreate,
		Read:   resourceIpLoadbalancingTcpFrontendRead,
		Update: resourceIpLoadbalancingTcpFrontendUpdate,
		Delete: resourceIpLoadbalancingTcpFrontendDelete,
		Importer: &schema.ResourceImporter{
			State: resourceIpLoadbalancingTcpFrontendImportState,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"allowed_source": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"dedicated_ipfo": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"default_farm_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"default_ssl_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
				ForceNew: false,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				ForceNew: false,
			},
			"ssl": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				ForceNew: false,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

func resourceIpLoadbalancingTcpFrontendImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	givenId := d.Id()
	splitId := strings.SplitN(givenId, "/", 2)
	if len(splitId) != 2 {
		return nil, fmt.Errorf("Import Id is not service_name/frontend id formatted")
	}
	serviceName := splitId[0]
	frontendId := splitId[1]
	d.SetId(frontendId)
	d.Set("service_name", serviceName)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceIpLoadbalancingTcpFrontendCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	allowedSources := stringsFromSchema(d, "allowed_source")
	dedicatedIpFo := stringsFromSchema(d, "dedicated_ipfo")

	for _, s := range allowedSources {
		if err := validateIpBlock(s); err != nil {
			return fmt.Errorf("Error validating `allowed_source` value: %s", err)
		}
	}

	for _, s := range dedicatedIpFo {
		if err := validateIpBlock(s); err != nil {
			return fmt.Errorf("Error validating `dedicated_ipfo` value: %s", err)
		}
	}

	frontend := &IpLoadbalancingTcpFrontend{
		Port:          d.Get("port").(string),
		Zone:          d.Get("zone").(string),
		AllowedSource: allowedSources,
		DedicatedIpFo: dedicatedIpFo,
		Disabled:      getNilBoolPointerFromData(d, "disabled"),
		Ssl:           getNilBoolPointerFromData(d, "ssl"),
		DisplayName:   d.Get("display_name").(string),
	}

	frontend.DefaultFarmId = getNilIntPointerFromData(d, "default_farm_id")
	frontend.DefaultSslId = getNilIntPointerFromData(d, "default_ssl_id")

	service := d.Get("service_name").(string)
	resp := &IpLoadbalancingTcpFrontend{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/tcp/frontend", service)

	err := config.OVHClient.Post(endpoint, frontend, resp)
	if err != nil {
		return fmt.Errorf("calling POST %s:\n\t %s", endpoint, err.Error())
	}
	return readIpLoadbalancingTcpFrontend(resp, d)

}

func resourceIpLoadbalancingTcpFrontendRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	service := d.Get("service_name").(string)
	r := &IpLoadbalancingTcpFrontend{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/tcp/frontend/%s", service, d.Id())

	err := config.OVHClient.Get(endpoint, &r)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %s", endpoint, err.Error())
	}
	return readIpLoadbalancingTcpFrontend(r, d)
}

func resourceIpLoadbalancingTcpFrontendUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	service := d.Get("service_name").(string)
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/tcp/frontend/%s", service, d.Id())

	allowedSources := stringsFromSchema(d, "allowed_source")
	dedicatedIpFo := stringsFromSchema(d, "dedicated_ipfo")

	for _, s := range allowedSources {
		if err := validateIpBlock(s); err != nil {
			return fmt.Errorf("Error validating `allowed_source` value: %s", err)
		}
	}

	for _, s := range dedicatedIpFo {
		if err := validateIpBlock(s); err != nil {
			return fmt.Errorf("Error validating `dedicated_ipfo` value: %s", err)
		}
	}

	frontend := &IpLoadbalancingTcpFrontend{
		Port:          d.Get("port").(string),
		Zone:          d.Get("zone").(string),
		AllowedSource: allowedSources,
		DedicatedIpFo: dedicatedIpFo,
		Disabled:      getNilBoolPointerFromData(d, "disabled"),
		Ssl:           getNilBoolPointerFromData(d, "ssl"),
		DisplayName:   d.Get("display_name").(string),
	}

	frontend.DefaultFarmId = getNilIntPointerFromData(d, "default_farm_id")
	frontend.DefaultSslId = getNilIntPointerFromData(d, "default_ssl_id")

	err := config.OVHClient.Put(endpoint, frontend, nil)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %s", endpoint, err.Error())
	}
	return nil
}

func readIpLoadbalancingTcpFrontend(r *IpLoadbalancingTcpFrontend, d *schema.ResourceData) error {
	d.Set("display_name", r.DisplayName)
	d.Set("port", r.Port)
	d.Set("zone", r.Zone)

	allowedSources := make([]string, 0)
	allowedSources = append(allowedSources, r.AllowedSource...)
	d.Set("allowed_source", allowedSources)

	dedicatedIpFos := make([]string, 0)
	dedicatedIpFos = append(dedicatedIpFos, r.DedicatedIpFo...)
	d.Set("dedicated_ipfo", dedicatedIpFos)

	if r.DefaultFarmId != nil {
		d.Set("default_farm_id", r.DefaultFarmId)
	}

	if r.DefaultSslId != nil {
		d.Set("default_ssl_id", r.DefaultSslId)
	}
	if r.Disabled != nil {
		d.Set("disabled", r.Disabled)
	}
	if r.Ssl != nil {
		d.Set("ssl", r.Ssl)
	}

	d.SetId(fmt.Sprintf("%d", r.FrontendId))

	return nil
}

func resourceIpLoadbalancingTcpFrontendDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	service := d.Get("service_name").(string)
	r := &IpLoadbalancingTcpFrontend{}
	endpoint := fmt.Sprintf("/ipLoadbalancing/%s/tcp/frontend/%s", service, d.Id())

	err := config.OVHClient.Delete(endpoint, &r)
	if err != nil {
		return fmt.Errorf("Error calling %s: %s \n", endpoint, err.Error())
	}

	return nil
}
