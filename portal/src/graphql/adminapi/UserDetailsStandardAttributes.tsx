import React, { useContext, useMemo } from "react";
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
import { StandardAttributes } from "../../types";
import { makeTimezoneOptions } from "../../util/timezone";
import { makeAlpha2Options } from "../../util/alpha2";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsStandardAttributes.module.scss";

export interface UserDetailsStandardAttributesProps {
  standardAttributes: StandardAttributes;
}

function formatDate(date?: Date): string {
  if (date == null) {
    return "";
  }
  return toBirthdate(date) ?? "";
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
    const { standardAttributes } = props;
    const { availableLanguages } = useSystemConfig();
    const { renderToString, locale: appLocale } = useContext(Context);

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

    const birthdate = standardAttributes.birthdate;
    const birthdateValue = useMemo(() => {
      if (birthdate == null) {
        return undefined;
      }
      const jsDate = parseBirthdate(birthdate);
      return jsDate;
    }, [birthdate]);

    const zoneinfo = standardAttributes.zoneinfo;
    const zoneinfoOptions = useMemo(() => {
      return makeTimezoneOptions();
    }, []);

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

    const alpha2Options = useMemo(() => makeAlpha2Options(), []);

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
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.given_name")}
          value={standardAttributes.given_name ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.family_name")}
          value={standardAttributes.family_name ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.middle_name")}
          value={standardAttributes.middle_name ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.nickname")}
          value={standardAttributes.nickname ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.picture")}
          value={standardAttributes.picture ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.profile")}
          value={standardAttributes.profile ?? ""}
        />
        <TextField
          className={styles.control}
          label={renderToString("standard-attribute.website")}
          value={standardAttributes.website ?? ""}
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
        />
        <DatePicker
          className={styles.control}
          label={renderToString("standard-attribute.birthdate")}
          firstDayOfWeek={DayOfWeek.Monday}
          firstWeekOfYear={FirstWeekOfYear.FirstFourDayWeek}
          value={birthdateValue}
          formatDate={formatDate}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.zoneinfo")}
          selectedKey={zoneinfo}
          options={zoneinfoOptions}
        />
        <Dropdown
          className={styles.control}
          label={renderToString("standard-attribute.locale")}
          selectedKey={locale}
          options={localeOptions}
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
          />
          <TextField
            className={styles.addressInput}
            label={renderToString("standard-attribute.locality")}
            value={standardAttributes.address?.locality}
          />
          <div className={styles.addressInputGroup}>
            <TextField
              className={cn(styles.addressInput, styles.postalCode)}
              label={renderToString("standard-attribute.postal_code")}
              value={standardAttributes.address?.postal_code}
            />
            <TextField
              className={cn(styles.addressInput, styles.region)}
              label={renderToString("standard-attribute.region")}
              value={standardAttributes.address?.region}
            />
            <Dropdown
              className={cn(styles.addressInput, styles.country)}
              label={renderToString("standard-attribute.country")}
              selectedKey={standardAttributes.address?.country}
              options={alpha2Options}
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
