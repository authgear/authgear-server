import React, {
  useContext,
  useMemo,
  useCallback,
  useState,
  Children,
} from "react";
import {
  Dropdown,
  IDropdownOption,
  DatePicker,
  DayOfWeek,
  FirstWeekOfYear,
  Text,
  ITextProps,
  ITheme,
  Label,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import FormTextField from "../../FormTextField";
import FormDropdown from "../../FormDropdown";
import FormPhoneTextField from "../../FormPhoneTextField";
import TextLink from "../../TextLink";
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
import { formatDatetime } from "../../util/formatDatetime";
import { generateLabel } from "../../util/label";
import { checkNumberInput, checkIntegerInput } from "../../util/input";
import {
  jsonPointerToString,
  parseJSONPointerIntoParentChild,
} from "../../util/jsonpointer";
import TextField from "../../TextField";

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

function formatDate(date?: Date): string {
  if (date == null) {
    return "";
  }
  return toBirthdate(date) ?? "";
}

function parseDateFromString(str: string): Date | null {
  return parseBirthdate(str) ?? null;
}

function UPDATED_AT_STYLES(_props: ITextProps, theme: ITheme) {
  return {
    root: {
      color: theme.semanticColors.inputPlaceholderText,
      borderBottom: `1px solid ${theme.palette.neutralTertiaryAlt}`,
      padding: "8px 0",
    },
  };
}

function HorizontalDivider() {
  const { themes } = useSystemConfig();
  const theme = themes.main;
  return (
    <div
      style={{
        backgroundColor: `${theme.palette.neutralTertiaryAlt}`,
        height: "1px",
      }}
    />
  );
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
  // @ts-expect-error
  const value = standardAttributes[fieldName];
  const label = "standard-attribute." + fieldName;
  return (
    <FormTextField
      className={className}
      value={value}
      onChange={onChange}
      parentJSONPointer=""
      fieldName={fieldName}
      label={renderToString(label)}
      placeholder={placeholder}
      disabled={disabled}
    />
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
  // @ts-expect-error
  const value = standardAttributes[fieldName];
  const label = "standard-attribute." + fieldName;
  return (
    <TextLink
      className={className}
      value={value}
      label={renderToString(label)}
    />
  );
}

interface CustomAttributeControlProps {
  attributeConfig: CustomAttributesAttributeConfig;
  customAttributes: CustomAttributesState;
  onChangeCustomAttributes?: (attrs: CustomAttributesState) => void;
}

// eslint-disable-next-line complexity
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
    (_: React.FormEvent<unknown>, option?: IDropdownOption) => {
      if (option == null || onChangeCustomAttributes == null) {
        return;
      }

      const { key } = option;
      if (typeof key === "string") {
        onChangeCustomAttributes({
          ...customAttributes,
          [pointer]: key,
        });
      }
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
        <FormTextField
          className={styles.customAttributeControl}
          value={value}
          onChange={onChange}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "number":
      return (
        <FormTextField
          className={styles.customAttributeControl}
          value={value}
          onChange={onChangeNumber}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "integer":
      return (
        <FormTextField
          className={styles.customAttributeControl}
          value={value}
          onChange={onChangeInteger}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "enum":
      return (
        <FormDropdown
          className={styles.customAttributeControl}
          selectedKey={value}
          onChange={onChangeDropdown}
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
          inputValue={customAttributes["phone_number" + pointer]}
          onChange={onChangePhoneNumber}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "email":
      return (
        <FormTextField
          className={styles.customAttributeControl}
          value={value}
          onChange={onChange}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "url":
      return (
        <FormTextField
          className={styles.customAttributeControl}
          value={value}
          onChange={onChange}
          parentJSONPointer={parent}
          fieldName={fieldName}
          label={label}
          disabled={disabled}
        />
      );
    case "country_code":
      return (
        <FormDropdown
          className={styles.customAttributeControl}
          selectedKey={value}
          onChange={onChangeDropdown}
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
  // eslint-disable-next-line complexity
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
        return (
          _e: React.FormEvent<unknown>,
          option?: IDropdownOption,
          _index?: number
        ) => {
          if (option != null) {
            if (onChangeStandardAttributes != null) {
              onChangeStandardAttributes({
                ...standardAttributes,
                [fieldName]: option.key,
              });
            }
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
        const options = [];
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
      // eslint-disable-next-line complexity
      (
        _e: React.FormEvent<unknown>,
        option?: IDropdownOption,
        _index?: number,
        _value?: string
      ) => {
        if (option != null && typeof option.key === "string") {
          // @ts-expect-error
          setGenderVariant(option.key);
          switch (option.key) {
            case "":
              if (onChangeStandardAttributes != null) {
                onChangeStandardAttributes({
                  ...standardAttributes,
                  gender: "",
                });
              }
              break;
            case "male":
              if (onChangeStandardAttributes != null) {
                onChangeStandardAttributes({
                  ...standardAttributes,
                  gender: "male",
                });
              }
              break;
            case "female":
              if (onChangeStandardAttributes != null) {
                onChangeStandardAttributes({
                  ...standardAttributes,
                  gender: "female",
                });
              }
              break;
            case "other":
              if (onChangeStandardAttributes != null) {
                onChangeStandardAttributes({
                  ...standardAttributes,
                  gender: genderString,
                });
              }
              break;
          }
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
    const birthdateValue = useMemo(() => {
      if (birthdate == null) {
        return undefined;
      }
      const jsDate = parseBirthdate(birthdate);
      return jsDate;
    }, [birthdate]);
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
      (
        _e: React.FormEvent<unknown>,
        option?: IDropdownOption,
        _index?: number
      ) => {
        if (option != null && typeof option.key === "string") {
          if (onChangeStandardAttributes != null) {
            onChangeStandardAttributes({
              ...standardAttributes,
              address: {
                ...standardAttributes.address,
                country: option.key,
              },
            });
          }
        }
      },
      [standardAttributes, onChangeStandardAttributes]
    );

    return (
      <>
        <Label className={styles.standardAttributesTitle}>
          <Text variant="xLarge">
            <FormattedMessage id="UserProfileForm.standard-attributes.title" />
          </Text>
        </Label>
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
          {isReadable("given_name") ? (
            <StandardAttributeTextField
              fieldName="given_name"
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
          {isReadable("family_name") ? (
            <StandardAttributeTextField
              fieldName="family_name"
              standardAttributes={standardAttributes}
              makeOnChangeText={makeOnChangeText}
              isDisabled={isDisabled}
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
          <Dropdown
            label={renderToString("standard-attribute.email")}
            selectedKey={standardAttributes.email}
            onChange={onChangeEmail}
            options={emailOptions}
            disabled={emailOptions.length <= 0}
          />
          <Dropdown
            label={renderToString("standard-attribute.phone_number")}
            selectedKey={standardAttributes.phone_number}
            onChange={onChangePhoneNumber}
            options={phoneNumberOptions}
            disabled={phoneNumberOptions.length <= 0}
          />
          <Dropdown
            label={renderToString("standard-attribute.preferred_username")}
            selectedKey={standardAttributes.preferred_username}
            onChange={onChangePreferredUsername}
            options={preferredUsernameOptions}
            disabled={preferredUsernameOptions.length <= 0}
          />
        </Div>
        {isReadable("gender") ? (
          <Div className={styles.twoColumnGroup}>
            <Dropdown
              label={renderToString("standard-attribute.gender")}
              selectedKey={genderVariant}
              options={genderOptions}
              onChange={onChangeGenderVariant}
              disabled={isDisabled("gender")}
            />
            <TextField
              value={genderVariant === "other" ? genderString : ""}
              onChange={onChangeGenderString}
              disabled={isDisabled("gender") || genderVariant !== "other"}
              label={
                /* Show a non-breaking space so that the label is still rendered */ "\u00a0"
              }
            />
          </Div>
        ) : null}
        {isReadable("birthdate") ? (
          <DatePicker
            label={renderToString("standard-attribute.birthdate")}
            firstDayOfWeek={DayOfWeek.Monday}
            firstWeekOfYear={FirstWeekOfYear.FirstFourDayWeek}
            showGoToToday={false}
            allowTextInput={true}
            value={birthdateValue}
            formatDate={formatDate}
            onSelectDate={onSelectBirthdate}
            parseDateFromString={parseDateFromString}
            placeholder="yyyy-MM-dd"
            disabled={isDisabled("birthdate")}
          />
        ) : null}
        <Div className={styles.twoColumnGroup}>
          {isReadable("zoneinfo") ? (
            <Dropdown
              label={renderToString("standard-attribute.zoneinfo")}
              selectedKey={zoneinfo}
              options={zoneinfoOptions}
              onChange={onChangeZoneinfo}
              disabled={isDisabled("zoneinfo")}
            />
          ) : null}
          {isReadable("locale") ? (
            <Dropdown
              label={renderToString("standard-attribute.locale")}
              selectedKey={locale}
              options={localeOptions}
              onChange={onChangeLocale}
              disabled={isDisabled("locale")}
            />
          ) : null}
        </Div>
        {isReadable("address") ? (
          <Div className={styles.addressGroup}>
            <FormTextField
              className={styles.gridAreaStreet}
              value={standardAttributes.address.street_address}
              onChange={onChangeStreetAddress}
              multiline={true}
              parentJSONPointer="/address"
              fieldName="street_address"
              label={renderToString("standard-attribute.street_address")}
              disabled={isDisabled("address")}
            />
            <FormTextField
              className={styles.gridAreaCity}
              value={standardAttributes.address.locality}
              onChange={onChangeLocality}
              parentJSONPointer="/address"
              fieldName="locality"
              label={renderToString("standard-attribute.locality")}
              disabled={isDisabled("address")}
            />
            <FormTextField
              className={styles.gridAreaPostalCode}
              value={standardAttributes.address.postal_code}
              onChange={onChangePostalCode}
              parentJSONPointer="/address"
              fieldName="postal_code"
              label={renderToString("standard-attribute.postal_code")}
              disabled={isDisabled("address")}
            />
            <FormTextField
              className={styles.gridAreaState}
              value={standardAttributes.address.region}
              onChange={onChangeRegion}
              parentJSONPointer="/address"
              fieldName="region"
              label={renderToString("standard-attribute.region")}
              disabled={isDisabled("address")}
            />
            <Dropdown
              className={styles.gridAreaCountry}
              label={renderToString("standard-attribute.country")}
              selectedKey={standardAttributes.address.country}
              options={alpha2Options}
              onChange={onChangeCountry}
              disabled={isDisabled("address")}
            />
          </Div>
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
      <>
        <Label className={styles.standardAttributesTitle}>
          <Text variant="xLarge">
            <FormattedMessage id="UserProfileForm.custom-attributes.title" />
          </Text>
        </Label>
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
      </>
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
    const { locale: appLocale, renderToString } = useContext(Context);
    const { canSave, onSave } = useFormContainerBaseContext();

    const updatedAt = standardAttributes.updated_at;
    const updatedAtFormatted: string | undefined | null = useMemo(() => {
      if (updatedAt == null) {
        return undefined;
      }

      return formatDatetime(appLocale, new Date(updatedAt * 1000));
    }, [appLocale, updatedAt]);

    return (
      <div className={styles.root}>
        <StandardAttributesForm
          identities={identities}
          standardAttributes={standardAttributes}
          onChangeStandardAttributes={onChangeStandardAttributes}
          standardAttributeAccessControl={standardAttributeAccessControl}
        />
        <HorizontalDivider />
        <CustomAttributesForm
          customAttributes={customAttributes}
          onChangeCustomAttributes={onChangeCustomAttributes}
          customAttributesConfig={customAttributesConfig}
        />
        {updatedAtFormatted != null ? (
          <Text variant="small" styles={UPDATED_AT_STYLES}>
            <FormattedMessage
              id="standard-attribute.updated_at"
              values={{
                datetime: updatedAtFormatted,
              }}
            />
          </Text>
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
