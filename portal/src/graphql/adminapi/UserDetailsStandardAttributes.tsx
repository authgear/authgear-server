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
  Label,
  Text,
  ITextProps,
  ITheme,
  TextField,
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

export interface UserDetailsStandardAttributesProps {
  identities: Identity[];
  standardAttributes: StandardAttributesState;
  onChangeStandardAttributes?: (attrs: StandardAttributesState) => void;
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
          text: renderToString(
            "UserDetailsStandardAttributes.gender.other.label"
          ),
        },
      ];
      return options;
    }, [renderToString]);
    const onChangeGenderVariant = useCallback(
      // eslint-disable-next-line
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

    const alpha2Options = useMemo(
      () => [{ key: "", text: "" }, ...makeAlpha2Options()],
      []
    );
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
        <Div className={styles.nameGroup}>
          <FormTextField
            className={styles.c1}
            value={standardAttributes.name}
            onChange={onChangeName}
            parentJSONPointer=""
            fieldName="name"
            label={renderToString("standard-attribute.name")}
          />
          <FormTextField
            className={styles.c2}
            value={standardAttributes.nickname}
            onChange={onChangeNickname}
            parentJSONPointer=""
            fieldName="nickname"
            label={renderToString("standard-attribute.nickname")}
          />
          <FormTextField
            className={styles.c3}
            value={standardAttributes.given_name}
            onChange={onChangeGiveName}
            parentJSONPointer=""
            fieldName="given_name"
            label={renderToString("standard-attribute.given_name")}
          />
          <FormTextField
            className={styles.c4}
            value={standardAttributes.middle_name}
            onChange={onChangeMiddleName}
            parentJSONPointer=""
            fieldName="middle_name"
            label={renderToString("standard-attribute.middle_name")}
          />
          <FormTextField
            className={styles.c5}
            value={standardAttributes.family_name}
            onChange={onChangeFamilyName}
            parentJSONPointer=""
            fieldName="family_name"
            label={renderToString("standard-attribute.family_name")}
          />
        </Div>
        <FormTextField
          className={styles.standalone}
          value={standardAttributes.picture}
          onChange={onChangePicture}
          parentJSONPointer=""
          fieldName="picture"
          label={renderToString("standard-attribute.picture")}
          placeholder={renderToString(
            "UserDetailsStandardAttributes.picture.placeholder"
          )}
        />
        <Div className={styles.singleColumnGroup}>
          <FormTextField
            value={standardAttributes.profile}
            onChange={onChangeProfile}
            parentJSONPointer=""
            fieldName="profile"
            label={renderToString("standard-attribute.profile")}
            placeholder={renderToString(
              "UserDetailsStandardAttributes.profile.placeholder"
            )}
          />
          <FormTextField
            value={standardAttributes.website}
            onChange={onChangeWebsite}
            parentJSONPointer=""
            fieldName="website"
            label={renderToString("standard-attribute.website")}
            placeholder={renderToString(
              "UserDetailsStandardAttributes.website.placeholder"
            )}
          />
        </Div>
        <Div className={styles.twoColumnGroup}>
          <Dropdown
            label={renderToString("standard-attribute.email")}
            selectedKey={standardAttributes.email}
            options={emailOptions}
          />
          <Dropdown
            label={renderToString("standard-attribute.phone_number")}
            selectedKey={standardAttributes.phone_number}
            options={phoneNumberOptions}
          />
          <Dropdown
            label={renderToString("standard-attribute.preferred_username")}
            selectedKey={standardAttributes.preferred_username}
            options={preferredUsernameOptions}
          />
        </Div>
        <Div className={styles.twoColumnGroup}>
          <Dropdown
            label={renderToString("standard-attribute.gender")}
            selectedKey={genderVariant}
            options={genderOptions}
            onChange={onChangeGenderVariant}
          />
          <TextField
            value={genderVariant === "other" ? genderString : ""}
            onChange={onChangeGenderString}
            disabled={genderVariant !== "other"}
            label={
              /* Show a non-breaking space so that the label is still rendered */ "\u00a0"
            }
          />
        </Div>
        <DatePicker
          className={styles.standalone}
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
        />
        <Div className={styles.twoColumnGroup}>
          <Dropdown
            label={renderToString("standard-attribute.zoneinfo")}
            selectedKey={zoneinfo}
            options={zoneinfoOptions}
            onChange={onChangeZoneinfo}
          />
          <Dropdown
            label={renderToString("standard-attribute.locale")}
            selectedKey={locale}
            options={localeOptions}
            onChange={onChangeLocale}
          />
        </Div>
        <Div className={styles.addressGroup}>
          <Label className={styles.gridAreaLabel}>
            <Text variant="xLarge">
              <FormattedMessage id="standard-attribute.address" />
            </Text>
          </Label>
          <FormTextField
            className={styles.gridAreaStreet}
            value={standardAttributes.address.street_address}
            onChange={onChangeStreetAddress}
            multiline={true}
            parentJSONPointer="/address"
            fieldName="street_address"
            label={renderToString("standard-attribute.street_address")}
          />
          <FormTextField
            className={styles.gridAreaCity}
            value={standardAttributes.address.locality}
            onChange={onChangeLocality}
            parentJSONPointer="/address"
            fieldName="locality"
            label={renderToString("standard-attribute.locality")}
          />
          <FormTextField
            className={styles.gridAreaPostalCode}
            value={standardAttributes.address.postal_code}
            onChange={onChangePostalCode}
            parentJSONPointer="/address"
            fieldName="postal_code"
            label={renderToString("standard-attribute.postal_code")}
          />
          <FormTextField
            className={styles.gridAreaState}
            value={standardAttributes.address.region}
            onChange={onChangeRegion}
            parentJSONPointer="/address"
            fieldName="region"
            label={renderToString("standard-attribute.region")}
          />
          <Dropdown
            className={styles.gridAreaCountry}
            label={renderToString("standard-attribute.country")}
            selectedKey={standardAttributes.address.country}
            options={alpha2Options}
            onChange={onChangeCountry}
          />
        </Div>
        {updatedAtFormatted != null && (
          <Text
            className={styles.standalone}
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
