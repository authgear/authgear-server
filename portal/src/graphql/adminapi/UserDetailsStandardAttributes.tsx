import React, { useContext, useMemo } from "react";
import {
  TextField,
  Dropdown,
  IDropdownOption,
  ComboBox,
  IComboBoxOption,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { StandardAttributes } from "../../types";

import styles from "./UserDetailsStandardAttributes.module.scss";

export interface UserDetailsStandardAttributesProps {
  standardAttributes: StandardAttributes;
}

const UserDetailsStandardAttributes: React.FC<UserDetailsStandardAttributesProps> =
  function UserDetailsStandardAttributes(
    props: UserDetailsStandardAttributesProps
  ) {
    const { standardAttributes } = props;
    const { renderToString } = useContext(Context);

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
      </div>
    );
  };

export default UserDetailsStandardAttributes;
