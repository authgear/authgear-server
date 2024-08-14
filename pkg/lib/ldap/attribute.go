package ldap

import "fmt"

var DefaultAttributeRegistry = &AttributeRegistry{}

type AttributeRegistry struct {
	attributes map[string]Attribute
}

func (r *AttributeRegistry) registerKnownAttribute(attr Attribute) Attribute {
	if r.attributes == nil {
		r.attributes = make(map[string]Attribute)
	}
	r.attributes[attr.Name] = attr
	return attr
}

func (r *AttributeRegistry) Get(attributeName string) (Attribute, bool) {
	a, ok := r.attributes[attributeName]
	if !ok {
		return Attribute{}, false
	}
	return a, true
}

type AttributeType string

const (
	AttributeTypeString = "string"
	AttributeTypeUUID   = "uuid"
)

func (t AttributeType) Decoder() AttributeDecoder {
	switch t {
	case AttributeTypeString:
		return StringAttributeDecoder{}
	case AttributeTypeUUID:
		return UUIDAttributeDecoder{}
	default:
		panic(fmt.Errorf("ldap: Unknwon attribute type %s", t))
	}
}

type Attribute struct {
	Type AttributeType
	Name string
}

// Business category is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.1
var AttributeBusinesCategory = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "businessCategory",
})

// Country name is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.2
var AttributeCountryName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "c",
})

// Common name is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.3
var AttributeCommonName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "cn",
})

// Domain component is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.4
var AttributeDomainComponent = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "dc",
})

// Description is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.5
var AttributeDescription = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "description",
})

// DestinationIndicator is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.6
var AttributeDestinationIndicator = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "destinationIndicator",
})

// DN is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.7
var AttributeDN = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "distinguishedName",
})

// DN Qualifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.8
var AttributeDNQualifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "dnQualifier",
})

// Facsimile Telephone Number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.10
var AttributeFacsimileTelephoneNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "facsimileTelephoneNumber",
})

// Generation Qualifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.11
var AttributeGenerationQualifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "generationQualifier",
})

// Given name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.12
var AttributeGivenName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "givenName",
})

// House identifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.13
var AttributeHouseIdentifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "houseIdentifier",
})

// Initials is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.14
var AttributeInitials = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "initials",
})

// International ISDN Number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.15
var AttributeInternationalISDNNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "internationalISDNNumber",
})

// Locality name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.16
var AttributeLocalityName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "l",
})

// Member is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.17
var AttributeMember = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "member",
})

// Name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.18
var AttributeName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "name",
})

// Organization name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.19
var AttributeOrganizationName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "o",
})

// Organization Unit Name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.20
var AttributeOrganizationUnitName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "ou",
})

// Owner is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.21
var AttributeOwner = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "owner",
})

// Physical Delivery Office Name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.22
var AttributePhysicalDeliveryOfficeName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "physicalDeliveryOfficeName",
})

// Psotal address is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.23
var AttributePostalAddress = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "postalAddress",
})

// Postal code is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.24
var AttributePostalCode = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "postalCode",
})

// Post Office Box is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.25
var AttributePostOfficeBox = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "postOfficeBox",
})

// Preferred Delivery Method is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.26
var AttributePreferredDeliveryMethod = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "preferredDeliveryMethod",
})

// Registered Address is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.27
var AttributeRegisteredAddress = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "registeredAddress",
})

// Role Occupant is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.28
var AttributeRoleOccupant = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "roleOccupant",
})

// Serial Number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.31
var AttributeSerialNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "serialNumber",
})

// Surname is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.32
var AttributeSurname = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "sn",
})

// State or Province name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.33
var AttributeStateORProvinceName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "st",
})

// Street address is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.34
var AttributeStreetAddress = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "street",
})

// Telephone number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.35
var AttributeTelephoneNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "telephoneNumber",
})

// Teletex terminal identifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.36
var AttributeTeletexTerminalIdentifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "teletexTerminalIdentifier",
})

// Telex number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.37
var AttributeTelexNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "telexNumber",
})

// Title is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.38
var AttributeTitle = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "title",
})

// User ID is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.39
var AttributeUID = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "uid",
})

// uid number is from https://learn.microsoft.com/en-us/windows/win32/adschema/a-uidnumber
var AttributeUIDNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "uidNumber",
})

// Unique Member is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.40
var AttributeUniqueMember = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "uniqueMember",
})

// User Password is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.41
var AttributeUserPassword = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "userPassword",
})

// X121 Address is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.42
var AttributeX121Address = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "x121Address",
})

// x500 unique identifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.43
var AttributeX500UniqueIdentifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "x500UniqueIdentifier",
})

// Associated Domain is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.1
var AttributeAssociatedDomain = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "associatedDomain",
})

// Associated Name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.2
var AttributeAssociatedName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "associatedName",
})

// Building Name is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.3
var AttributeBuildingName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "buildingName",
})

// Co is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.4
var AttributeCo = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "co",
})

// Document Author is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.5
var AttributeDocumentAuthor = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentAuthor",
})

// Document Identifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.6
var AttributeDocumentIdentifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentIdentifier",
})

// Document Location is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.7
var AttributeDocumentLocation = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentLocation",
})

// Document Publisher is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.8
var AttributeDocumentPublisher = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentPublisher",
})

// Document Title is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.9
var AttributeDocumentTitle = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentTitle",
})

// Document Version is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.10
var AttributeDocumentVersion = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "documentVersion",
})

// Drink is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.11
var AttributeDrink = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "drink",
})

// Home Phone is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.12
var AttributeHomePhone = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "homePhone",
})

// Home Postal Address is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.13
var AttributeHomePostalAddress = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "homePostalAddress",
})

// Host is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.14
var AttributeHost = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "host",
})

