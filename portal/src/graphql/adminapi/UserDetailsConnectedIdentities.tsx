import React, { useMemo, useCallback } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { List, PrimaryButton } from "@fluentui/react";

import ListCellLayout from "../../ListCellLayout";
import { formatDatetime } from "../../util/formatDatetime";
import { destructiveTheme } from "../../theme";

import styles from "./UserDetailsConnectedIdentities.module.scss";

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
}

interface Identity {
  id: string;
  type: string;
  claims: IdentityClaim;
  createdAt: string;
  updatedAt: string;
}

interface UserDetailsConnectedIdentitiesProps {
  identities: Identity[];
}

interface IdentityListItem {
  icon: React.ReactNode;
  identityName: string;
  connectedOn: string;
  onRemoveClicked?: () => void;
}

const IdentityListCell: React.FC<{
  item?: IdentityListItem;
}> = function IdentityListCell(props: { item?: IdentityListItem }) {
  const { item } = props;
  if (item == null) {
    return null;
  }
  return (
    <ListCellLayout>
      <div className={styles.cellInfo}>
        <div className={styles.cellIcon}>{item.icon}</div>
        <div className={styles.cellDetail}>
          <div className={styles.cellName}>{item.identityName}</div>
          <div className={styles.cellDesc}>
            <FormattedMessage
              id="UserDetails.connected-identities.connected-on"
              values={{ datetime: item.connectedOn }}
            />
          </div>
        </div>
      </div>
      <PrimaryButton
        className={styles.removeButton}
        theme={destructiveTheme}
        onClick={item.onRemoveClicked}
      >
        <FormattedMessage id="remove" />
      </PrimaryButton>
    </ListCellLayout>
  );
};

const UserDetailsConnectedIdentities: React.FC<UserDetailsConnectedIdentitiesProps> = function UserDetailsConnectedIdentities(
  props: UserDetailsConnectedIdentitiesProps
) {
  const { identities } = props;
  const { locale } = React.useContext(Context);
  const identityListItems: IdentityListItem[] = useMemo(() => {
    return identities.map((identity) => {
      const identityName = identity.claims.email ?? "---";
      const connectedOn = formatDatetime(locale, identity.createdAt) ?? "";
      const icon = (
        <div
          style={{ width: "20px", height: "20px", backgroundColor: "grey" }}
        ></div>
      );
      return {
        icon,
        identityName,
        connectedOn,
      };
    });
  }, [locale, identities]);

  const onRenderIdentityCell = useCallback(
    (item?: IdentityListItem, _index?: number): React.ReactNode => {
      return <IdentityListCell item={item} />;
    },
    []
  );

  return (
    <div className={styles.root}>
      <h3 className={styles.header}>
        <FormattedMessage id="UserDetails.connected-identities.header" />
      </h3>
      <List items={identityListItems} onRenderCell={onRenderIdentityCell} />
    </div>
  );
};

export default UserDetailsConnectedIdentities;
