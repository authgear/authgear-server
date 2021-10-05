import React, { useContext } from "react";
import { TextField } from "@fluentui/react";
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
      </div>
    );
  };

export default UserDetailsStandardAttributes;