// Info is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.15
var AttributeInfo = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "info",
})

// Mail is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.16
var AttributeMail = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "mail",
})

// Manager is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.17
var AttributeManager = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "manager",
})

// Mobile is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.18
var AttributeMobile = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "mobile",
})

// Organizational Status is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.19
var AttributeOrganizationalStatus = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "organizationalStatus",
})

// Pager is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.20
var AttributePager = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "pager",
})

// Personal Title is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.21
var AttributePersonalTitle = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "personalTitle",
})

// Room Number is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.22
var AttributeRoomNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "roomNumber",
})

// Secretary is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.23
var AttributeSecretary = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "secretary",
})

// Unique Identifier is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.24
var AttributeUniqueIdentifier = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "uniqueIdentifier",
})

// User Class is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.25
var AttributeUserClass = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "userClass",
})

// AccountExpires is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/acdfe32c-ce53-4073-b9b4-40d1130038dc
var AttributeAccountExpires = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "accountExpires",
})

// CanonicalName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/f613bea0-118a-4807-a4ca-c42c423e2002
var AttributeCanonicalName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "canonicalName",
})

// CarLicense is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/6d5cad6a-5310-438c-9155-8d6973ccb366
var AttributeCarLicense = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "carLicense",
})

// Comment is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/37fa0713-39bf-41ff-98ab-2154a0282c84
var AttributeComment = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "comment",
})

// Company is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/4d3da9d9-8224-4701-ae24-2ed7423b3777
var AttributeCompany = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "company",
})

// CountryCode is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/76cb42fe-2ac6-4a62-9960-eeae986d96a6
var AttributeCountryCode = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "countryCode",
})

// CreateTimeStamp is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/0449fad6-968e-4bbb-bb94-e37c93b88cf9
var AttributeCreateTimeStamp = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "createTimeStamp",
})

// Department is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/8d5d010b-033c-47dc-a413-284e1644c90d
var AttributeDepartment = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "department",
})

// DepartmentNumber is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/b7822883-221f-4c6b-890f-e41f0034cf12
var AttributeDepartmentNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "departmentNumber",
})

// DisplayName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/18d4b197-e223-4119-a55d-396c2fd835d6
var AttributeDisplayName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "displayName",
})

// DisplayNamePrintable is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/393efbf0-c841-4a71-9fdb-1a80d1c565c9
var AttributeDisplayNamePrintable = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "displayNamePrintable",
})

// DistinguishedName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/d66f4eca-f8bb-4cc5-86dc-040aa0ca14ef
var AttributeDistinguishedName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "distinguishedName",
})

// DMDLocation is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/1cccbbb3-dde1-44a5-9c81-66336efed81c
var AttributeDMDLocation = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "dMDLocation",
})

// DmdName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/fa4959ab-7969-4920-9cc9-6cdc6aba6219
var AttributeDmdName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "dmdName",
})

// EmployeeID is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/aca05145-3f3a-48c2-936e-cb64a97d9ae0
var AttributeEmployeeID = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "employeeID",
})

// EmployeeNumber is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/0afba1a7-ff6b-4878-97d0-f099de319dfb
var AttributeEmployeeNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "employeeNumber",
})

// EmployeeType is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/ad00635d-41bf-4551-b0be-def20125669c
var AttributeEmployeeType = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "employeeType",
})

// IconPath is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/3595b632-6b75-48f8-b13b-f610c8bd6f15
var AttributeIconPath = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "iconPath",
})

// IsDeleted is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/8dc0ccda-5cf1-4686-8813-fc48ae2de17f
var AttributeIsDeleted = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "isDeleted",
})

// Location is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/f9e6a136-9746-4bc0-ab8c-1d237f668046
var AttributeLocation = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "location",
})

// MailAddress is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/691e1b43-6e61-4331-9e80-d779ebdcb589
var AttributeMailAddress = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "mailAddress",
})

// MiddleName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/17876454-d2fa-43b5-8df4-df94721fb37f
var AttributeMiddleName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "middleName",
})

// ObjectGUID is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/f5f15ec2-427e-4ebe-bb64-2493cf1d032f
var AttributeObjectGUID = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeUUID,
	Name: "objectGUID",
})

// OtherFacsimileTelephoneNumber is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/f0653a90-9b8d-4e3d-ac31-23fbc853ea95
var AttributeOtherFacsimileTelephoneNumber = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherFacsimileTelephoneNumber",
})

// OtherHomePhone is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/967f8d3a-03f9-4682-8dc0-9961b40b7039
var AttributeOtherHomePhone = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherHomePhone",
})

// OtherIpPhone is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/bc5c8347-2d50-4b8e-a876-667abf60ec27
var AttributeOtherIpPhone = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherIpPhone",
})

// OtherMobile is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/fd93fc9d-ec3f-4dce-a490-dfdc78013ade
var AttributeOtherMobile = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherMobile",
})

// OtherPager is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/add2c35a-f829-4936-a0f7-60b0aff688fd
var AttributeOtherPager = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherPager",
})

// OtherTelephone is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/4cace7cb-8584-49e3-b7a9-1df5cf8470f3
var AttributeOtherTelephone = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "otherTelephone",
})

// UPNSuffixes is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/1f92e6cc-2cb0-45f0-8fbe-83df09cbd8c3
var AttributeUPNSuffixes = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "uPNSuffixes",
})

// Url is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/82a50d8a-e2b2-4291-a6b6-dcbb9586364d
var AttributeUrl = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "url",
})

// UserPrincipalName is from https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adls/63f5e067-d1b3-4e6e-9e53-a92953b6005b
var AttributeUserPrincipalName = DefaultAttributeRegistry.registerKnownAttribute(Attribute{
	Type: AttributeTypeString,
	Name: "userPrincipalName",
})
