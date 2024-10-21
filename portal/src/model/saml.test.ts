import { describe, it, expect } from "@jest/globals";
import { parseServiceProviderMetadata } from "./saml";

const testMetadata = `
<?xml version='1.0' encoding='UTF-8'?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
  xmlns:ns2="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:ns3="http://www.w3.org/2000/09/xmldsig#"
  xmlns:ns4="http://www.w3.org/2001/04/xmlenc#" ID="test"
  entityID="http://portal.authgear.com">
  <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
    <ds:SignedInfo>
      <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
      <ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" />
      <ds:Reference URI="#test">
        <ds:Transforms>
          <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature" />
          <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#" />
        </ds:Transforms>
        <ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256" />
        <ds:DigestValue>test</ds:DigestValue>
      </ds:Reference>
    </ds:SignedInfo>
    <ds:SignatureValue>
      test</ds:SignatureValue>
    <ds:KeyInfo>
      <ds:X509Data>
        <ds:X509Certificate>
          test</ds:X509Certificate>
      </ds:X509Data>
      <ds:KeyValue>
        <ds:RSAKeyValue>
          <ds:Modulus>
            test</ds:Modulus>
          <ds:Exponent>AQAB</ds:Exponent>
        </ds:RSAKeyValue>
      </ds:KeyValue>
    </ds:KeyInfo>
  </ds:Signature>
  <IDPSSODescriptor WantAuthnRequestsSigned="true"
    protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <ns3:KeyInfo>
        <ns3:KeyName>http://portal.authgear.com/idp/signing</ns3:KeyName>
        <ns3:X509Data>
          <ns3:X509Certificate>
            testidpsigningcertificate</ns3:X509Certificate>
        </ns3:X509Data>
      </ns3:KeyInfo>
    </KeyDescriptor>
    <KeyDescriptor use="encryption">
      <ns3:KeyInfo>
        <ns3:KeyName>http://portal.authgear.com/idp/encryption</ns3:KeyName>
        <ns3:X509Data>
          <ns3:X509Certificate>
            testidpencryptioncertificate</ns3:X509Certificate>
        </ns3:X509Data>
      </ns3:KeyInfo>
    </KeyDescriptor>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
      Location="http://portal.authgear.com/slo" />
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="http://portal.authgear.com/slo" />
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
      Location="http://portal.authgear.com/sso" />
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="http://portal.authgear.com/sso" />
  </IDPSSODescriptor>
  <SPSSODescriptor AuthnRequestsSigned="true"
    protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <ns3:KeyInfo>
        <ns3:KeyName>http://portal.authgear.com/sp/signing</ns3:KeyName>
        <ns3:X509Data>
          <ns3:X509Certificate>
            testspsigningcertificate</ns3:X509Certificate>
        </ns3:X509Data>
      </ns3:KeyInfo>
    </KeyDescriptor>
    <KeyDescriptor use="encryption">
      <ns3:KeyInfo>
        <ns3:KeyName>http://portal.authgear.com/sp/encryption</ns3:KeyName>
        <ns3:X509Data>
          <ns3:X509Certificate>
            testspencryptioncertificate</ns3:X509Certificate>
        </ns3:X509Data>
      </ns3:KeyInfo>
    </KeyDescriptor>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
      Location="http://portal.authgear.com/sp/slo" />
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="http://portal.authgear.com/sp/slo" />
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="http://portal.authgear.com/sp/acs"
      index="0" isDefault="true" />
  </SPSSODescriptor>
</EntityDescriptor>
`;

describe("parseServiceProviderMetadata", () => {
  it("should parse a service provider metadata", () => {
    const parseResult = parseServiceProviderMetadata(testMetadata);
    expect(parseResult).toEqual({
      acsURL: "http://portal.authgear.com/sp/acs",
      sloEnabled: true,
      sloCallbackURL: "http://portal.authgear.com/sp/slo",
      sloCallbackBinding: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
      authnRequestsSigned: true,
      certificate:
        "-----BEGIN CERTIFICATE-----\ntestspsigningcertificate\n-----END CERTIFICATE-----",
    });
  });
});
