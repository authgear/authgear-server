package sso

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL   string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinMeURL      string = "https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))"
	linkedinContactURL string = "https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))"
)

type LinkedInImpl struct {
	ProviderConfig               oauthrelyingparty.ProviderConfig
	ClientSecret                 string
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (f *LinkedInImpl) Config() oauthrelyingparty.ProviderConfig {
	return f.ProviderConfig
}

func (f *LinkedInImpl) GetAuthorizationURL(param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return oauthrelyingpartyutil.MakeAuthorizationURL(linkedinAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     f.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        f.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Linkedin doesn't support prompt parameter
		// https://docs.microsoft.com/en-us/linkedin/shared/authentication/authorization-code-flow?tabs=HTTPS#step-2-request-an-authorization-code

		// Nonce is unset
	}.Query()), nil
}

func (f *LinkedInImpl) GetUserProfile(param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	accessTokenResp, err := oauthrelyingpartyutil.FetchAccessTokenResp(
		f.HTTPClient.Client,
		param.Code,
		linkedinTokenURL,
		param.RedirectURI,
		f.ProviderConfig.ClientID(),
		f.ClientSecret,
	)
	if err != nil {
		return
	}

	meResponse, err := oauthrelyingpartyutil.FetchUserProfile(f.HTTPClient.Client, accessTokenResp, linkedinMeURL)
	if err != nil {
		return
	}

	contactResponse, err := oauthrelyingpartyutil.FetchUserProfile(f.HTTPClient.Client, accessTokenResp, linkedinContactURL)
	if err != nil {
		return
	}

	// {
	//     "primary_contact": {
	//         "elements": [
	//             {
	//                 "handle": "urn:li:emailAddress:redacted",
	//                 "handle~": {
	//                     "emailAddress": "user@example.com"
	//                 },
	//                 "primary": true,
	//                 "type": "EMAIL"
	//             }
	//         ]
	//     },
	//     "profile": {
	//         "id": "redacted",
	//         "localizedFirstName": "redacted",
	//         "localizedLastName": "redacted",
	//         "profilePicture": {
	//             "displayImage": "urn:li:digitalmediaAsset:redacted",
	//             "displayImage~": {
	//                 "elements": [
	//                     {
	//                         "artifact": "urn:li:digitalmediaMediaArtifact:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_100_100)",
	//                         "authorizationMethod": "PUBLIC",
	//                         "data": {
	//                             "com.linkedin.digitalmedia.mediaartifact.StillImage": {
	//                                 "displayAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "displaySize": {
	//                                     "height": 100,
	//                                     "uom": "PX",
	//                                     "width": 100
	//                                 },
	//                                 "mediaType": "image/jpeg",
	//                                 "rawCodecSpec": {
	//                                     "name": "jpeg",
	//                                     "type": "image"
	//                                 },
	//                                 "storageAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "storageSize": {
	//                                     "height": 100,
	//                                     "width": 100
	//                                 }
	//                             }
	//                         },
	//                         "identifiers": [
	//                             {
	//                                 "file": "urn:li:digitalmediaFile:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_100_100,0)",
	//                                 "identifier": "https://media-exp1.licdn.com/dms/image/C5603AQE9WylLgWcyuA/profile-displayphoto-shrink_100_100/0/1631684043723?e=1637193600&v=beta&t=h8Wz-EdTjSD9FxQL_oO6hrQ4DdwzGfl5fPPe2cEDPIs",
	//                                 "identifierExpiresInSeconds": 1637193600,
	//                                 "identifierType": "EXTERNAL_URL",
	//                                 "index": 0,
	//                                 "mediaType": "image/jpeg"
	//                             }
	//                         ]
	//                     },
	//                     {
	//                         "artifact": "urn:li:digitalmediaMediaArtifact:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_200_200)",
	//                         "authorizationMethod": "PUBLIC",
	//                         "data": {
	//                             "com.linkedin.digitalmedia.mediaartifact.StillImage": {
	//                                 "displayAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "displaySize": {
	//                                     "height": 200,
	//                                     "uom": "PX",
	//                                     "width": 200
	//                                 },
	//                                 "mediaType": "image/jpeg",
	//                                 "rawCodecSpec": {
	//                                     "name": "jpeg",
	//                                     "type": "image"
	//                                 },
	//                                 "storageAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "storageSize": {
	//                                     "height": 200,
	//                                     "width": 200
	//                                 }
	//                             }
	//                         },
	//                         "identifiers": [
	//                             {
	//                                 "file": "urn:li:digitalmediaFile:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_200_200,0)",
	//                                 "identifier": "https://media-exp1.licdn.com/dms/image/C5603AQE9WylLgWcyuA/profile-displayphoto-shrink_200_200/0/1631684043723?e=1637193600&v=beta&t=8CDBMjGCkpk_CO8VgAkVXWeKAu8gYiUTTXPbtMazMUg",
	//                                 "identifierExpiresInSeconds": 1637193600,
	//                                 "identifierType": "EXTERNAL_URL",
	//                                 "index": 0,
	//                                 "mediaType": "image/jpeg"
	//                             }
	//                         ]
	//                     },
	//                     {
	//                         "artifact": "urn:li:digitalmediaMediaArtifact:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_400_400)",
	//                         "authorizationMethod": "PUBLIC",
	//                         "data": {
	//                             "com.linkedin.digitalmedia.mediaartifact.StillImage": {
	//                                 "displayAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "displaySize": {
	//                                     "height": 400,
	//                                     "uom": "PX",
	//                                     "width": 400
	//                                 },
	//                                 "mediaType": "image/jpeg",
	//                                 "rawCodecSpec": {
	//                                     "name": "jpeg",
	//                                     "type": "image"
	//                                 },
	//                                 "storageAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "storageSize": {
	//                                     "height": 400,
	//                                     "width": 400
	//                                 }
	//                             }
	//                         },
	//                         "identifiers": [
	//                             {
	//                                 "file": "urn:li:digitalmediaFile:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_400_400,0)",
	//                                 "identifier": "https://media-exp1.licdn.com/dms/image/C5603AQE9WylLgWcyuA/profile-displayphoto-shrink_400_400/0/1631684043723?e=1637193600&v=beta&t=9tCLl0cAbswfKYUgJqDN41QT368cFsq_7TeXyPjixOY",
	//                                 "identifierExpiresInSeconds": 1637193600,
	//                                 "identifierType": "EXTERNAL_URL",
	//                                 "index": 0,
	//                                 "mediaType": "image/jpeg"
	//                             }
	//                         ]
	//                     },
	//                     {
	//                         "artifact": "urn:li:digitalmediaMediaArtifact:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_800_800)",
	//                         "authorizationMethod": "PUBLIC",
	//                         "data": {
	//                             "com.linkedin.digitalmedia.mediaartifact.StillImage": {
	//                                 "displayAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "displaySize": {
	//                                     "height": 800,
	//                                     "uom": "PX",
	//                                     "width": 800
	//                                 },
	//                                 "mediaType": "image/jpeg",
	//                                 "rawCodecSpec": {
	//                                     "name": "jpeg",
	//                                     "type": "image"
	//                                 },
	//                                 "storageAspectRatio": {
	//                                     "formatted": "1.00:1.00",
	//                                     "heightAspect": 1,
	//                                     "widthAspect": 1
	//                                 },
	//                                 "storageSize": {
	//                                     "height": 800,
	//                                     "width": 800
	//                                 }
	//                             }
	//                         },
	//                         "identifiers": [
	//                             {
	//                                 "file": "urn:li:digitalmediaFile:(urn:li:digitalmediaAsset:C5603AQE9WylLgWcyuA,urn:li:digitalmediaMediaArtifactClass:profile-displayphoto-shrink_800_800,0)",
	//                                 "identifier": "https://media-exp1.licdn.com/dms/image/C5603AQE9WylLgWcyuA/profile-displayphoto-shrink_800_800/0/1631684043723?e=1637193600&v=beta&t=hvhZcRfvDrgE64iXNX1J2eWUMAytTtD8SdD006lc3_o",
	//                                 "identifierExpiresInSeconds": 1637193600,
	//                                 "identifierType": "EXTERNAL_URL",
	//                                 "index": 0,
	//                                 "mediaType": "image/jpeg"
	//                             }
	//                         ]
	//                     }
	//                 ],
	//                 "paging": {
	//                     "count": 10,
	//                     "links": [],
	//                     "start": 0
	//                 }
	//             }
	//         }
	//     }
	// }
	combinedResponse := map[string]interface{}{
		"profile":         meResponse,
		"primary_contact": contactResponse,
	}

	authInfo.ProviderRawProfile = combinedResponse
	id, attrs := decodeLinkedIn(combinedResponse)
	authInfo.ProviderUserID = id

	emailRequired := f.ProviderConfig.EmailClaimConfig().Required()
	attrs, err = stdattrs.Extract(attrs, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = attrs

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func decodeLinkedIn(userInfo map[string]interface{}) (string, stdattrs.T) {
	profile := userInfo["profile"].(map[string]interface{})
	id := profile["id"].(string)

	// Extract email
	email := ""
	{
		primaryContact, _ := userInfo["primary_contact"].(map[string]interface{})
		elements, _ := primaryContact["elements"].([]interface{})
		for _, e := range elements {
			element, _ := e.(map[string]interface{})
			if primary, ok := element["primary"].(bool); !ok || !primary {
				continue
			}
			if typ, ok := element["type"].(string); !ok || typ != "EMAIL" {
				continue
			}
			handleTilde, ok := element["handle~"].(map[string]interface{})
			if !ok {
				continue
			}
			email, _ = handleTilde["emailAddress"].(string)
		}
	}

	// Extract given_name and family_name
	firstName, _ := profile["localizedFirstName"].(string)
	lastName, _ := profile["localizedLastName"].(string)

	// Extract picture
	var picture string
	{
		profilePicture, _ := profile["profilePicture"].(map[string]interface{})
		displayImage, _ := profilePicture["displayImage~"].(map[string]interface{})
		elements, _ := displayImage["elements"].([]interface{})
		if len(elements) > 0 {
			lastElementIface := elements[len(elements)-1]
			lastElement, _ := lastElementIface.(map[string]interface{})
			identifiers, _ := lastElement["identifiers"].([]interface{})
			if len(identifiers) > 0 {
				firstIdentifierIface := identifiers[0]
				firstIdentifier, _ := firstIdentifierIface.(map[string]interface{})
				picture, _ = firstIdentifier["identifier"].(string)
			}
		}
	}

	return id, stdattrs.T{
		stdattrs.Email:      email,
		stdattrs.GivenName:  firstName,
		stdattrs.FamilyName: lastName,
		stdattrs.Picture:    picture,
	}
}

var (
	_ OAuthProvider = &LinkedInImpl{}
)
