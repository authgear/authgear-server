import React, { useContext, useMemo, useCallback } from "react";
import cn from "classnames";
import {
  TextField,
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
import { useSystemConfig } from "../../context/SystemConfigContext";
import { parseBirthdate, toBirthdate } from "../../util/birthdate";
import { StandardAttributes, StandardAttributesAddress } from "../../types";
import { makeTimezoneOptions } from "../../util/timezone";
import { makeAlpha2Options } from "../../util/alpha2";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsStandardAttributes.module.scss";

export interface UserDetailsStandardAttributesProps {
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
  function UserDetailsStandardAttributes(
    props: UserDetailsStandardAttributesProps
  ) {
    const { standardAttributes, onChangeStandardAttributes } = props;
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

    const email = standardAttributes.email;
    const emailOptions: IDropdownOption[] = useMemo(() => {
      if (email == null) {
        return [];
      }
      return [{ key: email, text: email }];
    }, [email]);

    const phoneNumber = standardAttributes.phone_number;
    const phoneNumberOptions: IDropdownOption[] = useMemo(() => {
      if (phoneNumber == null) {
        return [];
      }
      return [{ key: phoneNumber, text: phoneNumber }];
    }, [phoneNumber]);

    const preferredUsername = standardAttributes.preferred_username;
    const preferredUsernameOptions: IDropdownOption[] = useMemo(() => {
      if (preferredUsername == null) {
        return [];
      }
      return [{ key: preferredUsername, text: preferredUsername }];
    }, [preferredUsername]);

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
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.name")}
          value={standardAttributes.name ?? ""}
          onChange={onChangeName}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.given_name")}
          value={standardAttributes.given_name ?? ""}
          onChange={onChangeGiveName}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.family_name")}
          value={standardAttributes.family_name ?? ""}
          onChange={onChangeFamilyName}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.middle_name")}
          value={standardAttributes.middle_name ?? ""}
          onChange={onChangeMiddleName}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.nickname")}
          value={standardAttributes.nickname ?? ""}
          onChange={onChangeNickname}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.picture")}
          value={standardAttributes.picture ?? ""}
          onChange={onChangePicture}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.profile")}
          value={standardAttributes.profile ?? ""}
          onChange={onChangeProfile}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.website")}
          value={standardAttributes.website ?? ""}
          onChange={onChangeWebsite}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.email")}
          selectedKey={email}
          options={emailOptions}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.phone_number")}
          selectedKey={phoneNumber}
          options={phoneNumberOptions}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.preferred_username")}
          selectedKey={preferredUsername}
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
          <TextField
            className={styles.addressInput}
            label={renderToString("standard-attribute.street_address")}
            multiline={true}
            value={standardAttributes.address?.street_address}
            onChange={onChangeStreetAddress}
          />
          <TextField
            className={styles.addressInput}
            label={renderToString("standard-attribute.locality")}
            value={standardAttributes.address?.locality}
            onChange={onChangeLocality}
          />
          <div className={styles.addressInputGroup}>
            <TextField
              className={cn(styles.addressInput, styles.postalCode)}
              label={renderToString("standard-attribute.postal_code")}
              value={standardAttributes.address?.postal_code}
              onChange={onChangePostalCode}
            />
            <TextField
              className={cn(styles.addressInput, styles.region)}
              label={renderToString("standard-attribute.region")}
              value={standardAttributes.address?.region}
              onChange={onChangeRegion}
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
