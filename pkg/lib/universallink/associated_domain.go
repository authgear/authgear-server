package universallink

import "github.com/authgear/authgear-server/pkg/lib/config"

type AssociatedDomainsService struct {
	OAuthConfig *config.OAuthConfig
}

func (s *AssociatedDomainsService) PopulateIOSAssociatedDomains(data map[string]interface{}) {
	appLinksDetails := []interface{}{}

	for _, c := range s.OAuthConfig.Clients {
		if !c.UniversalLink.IOS.Enabled {
			continue
		}

		appLinksDetails = append(appLinksDetails, map[string]interface{}{
			"appID": c.UniversalLink.IOS.BundleID,
			"paths": []interface{}{"/clients/" + c.ClientID + "/flows/verify_login_link"},
		})
	}

	data["app_links"] = map[string]interface{}{
		"details": appLinksDetails,
	}
}

func (s *AssociatedDomainsService) PopulateAndroidAssociatedDomains(data *[]interface{}) {
	for _, c := range s.OAuthConfig.Clients {
		if !c.UniversalLink.Android.Enabled {
			continue
		}

		*data = append(*data, map[string]interface{}{
			"relation": []interface{}{"delegate_permission/common.handle_all_urls"},
			"target": map[string]interface{}{
				"namespace":                "android_app",
				"package_name":             c.UniversalLink.Android.BundleID,
				"sha256_cert_fingerprints": []interface{}{c.UniversalLink.Android.ShaSignature},
			},
		})
	}
}
