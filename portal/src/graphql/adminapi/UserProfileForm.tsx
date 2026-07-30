import React, {
  useContext,
  useMemo,
  useCallback,
  useState,
  Children,
} from "react";
import type { IDropdownOption } from "@fluentui/react";
import { Select, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import FormPhoneTextField from "../../FormPhoneTextField";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { parseBirthdate, toBirthdate } from "../../util/birthdate";
import {
  StandardAttributes,
  StandardAttributesAddress,
  Identity,
  IdentityClaims,
  AccessControlLevelString,
  CustomAttributesAttributeConfig,
} from "../../types";
import { makeTimezoneOptions } from "../../util/timezone";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { generateLabel } from "../../util/label";
import { checkNumberInput, checkIntegerInput } from "../../util/input";
import {
  jsonPointerToString,
  parseJSONPointerIntoParentChild,
} from "../../util/jsonpointer";
import { TextField } from "../../components/v2/TextField/TextField";
import { FormField } from "../../components/v2/FormField/FormField";

import styles from "./UserProfileForm.module.css";
import PrimaryButton from "../../PrimaryButton";
import { useFormContainerBaseContext } from "../../FormContainerBase";

export interface StandardAttributesAddressState {
  street_address: string;
  locality: string;
  region: string;
  postal_code: string;
  country: string;
}

// We must use string to represent the form state,
// otherwise form dirtyness checking will be incorrect.
export interface StandardAttributesState {
  email: string;
  phone_number: string;
  preferred_username: string;
  family_name: string;
  given_name: string;
  middle_name: string;
  name: string;
  nickname: string;
  picture: string;
  profile: string;
  website: string;
  gender: string;
  birthdate: string | undefined;
  zoneinfo: string;
  locale: string;
  address: StandardAttributesAddressState;
  updated_at?: number;
}

export type CustomAttributesState = Record<string, string>;

export interface UserProfileFormProps {
  identities: Identity[];
  standardAttributes: StandardAttributesState;
  onChangeStandardAttributes?: (attrs: StandardAttributesState) => void;
  standardAttributeAccessControl: Record<string, AccessControlLevelString>;
  customAttributesConfig: CustomAttributesAttributeConfig[];
  customAttributes: CustomAttributesState;
  onChangeCustomAttributes?: (attrs: CustomAttributesState) => void;
}

type GenderVariant = "" | "male" | "female" | "other";

function getInitialGenderVariant(gender: string | undefined): GenderVariant {
  if (gender == null || gender === "") {
    return "";
  }
  if (gender === "male" || gender === "female") {
    return gender;
  }
  return "other";
}

function parseDateFromString(str: string): Date | null {
  return parseBirthdate(str) ?? null;
}


interface DivProps {
  className?: string;
  children?: React.ReactNode;
}

function Div(props: DivProps) {
  const { className, children } = props;
  const array = Children.toArray(children);
  const isEmpty = array.length === 0;
  if (isEmpty) {
    return null;
  }
  return <div className={className}>{children}</div>;
}

const EMPTY_SELECT_VALUE = "__empty__";

interface ProfileSelectFieldProps {
  className?: string;
  label: React.ReactNode;
  value: string;
  options: IDropdownOption[];
  onValueChange: (value: string) => void;
  disabled?: boolean;
  parentJSONPointer?: string;
  fieldName?: string;
}

function ProfileSelectField({
  className,
  label,
  value,
  options,
  onValueChange,
  disabled,
  parentJSONPointer,
  fieldName,
}: ProfileSelectFieldProps): React.ReactElement {
  return (
    <div className={className}>
      <FormField
        size="2"
        labelSize="2"
        label={label}
        labelSpace="1"
        parentJSONPointer={parentJSONPointer}
        fieldName={fieldName}
      >
        <Select.Root
          value={value === "" ? EMPTY_SELECT_VALUE : value}
          onValueChange={(newValue) => {
            onValueChange(
              newValue === EMPTY_SELECT_VALUE ? "" : newValue
            );
          }}
          disabled={disabled}
        >
          <Select.Trigger
            variant="surface"
            className={styles.selectTrigger}
          />
          <Select.Content>
            {options.map((option) => {
              const optionValue = String(option.key);
              return (
                <Select.Item
                  key={optionValue}
                  className={
                    option.hidden ? styles.hiddenSelectItem : undefined
                  }
                  value={
                    optionValue === "" ? EMPTY_SELECT_VALUE : optionValue
                  }
                >
                  {option.text || "\u00a0"}
                </Select.Item>
              );
            })}
          </Select.Content>
        </Select.Root>
      </FormField>
    </div>
  );
}

interface StandardAttributeTextFieldProps {
  standardAttributes: StandardAttributesState;
  fieldName: keyof StandardAttributes;
  makeOnChangeText: (
    fieldName: keyof StandardAttributes
  ) => (e: React.FormEvent<unknown>, v?: string) => void;
  isDisabled: (fieldName: keyof StandardAttributes) => boolean;
  placeholder?: string;
  className?: string;
}

function StandardAttributeTextField(props: StandardAttributeTextFieldProps) {
  const {
    standardAttributes,
    fieldName,
    makeOnChangeText,
    isDisabled,
    placeholder,
    className,
  } = props;
  const { renderToString } = useContext(Context);
  const onChange = useMemo(
    () => makeOnChangeText(fieldName),
    [makeOnChangeText, fieldName]
  );
  const disabled = useMemo(
    () => isDisabled(fieldName),
    [isDisabled, fieldName]
  );
  const value = (standardAttributes as any)[fieldName];
  const label = "standard-attribute." + fieldName;
  return (
    <div className={className}>
      <TextField
        size="2"
        type="text"
        value={value ?? ""}
        onChange={(event) => {
          onChange(event, event.target.value);
        }}
        parentJSONPointer=""
        fieldName={fieldName}
        label={renderToString(label)}
        placeholder={placeholder}
        disabled={disabled}
      />
    </div>
  );
}

interface StandardAttributeLabelProps {
  standardAttributes: StandardAttributesState;
  fieldName: keyof StandardAttributes;
  className?: string;
}

function StandardAttributeLabel(props: StandardAttributeLabelProps) {
  const { standardAttributes, fieldName, className } = props;
  const { renderToString } = useContext(Context);
  const value = (standardAttributes as any)[fieldName];
  const label = "standard-attribute." + fieldName;
  return (
    <div className={className}>
      <TextField
        size="2"
        type="url"
        value={value ?? ""}
        label={renderToString(label)}
        disabled={true}
        readOnly={true}
      />
    </div>
  );
}

interface CustomAttributeControlProps {
  attributeConfig: CustomAttributesAttributeConfig;
  customAttributes: CustomAttributesState;
  onChangeCustomAttributes?: (attrs: CustomAttributesState) => void;
}

function CustomAttributeControl(props: CustomAttributeControlProps) {
  const { attributeConfig, customAttributes, onChangeCustomAttributes } = props;
  const {
    pointer,
    type: typ,
    access_control: { portal_ui: accessControl },
    enum: enu,
  } = attributeConfig;

  const enumOptions: IDropdownOption[] = useMemo(() => {
    const options = [
      {
        key: "",
        text: "",
      },
    ];
    for (const variant of enu ?? []) {
      options.push({
        key: variant,
        text: generateLabel(variant),
      });
    }
    return options;
  }, [enu]);

  const { alpha2Options: o } = useMakeAlpha2Options();
  const alpha2Options = useMemo(() => [{ key: "", text: "" }, ...o], [o]);

  const onChange = useCallback(
    (_: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue == null || onChangeCustomAttributes == null) {
        return;
      }

      onChangeCustomAttributes({
        ...customAttributes,
        [pointer]: newValue,
      });
    },
    [customAttributes, onChangeCustomAttributes, pointer]
  );

  const onChangeNumber = useCallback(
    (_: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue == null || onChangeCustomAttributes == null) {
        return;
      }

      const good = checkNumberInput(newValue);
      if (!good) {
        return;
      }

      onChangeCustomAttributes({
        ...customAttributes,
        [pointer]: newValue,
      });
    },
    [customAttributes, onChangeCustomAttributes, pointer]
  );

  const onChangeInteger = useCallback(
    (_: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue == null || onChangeCustomAttributes == null) {
        return;
      }

      const good = checkIntegerInput(newValue);
      if (!good) {
        return;
      }

      onChangeCustomAttributes({
        ...customAttributes,
        [pointer]: newValue,
      });
    },
    [customAttributes, onChangeCustomAttributes, pointer]
  );

  const onChangeDropdown = useCallback(
    (newValue: string) => {
      if (onChangeCustomAttributes == null) {
        return;
      }
      onChangeCustomAttributes({
        ...customAttributes,
        [pointer]: newValue,
      });
    },
    [customAttributes, onChangeCustomAttributes, pointer]
  );

  const onChangePhoneNumber = useCallback(
    (values: { e164?: string; rawInputValue: string }) => {
      if (onChangeCustomAttributes == null) {
        return;
      }
      const { e164, rawInputValue } = values;

      onChangeCustomAttributes({
        ...customAttributes,
        [pointer]: e164 != null ? e164 : rawInputValue,
        ["phone_number" + pointer]: rawInputValue,
      });
    },
    [customAttributes, onChangeCustomAttributes, pointer]
  );

  const value = customAttributes[pointer];
  const disabled = accessControl === "readonly";

  const parentChild = useMemo(() => {
    return parseJSONPointerIntoParentChild(pointer);
  }, [pointer]);

  const { parent, fieldName, label } = useMemo(() => {
    if (parentChild == null) {
      return {
        parent: "",
        fieldName: "",
        label: "",
      };
    }
    const [parent, fieldName] = parentChild;
    const label = generateLabel(fieldName);
    return {
      parent,
      fieldName,
      label,
    };
  }, [parentChild]);

  if (accessControl !== "readonly" && accessControl !== "readwrite") {
    return null;
  }

  switch (typ) {
    case "string":
      return (
        <div className={styles.customAttributeControl}>
          <TextField
            size="2"
            type="text"
            value={value}
            onChange={(event) => {
              onChange(event, event.target.value);
            }}
            parentJSONPointer={parent}
            fieldName={fieldName}
            label={label}
            disabled={disabled}
          />
        </div>
      );
    case "number":
      return (
        <div className={styles.customAttributeControl}>
          <TextField
            size="2"
            type="text"
            value={value}
            onChange={(event) => {
              onChangeNumber(event, event.target.value);
            }}
            parentJSONPointer={parent}
            fieldName={fieldName}
            label={label}
            disabled={disabled}
          />
        </div>
      );
    case "integer":
      return (
        <div className={styles.customAttributeControl}>
          <TextField
            size="2"
            type="text"
            value={value}
            onChange={(event) => {
              onChangeInteger(event, event.target.value);
            }}
            parentJSONPointer={parent}
            fieldName={fieldName}
            label={label}
            disabled={disabled}
          />
        </div>
      );
    case "enum":
      return (
        <ProfileSelectField
          className={styles.customAttributeControl}
          value={value}
          onValueChange={onChangeDropdown}
          options={enumOptions}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "phone_number":
      return (
        <FormPhoneTextField
          className={styles.customAttributeControl}
          initialInputValue={customAttributes["phone_number" + pointer]}
          onChange={onChangePhoneNumber}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "email":
      return (
        <div className={styles.customAttributeControl}>
          <TextField
            size="2"
            type="email"
            value={value}
            onChange={(event) => {
              onChange(event, event.target.value);
            }}
            parentJSONPointer={parent}
            fieldName={fieldName}
            label={label}
            disabled={disabled}
          />
        </div>
      );
    case "url":
      return (
        <div className={styles.customAttributeControl}>
          <TextField
            size="2"
            type="url"
            value={value}
            onChange={(event) => {
              onChange(event, event.target.value);
            }}
            parentJSONPointer={parent}
            fieldName={fieldName}
            label={label}
            disabled={disabled}
          />
        </div>
      );
    case "country_code":
      return (
        <ProfileSelectField
          className={styles.customAttributeControl}
          value={value}
          onValueChange={onChangeDropdown}
          options={alpha2Options}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
  }
}

interface StandardAttributesFormProps {
  identities: Identity[];
  standardAttributes: StandardAttributesState;
  onChangeStandardAttributes?: (attrs: StandardAttributesState) => void;
  standardAttributeAccessControl: Record<string, AccessControlLevelString>;
}

const StandardAttributesForm: React.VFC<StandardAttributesFormProps> =
  function StandardAttributesForm(props: StandardAttributesFormProps) {
    const {
      standardAttributes,
      onChangeStandardAttributes,
      identities,
      standardAttributeAccessControl,
    } = props;

    const { availableLanguages } = useSystemConfig();
    const { renderToString } = useContext(Context);

    const isReadable = useCallback(
      (fieldName: keyof StandardAttributes) => {
        const ptr = jsonPointerToString([fieldName]);
        const level = standardAttributeAccessControl[ptr];
        return level === "readonly" || level === "readwrite";
      },
      [standardAttributeAccessControl]
    );

    const isDisabled = useCallback(
      (fieldName: keyof StandardAttributes) => {
        const ptr = jsonPointerToString([fieldName]);
        const level = standardAttributeAccessControl[ptr];
        return level !== "readwrite";
      },
      [standardAttributeAccessControl]
    );

    const makeOnChangeText = useCallback(
      (fieldName: keyof StandardAttributes) => {
        return (_e: React.FormEvent<unknown>, newValue?: string) => {
          if (newValue == null || onChangeStandardAttributes == null) {
            return;
          }

          onChangeStandardAttributes({
            ...standardAttributes,
            [fieldName]: newValue,
          });
        };
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    const makeOnChangeAddressText = useCallback(
      (fieldName: keyof StandardAttributesAddress) => {
        return (_e: React.FormEvent<unknown>, newValue?: string) => {
          if (newValue == null || onChangeStandardAttributes == null) {
            return;
          }

          onChangeStandardAttributes({
            ...standardAttributes,
            address: {
              ...standardAttributes.address,
              [fieldName]: newValue,
            },
          });
        };
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    const makeOnChangeDropdown = useCallback(
      (fieldName: keyof StandardAttributes) => {
        return (newValue: string) => {
          if (onChangeStandardAttributes != null) {
            onChangeStandardAttributes({
              ...standardAttributes,
              [fieldName]: newValue,
            });
          }
        };
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    const onChangeEmail = useMemo(
      () => makeOnChangeDropdown("email"),
      [makeOnChangeDropdown]
    );
    const onChangePhoneNumber = useMemo(
      () => makeOnChangeDropdown("phone_number"),
      [makeOnChangeDropdown]
    );
    const onChangePreferredUsername = useMemo(
      () => makeOnChangeDropdown("preferred_username"),
      [makeOnChangeDropdown]
    );

    const onChangeStreetAddress = useMemo(
      () => makeOnChangeAddressText("street_address"),
      [makeOnChangeAddressText]
    );

    const onChangeLocality = useMemo(
      () => makeOnChangeAddressText("locality"),
      [makeOnChangeAddressText]
    );

    const onChangePostalCode = useMemo(
      () => makeOnChangeAddressText("postal_code"),
      [makeOnChangeAddressText]
    );

    const onChangeRegion = useMemo(
      () => makeOnChangeAddressText("region"),
      [makeOnChangeAddressText]
    );

    const makeIdentityDropdownOptions = useCallback(
      (
        stdAttrKey: keyof StandardAttributesState,
        identityClaimKey: keyof IdentityClaims
      ): IDropdownOption[] => {
        const options: IDropdownOption[] = [];
        const value = standardAttributes[stdAttrKey];
        const seen = new Set();

        for (const i of identities) {
          const identityValue = i.claims[identityClaimKey];
          if (
            identityValue != null &&
            typeof identityValue === "string" &&
            !seen.has(identityValue)
          ) {
            seen.add(identityValue);
            options.push({
              key: identityValue,
              text: identityValue,
            });
          }
        }

        if (
          value != null &&
          typeof value === "string" &&
          value !== "" &&
          !seen.has(value)
        ) {
          options.push({
            key: value,
            text: value,
            hidden: true,
          });
        }

        return options;
      },
      [identities, standardAttributes]
    );

    const emailOptions = useMemo(
      () => makeIdentityDropdownOptions("email", "email"),
      [makeIdentityDropdownOptions]
    );

    const phoneNumberOptions = useMemo(
      () => makeIdentityDropdownOptions("phone_number", "phone_number"),
      [makeIdentityDropdownOptions]
    );

    const preferredUsernameOptions = useMemo(
      () =>
        makeIdentityDropdownOptions("preferred_username", "preferred_username"),
      [makeIdentityDropdownOptions]
    );

    const [genderVariant, setGenderVariant] = useState<GenderVariant>(
      getInitialGenderVariant(standardAttributes.gender)
    );
    const [genderString, setGenderString] = useState<string>(
      standardAttributes.gender
    );
    const genderOptions: IDropdownOption[] = useMemo(() => {
      const options: IDropdownOption[] = [
        { key: "", text: "" },
        { key: "male", text: "male" },
        { key: "female", text: "female" },
        {
          key: "other",
          text: renderToString("UserProfileForm.gender.other.label"),
        },
      ];
      return options;
    }, [renderToString]);
    const onChangeGenderVariant = useCallback(
      (newValue: string) => {
        const variant = newValue as GenderVariant;
        setGenderVariant(variant);
        if (onChangeStandardAttributes != null) {
          onChangeStandardAttributes({
            ...standardAttributes,
            gender: variant === "other" ? genderString : variant,
          });
        }
      },
      [standardAttributes, onChangeStandardAttributes, genderString]
    );
    const onChangeGenderString = useCallback(
      (_e: React.FormEvent<unknown>, newValue?: string) => {
        if (newValue != null) {
          setGenderString(newValue);
          if (genderVariant === "other") {
            if (onChangeStandardAttributes != null) {
              onChangeStandardAttributes({
                ...standardAttributes,
                gender: newValue,
              });
            }
          }
        }
      },
      [genderVariant, onChangeStandardAttributes, standardAttributes]
    );

    const birthdate = standardAttributes.birthdate;
    const onSelectBirthdate = useCallback(
      (date: Date | null | undefined) => {
        if (onChangeStandardAttributes == null) {
          return;
        }

        if (date == null || isNaN(date.getTime())) {
          onChangeStandardAttributes({
            ...standardAttributes,
            birthdate: undefined,
          });
        } else {
          onChangeStandardAttributes({
            ...standardAttributes,
            birthdate: toBirthdate(date),
          });
        }
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    const zoneinfo = standardAttributes.zoneinfo;
    const zoneinfoOptions = useMemo(
      () => [{ key: "", text: "" }, ...makeTimezoneOptions()],
      []
    );
    const onChangeZoneinfo = useMemo(
      () => makeOnChangeDropdown("zoneinfo"),
      [makeOnChangeDropdown]
    );

    const locale = standardAttributes.locale;
    const localeOptions = useMemo(() => {
      let found = false;
      const options: IDropdownOption[] = [
        {
          key: "",
          text: "",
        },
      ];
      for (const tag of availableLanguages) {
        options.push({
          key: tag,
          text: renderToString("Locales." + tag),
        });
        if (locale === tag) {
          found = true;
        }
      }

      if (!found) {
        options.push({
          key: locale,
          text: locale,
          hidden: true,
        });
      }

      return options;
    }, [locale, renderToString, availableLanguages]);
    const onChangeLocale = useMemo(
      () => makeOnChangeDropdown("locale"),
      [makeOnChangeDropdown]
    );

    const { alpha2Options: o } = useMakeAlpha2Options();
    const alpha2Options = useMemo(() => [{ key: "", text: "" }, ...o], [o]);

    const onChangeCountry = useCallback(
      (newValue: string) => {
        if (onChangeStandardAttributes != null) {
          onChangeStandardAttributes({
            ...standardAttributes,
            address: {
              ...standardAttributes.address,
              country: newValue,
            },
          });
        }
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    return (
      <>
        {/* Personal Information card */}
        <div className={styles.sectionCard}>
          <div className={styles.sectionHeader}>
            <Text as="p" size="3" weight="medium">
              <FormattedMessage id="UserProfileForm.personal-information.title" />
            </Text>
          </div>
          <div className={styles.sectionFields}>
            <Div className={styles.twoColumnGroup}>
              {isReadable("given_name") ? (
                <StandardAttributeTextField
                  fieldName="given_name"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                />
              ) : null}
              {isReadable("family_name") ? (
                <StandardAttributeTextField
                  fieldName="family_name"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                />
              ) : null}
            </Div>
            <Div className={styles.nameGroup}>
              {isReadable("name") ? (
                <StandardAttributeTextField
                  fieldName="name"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                />
              ) : null}
              {isReadable("nickname") ? (
                <StandardAttributeTextField
                  fieldName="nickname"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                />
              ) : null}
              {isReadable("middle_name") ? (
                <StandardAttributeTextField
                  fieldName="middle_name"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                />
              ) : null}
            </Div>
            <Div className={styles.twoColumnGroup}>
              {isReadable("birthdate") ? (
                <TextField
                  size="2"
                  type="date"
                  inputClassName={styles.birthdateInput}
                  label={renderToString("standard-attribute.birthdate")}
                  value={birthdate ?? ""}
                  onChange={(event) => {
                    onSelectBirthdate(parseDateFromString(event.target.value));
                  }}
                  disabled={isDisabled("birthdate")}
                />
              ) : null}
              {isReadable("gender") ? (
                <ProfileSelectField
                  label={renderToString("standard-attribute.gender")}
                  value={genderVariant}
                  options={genderOptions}
                  onValueChange={onChangeGenderVariant}
                  disabled={isDisabled("gender")}
                />
              ) : null}
            </Div>
            {isReadable("gender") && genderVariant === "other" ? (
              <TextField
                size="2"
                type="text"
                value={genderString}
                onChange={(event) => {
                  onChangeGenderString(event, event.target.value);
                }}
                disabled={isDisabled("gender")}
                label={renderToString("standard-attribute.gender") + " (custom)"}
              />
            ) : null}
            <Div className={styles.twoColumnGroup}>
              {isReadable("zoneinfo") ? (
                <ProfileSelectField
                  label={renderToString("standard-attribute.zoneinfo")}
                  value={zoneinfo}
                  options={zoneinfoOptions}
                  onValueChange={onChangeZoneinfo}
                  disabled={isDisabled("zoneinfo")}
                />
              ) : null}
              {isReadable("locale") ? (
                <ProfileSelectField
                  label={renderToString("standard-attribute.locale")}
                  value={locale}
                  options={localeOptions}
                  onValueChange={onChangeLocale}
                  disabled={isDisabled("locale")}
                />
              ) : null}
            </Div>
            {isReadable("picture") ? (
              <StandardAttributeLabel
                fieldName="picture"
                standardAttributes={standardAttributes}
              />
            ) : null}
            <Div className={styles.singleColumnGroup}>
              {isReadable("profile") ? (
                <StandardAttributeTextField
                  fieldName="profile"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                  placeholder={renderToString(
                    "UserProfileForm.profile.placeholder"
                  )}
                />
              ) : null}
              {isReadable("website") ? (
                <StandardAttributeTextField
                  fieldName="website"
                  standardAttributes={standardAttributes}
                  makeOnChangeText={makeOnChangeText}
                  isDisabled={isDisabled}
                  placeholder={renderToString(
                    "UserProfileForm.website.placeholder"
                  )}
                />
              ) : null}
            </Div>
            <Div className={styles.twoColumnGroup}>
              <ProfileSelectField
                label={renderToString("standard-attribute.email")}
                value={standardAttributes.email}
                onValueChange={onChangeEmail}
                options={emailOptions}
                disabled={emailOptions.length <= 0}
              />
              <ProfileSelectField
                label={renderToString("standard-attribute.phone_number")}
                value={standardAttributes.phone_number}
                onValueChange={onChangePhoneNumber}
                options={phoneNumberOptions}
                disabled={phoneNumberOptions.length <= 0}
              />
            </Div>
            <Div className={styles.singleColumnGroup}>
              <ProfileSelectField
                label={renderToString("standard-attribute.preferred_username")}
                value={standardAttributes.preferred_username}
                onValueChange={onChangePreferredUsername}
                options={preferredUsernameOptions}
                disabled={preferredUsernameOptions.length <= 0}
              />
            </Div>
          </div>
        </div>

        {/* Contact & Address card */}
        {isReadable("address") ? (
          <div className={styles.sectionCard}>
            <div className={styles.sectionHeader}>
              <Text as="p" size="3" weight="medium">
                <FormattedMessage id="UserProfileForm.contact-address.title" />
              </Text>
            </div>
            <div className={styles.sectionFields}>
              <Div className={styles.addressGroup}>
                <TextField
                  size="2"
                  type="text"
                  inputClassName={styles.gridAreaStreet}
                  value={standardAttributes.address.street_address}
                  onChange={(event) => {
                    onChangeStreetAddress(event, event.target.value);
                  }}
                  parentJSONPointer="/address"
                  fieldName="street_address"
                  label={renderToString("standard-attribute.street_address")}
                  disabled={isDisabled("address")}
                />
                <Div className={styles.twoColumnGroupInline}>
                  <TextField
                    size="2"
                    type="text"
                    value={standardAttributes.address.locality}
                    onChange={(event) => {
                      onChangeLocality(event, event.target.value);
                    }}
                    parentJSONPointer="/address"
                    fieldName="locality"
                    label={renderToString("standard-attribute.locality")}
                    disabled={isDisabled("address")}
                  />
                  <TextField
                    size="2"
                    type="text"
                    value={standardAttributes.address.region}
                    onChange={(event) => {
                      onChangeRegion(event, event.target.value);
                    }}
                    parentJSONPointer="/address"
                    fieldName="region"
                    label={renderToString("standard-attribute.region")}
                    disabled={isDisabled("address")}
                  />
                </Div>
                <Div className={styles.twoColumnGroupInline}>
                  <TextField
                    size="2"
                    type="text"
                    value={standardAttributes.address.postal_code}
                    onChange={(event) => {
                      onChangePostalCode(event, event.target.value);
                    }}
                    parentJSONPointer="/address"
                    fieldName="postal_code"
                    label={renderToString("standard-attribute.postal_code")}
                    disabled={isDisabled("address")}
                  />
                  <ProfileSelectField
                    label={renderToString("standard-attribute.country")}
                    value={standardAttributes.address.country}
                    options={alpha2Options}
                    onValueChange={onChangeCountry}
                    disabled={isDisabled("address")}
                  />
                </Div>
              </Div>
            </div>
          </div>
        ) : null}
      </>
    );
  };

interface CustomAttributesFormProps {
  customAttributes: CustomAttributesState;
  onChangeCustomAttributes?: (attrs: CustomAttributesState) => void;
  customAttributesConfig: CustomAttributesAttributeConfig[];
}

const CustomAttributesForm: React.VFC<CustomAttributesFormProps> =
  function CustomAttributesForm(props: CustomAttributesFormProps) {
    const {
      customAttributes,
      onChangeCustomAttributes,
      customAttributesConfig,
    } = props;

    return (
      <div className={styles.sectionCard}>
        <div className={styles.sectionHeader}>
          <Text as="p" size="3" weight="medium">
            <FormattedMessage id="UserProfileForm.custom-attributes.title" />
          </Text>
        </div>
        <div className={styles.sectionFields}>
          <div className={styles.customAttributesForm}>
            {customAttributesConfig.map((c) => {
              return (
                <CustomAttributeControl
                  key={c.id}
                  attributeConfig={c}
                  customAttributes={customAttributes}
                  onChangeCustomAttributes={onChangeCustomAttributes}
                />
              );
            })}
          </div>
        </div>
      </div>
    );
  };

const UserProfileForm: React.VFC<UserProfileFormProps> =
  function UserProfileForm(props: UserProfileFormProps) {
    const {
      identities,
      standardAttributes,
      onChangeStandardAttributes,
      standardAttributeAccessControl,
      customAttributes,
      onChangeCustomAttributes,
      customAttributesConfig,
    } = props;
    const { renderToString } = useContext(Context);
    const { canSave, onSave } = useFormContainerBaseContext();

    return (
      <div className={styles.root}>
        <StandardAttributesForm
          identities={identities}
          standardAttributes={standardAttributes}
          onChangeStandardAttributes={onChangeStandardAttributes}
          standardAttributeAccessControl={standardAttributeAccessControl}
        />
        {customAttributesConfig.length > 0 ? (
          <CustomAttributesForm
            customAttributes={customAttributes}
            onChangeCustomAttributes={onChangeCustomAttributes}
            customAttributesConfig={customAttributesConfig}
          />
        ) : null}
        <div>
          <PrimaryButton
            text={renderToString("save")}
            disabled={!canSave}
            onClick={onSave}
          />
        </div>
      </div>
    );
  };

export default UserProfileForm;
