import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  IContextualMenuProps,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";

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
    <ListCellLayout className={styles.cellContainer}>
      <div className={styles.cellIcon}>{item.icon}</div>
      <Text className={styles.cellName}>{item.identityName}</Text>
      <Text className={styles.cellDesc}>
        <FormattedMessage
          id="UserDetails.connected-identities.connected-on"
          values={{ datetime: item.connectedOn }}
        />
      </Text>
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
  const { locale, renderToString } = useContext(Context);
  const navigate = useNavigate();
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

  const addIdentitiesMenuProps: IContextualMenuProps = useMemo(
    () => ({
      items: [
        {
          key: "email",
          text: renderToString("UserDetails.connected-identities.email"),
          iconProps: { iconName: "Mail" },
          onClick: () => navigate("./add-email"),
        },
        {
          key: "phone",
          text: renderToString("UserDetails.connected-identities.phone"),
          iconProps: { iconName: "CellPhone" },
          onClick: () => navigate("./add-phone"),
        },
        {
          key: "username",
          text: renderToString("UserDetails.connected-identities.username"),
          iconProps: { iconName: "Accounts" },
          onClick: () => navigate("./add-username"),
        },
      ],
      directionalHintFixed: true,
    }),
    [renderToString, navigate]
  );

  return (
    <div className={styles.root}>
      <section className={styles.headerSection}>
        <Text as="h2" className={styles.header}>
          <FormattedMessage id="UserDetails.connected-identities.title" />
        </Text>
        <PrimaryButton
          iconProps={{ iconName: "CirclePlus" }}
          menuProps={addIdentitiesMenuProps}
          styles={{
            menuIcon: { paddingLeft: "3px" },
            icon: { paddingRight: "3px" },
          }}
        >
          <FormattedMessage id="UserDetails.connected-identities.add-identity" />
        </PrimaryButton>
      </section>
      <List
        className={styles.list}
        items={identityListItems}
        onRenderCell={onRenderIdentityCell}
      />
    </div>
  );
};

export default UserDetailsConnectedIdentities;
