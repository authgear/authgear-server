import React, { useContext, useMemo, useCallback } from "react";
import cn from "classnames";
import {
  Dropdown,
  IDropdownOption,
  ComboBox,
  IComboBoxOption,
  DatePicker,
  DayOfWeek,
  FirstWeekOfYear,
  Label,
  Text,
  ITextProps,
  ITheme,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import FormTextField from "../../FormTextField";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { parseBirthdate, toBirthdate } from "../../util/birthdate";
import {
  StandardAttributes,
  StandardAttributesAddress,
  Identity,
  IdentityClaims,
} from "../../types";
import { makeTimezoneOptions } from "../../util/timezone";
import { makeAlpha2Options } from "../../util/alpha2";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsStandardAttributes.module.scss";

export interface UserDetailsStandardAttributesProps {
  identities: Identity[];
  standardAttributes: StandardAttributes;
  onChangeStandardAttributes?: (attrs: StandardAttributes) => void;
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

const UserDetailsStandardAttributes: React.FC<UserDetailsStandardAttributesProps> =
  // eslint-disable-next-line complexity
  function UserDetailsStandardAttributes(
    props: UserDetailsStandardAttributesProps
  ) {
    const { standardAttributes, onChangeStandardAttributes, identities } =
      props;
    const { availableLanguages } = useSystemConfig();
    const { renderToString, locale: appLocale } = useContext(Context);

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

    const onChangeName = useMemo(
      () => makeOnChangeText("name"),
      [makeOnChangeText]
    );

    const onChangeGiveName = useMemo(
      () => makeOnChangeText("given_name"),
      [makeOnChangeText]
    );

    const onChangeFamilyName = useMemo(
      () => makeOnChangeText("family_name"),
      [makeOnChangeText]
    );

    const onChangeMiddleName = useMemo(
      () => makeOnChangeText("middle_name"),
      [makeOnChangeText]
    );

    const onChangeNickname = useMemo(
      () => makeOnChangeText("nickname"),
      [makeOnChangeText]
    );

    const onChangePicture = useMemo(
      () => makeOnChangeText("picture"),
      [makeOnChangeText]
    );

    const onChangeProfile = useMemo(
      () => makeOnChangeText("profile"),
      [makeOnChangeText]
    );

    const onChangeWebsite = useMemo(
      () => makeOnChangeText("website"),
      [makeOnChangeText]
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
        stdAttrKey: keyof StandardAttributes,
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

        if (value != null && typeof value === "string" && !seen.has(value)) {
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

    const gender = standardAttributes.gender;
    const genderOptions: IComboBoxOption[] = useMemo(() => {
      const predefinedOptions: IComboBoxOption[] = [
        { key: "male", text: "male" },
        { key: "female", text: "female" },
      ];
      if (gender != null) {
        const index = predefinedOptions.findIndex((a) => a.key === gender);
        if (index < 0) {
          predefinedOptions.push({
            key: gender,
            text: gender,
            hidden: true,
          });
        }
      }
      return predefinedOptions;
    }, [gender]);
    const onChangeGender = useCallback(
      (
        _e: React.FormEvent<unknown>,
        option?: IComboBoxOption,
        index?: number,
        _value?: string
      ) => {
        if (option != null && index != null && typeof option.key === "string") {
          if (onChangeStandardAttributes != null) {
            onChangeStandardAttributes({
              ...standardAttributes,
              gender: option.key,
            });
          }
        }
      },
      [standardAttributes, onChangeStandardAttributes]
    );
    const onChangeGenderPending = useCallback(
      (option?: IComboBoxOption, index?: number, value?: string) => {
        // We are only interested in the typing case.
        if (option == null && index == null && value != null) {
          if (onChangeStandardAttributes != null) {
            onChangeStandardAttributes({
              ...standardAttributes,
              gender: value,
            });
          }
        }
      },
      [standardAttributes, onChangeStandardAttributes]
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
    const zoneinfoOptions = useMemo(() => {
      return makeTimezoneOptions();
    }, []);
    const onChangeZoneinfo = useMemo(
      () => makeOnChangeDropdown("zoneinfo"),
      [makeOnChangeDropdown]
    );

    const locale = standardAttributes.locale;
    const localeOptions = useMemo(() => {
      let found = false;
      const options: IDropdownOption[] = [];
      for (const tag of availableLanguages) {
        options.push({
          key: tag,
          text: renderToString("Locales." + tag),
        });
        if (locale != null && locale === tag) {
          found = true;
        }
      }

      if (locale != null && !found) {
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

    const alpha2Options = useMemo(() => makeAlpha2Options(), []);
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

    const updatedAt = standardAttributes.updated_at;
    const updatedAtFormatted: string | undefined | null = useMemo(() => {
      if (updatedAt == null) {
        return undefined;
      }

      return formatDatetime(appLocale, new Date(updatedAt * 1000));
    }, [appLocale, updatedAt]);

    return (
      <div className={styles.root}>
        <FormTextField
          className={styles.control}
          value={standardAttributes.name ?? ""}
          onChange={onChangeName}
          parentJSONPointer=""
          fieldName="name"
          label={renderToString("standard-attribute.name")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.given_name ?? ""}
          onChange={onChangeGiveName}
          parentJSONPointer=""
          fieldName="given_name"
          label={renderToString("standard-attribute.given_name")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.family_name ?? ""}
          onChange={onChangeFamilyName}
          parentJSONPointer=""
          fieldName="family_name"
          label={renderToString("standard-attribute.family_name")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.middle_name ?? ""}
          onChange={onChangeMiddleName}
          parentJSONPointer=""
          fieldName="middle_name"
          label={renderToString("standard-attribute.middle_name")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.nickname ?? ""}
          onChange={onChangeNickname}
          parentJSONPointer=""
          fieldName="nickname"
          label={renderToString("standard-attribute.nickname")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.picture ?? ""}
          onChange={onChangePicture}
          parentJSONPointer=""
          fieldName="picture"
          label={renderToString("standard-attribute.picture")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.profile ?? ""}
          onChange={onChangeProfile}
          parentJSONPointer=""
          fieldName="profile"
          label={renderToString("standard-attribute.profile")}
        />
        <FormTextField
          className={styles.control}
          value={standardAttributes.website ?? ""}
          onChange={onChangeWebsite}
          parentJSONPointer=""
          fieldName="website"
          label={renderToString("standard-attribute.website")}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.email")}
          selectedKey={standardAttributes.email}
          options={emailOptions}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.phone_number")}
          selectedKey={standardAttributes.phone_number}
          options={phoneNumberOptions}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.preferred_username")}
          selectedKey={standardAttributes.preferred_username}
          options={preferredUsernameOptions}
        />
        <ComboBox
          className={styles.control}
          label={renderToString("standard-attribute.gender")}
          selectedKey={gender}
          options={genderOptions}
          allowFreeform={true}
          onChange={onChangeGender}
          onPendingValueChanged={onChangeGenderPending}
        />
        <DatePicker
          className={styles.control}
          label={renderToString("standard-attribute.birthdate")}
          firstDayOfWeek={DayOfWeek.Monday}
          firstWeekOfYear={FirstWeekOfYear.FirstFourDayWeek}
          showGoToToday={false}
          allowTextInput={true}
          value={birthdateValue}
          formatDate={formatDate}
          onSelectDate={onSelectBirthdate}
          parseDateFromString={parseDateFromString}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.zoneinfo")}
          selectedKey={zoneinfo}
          options={zoneinfoOptions}
          onChange={onChangeZoneinfo}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.locale")}
          selectedKey={locale}
          options={localeOptions}
          onChange={onChangeLocale}
        />
        <div className={styles.addressGroup}>
          <Label className={styles.addressInput}>
            <Text variant="xLarge">
              <FormattedMessage id="standard-attribute.address" />
            </Text>
          </Label>
          <FormTextField
            className={styles.addressInput}
            value={standardAttributes.address?.street_address ?? ""}
            onChange={onChangeStreetAddress}
            multiline={true}
            parentJSONPointer="/address"
            fieldName="street_address"
            label={renderToString("standard-attribute.street_address")}
          />
          <FormTextField
            className={styles.addressInput}
            value={standardAttributes.address?.locality ?? ""}
            onChange={onChangeLocality}
            parentJSONPointer="/address"
            fieldName="locality"
            label={renderToString("standard-attribute.locality")}
          />
          <div className={styles.addressInputGroup}>
            <FormTextField
              className={cn(styles.addressInput, styles.postalCode)}
              value={standardAttributes.address?.postal_code ?? ""}
              onChange={onChangePostalCode}
              parentJSONPointer="/address"
              fieldName="postal_code"
              label={renderToString("standard-attribute.postal_code")}
            />
            <FormTextField
              className={cn(styles.addressInput, styles.region)}
              value={standardAttributes.address?.region ?? ""}
              onChange={onChangeRegion}
              parentJSONPointer="/address"
              fieldName="region"
              label={renderToString("standard-attribute.region")}
            />
            <Dropdown
              className={cn(styles.addressInput, styles.country)}
              label={renderToString("standard-attribute.country")}
              selectedKey={standardAttributes.address?.country}
              options={alpha2Options}
              onChange={onChangeCountry}
            />
          </div>
        </div>
        {updatedAtFormatted != null && (
          <Text
            className={styles.control}
            variant="small"
            styles={UPDATED_AT_STYLES}
          >
            <FormattedMessage
              id="standard-attribute.updated_at"
              values={{
                datetime: updatedAtFormatted,
              }}
            />
          </Text>
        )}
      </div>
    );
  };

export default UserDetailsStandardAttributes;
