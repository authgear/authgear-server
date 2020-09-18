import React, { useMemo, useCallback, useContext } from "react";
import cn from "classnames";
import { useNavigate } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  DefaultButton,
  Icon,
  IContextualMenuProps,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";

import ListCellLayout from "../../ListCellLayout";
import { formatDatetime } from "../../util/formatDatetime";
import { destructiveTheme, verifyButtonTheme } from "../../theme";

import styles from "./UserDetailsConnectedIdentities.module.scss";

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
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

const identityTypes = ["email", "phone", "username"] as const;
type IdentityType = typeof identityTypes[number];

interface EmailIdentityListItem {
  email: string;
  verified: boolean;
  connectedOn: string;
}

interface PhoneIdentityListItem {
  phone: string;
  verified: boolean;
  addedOn: string;
}

interface UsernameIdentityListItem {
  username: string;
  addedOn: string;
}

interface IdentityLists {
  email: EmailIdentityListItem[];
  phone: PhoneIdentityListItem[];
  username: UsernameIdentityListItem[];
}

interface IdentityListCellProps {
  identityType: IdentityType;
  identityName: string;
  addedOn?: string;
  connectedOn?: string;
  verified?: boolean;
  toggleVerified?: (verified: boolean) => void;
  remove?: () => void;
}

interface VerifyButtonProps {
  verified: boolean;
  toggleVerified: (verified: boolean) => void;
}

const iconMap: Record<IdentityType, React.ReactNode> = {
  email: <Icon iconName="Mail" />,
  phone: <Icon iconName="CellPhone" />,
  username: <Icon iconName="Accounts" />,
};

const removeButtonTextId: Record<IdentityType, string> = {
  email: "disconnect",
  phone: "remove",
  username: "remove",
};

const VerifyButton: React.FC<VerifyButtonProps> = function VerifyButton(
  props: VerifyButtonProps
) {
  const { verified, toggleVerified } = props;

  const onClickVerify = useCallback(() => {
    toggleVerified(true);
  }, [toggleVerified]);

  const onClickUnverify = useCallback(() => {
    toggleVerified(false);
  }, [toggleVerified]);

  if (verified) {
    return (
      <DefaultButton
        className={cn(styles.controlButton, styles.unverifyButton)}
        onClick={onClickUnverify}
      >
        <FormattedMessage id={"unverify"} />
      </DefaultButton>
    );
  }

  return (
    <PrimaryButton
      className={cn(styles.controlButton, styles.verifyButton)}
      theme={verifyButtonTheme}
      onClick={onClickVerify}
    >
      <FormattedMessage id={"verify"} />
    </PrimaryButton>
  );
};

const IdentityListCell: React.FC<IdentityListCellProps> = function IdentityListCell(
  props: IdentityListCellProps
) {
  const {
    identityType,
    identityName,
    connectedOn,
    addedOn,
    verified,
    toggleVerified,
    remove,
  } = props;

  const icon = iconMap[identityType];

  return (
    <ListCellLayout className={styles.cellContainer}>
      <div className={styles.cellIcon}>{icon}</div>
      <Text className={styles.cellName}>{identityName}</Text>
      {verified != null && (
        <>
          {verified ? (
            <Text className={styles.cellDescVerified}>
              <FormattedMessage id="UserDetails.connected-identities.verified" />
            </Text>
          ) : (
            <Text className={styles.cellDescUnverified}>
              <FormattedMessage id="UserDetails.connected-identities.unverified" />
            </Text>
          )}
          <Text className={styles.cellDescSeparator}>{" | "}</Text>
        </>
      )}
      <Text className={styles.cellDesc}>
        {connectedOn != null && (
          <FormattedMessage
            id="UserDetails.connected-identities.connected-on"
            values={{ datetime: connectedOn }}
          />
        )}
        {addedOn != null && (
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: addedOn }}
          />
        )}
      </Text>
      {verified != null && toggleVerified != null && (
        <VerifyButton verified={verified} toggleVerified={toggleVerified} />
      )}
      <DefaultButton
        className={cn(styles.controlButton, styles.removeButton)}
        theme={destructiveTheme}
        onClick={remove}
      >
        <FormattedMessage id={removeButtonTextId[identityType]} />
      </DefaultButton>
    </ListCellLayout>
  );
};

const UserDetailsConnectedIdentities: React.FC<UserDetailsConnectedIdentitiesProps> = function UserDetailsConnectedIdentities(
  props: UserDetailsConnectedIdentitiesProps
) {
  const { identities } = props;
  const { locale, renderToString } = useContext(Context);
  const navigate = useNavigate();

  const identityLists: IdentityLists = useMemo(() => {
    const emailIdentityList: EmailIdentityListItem[] = [];
    const phoneIdentityList: PhoneIdentityListItem[] = [];
    const usernameIdentityList: UsernameIdentityListItem[] = [];

    // TODO: get actual verified state
    for (const identity of identities) {
      const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
      if (identity.claims.email != null) {
        emailIdentityList.push({
          email: identity.claims.email,
          verified: true,
          connectedOn: createdAtStr,
        });
      }

      if (identity.claims.phone_number != null) {
        phoneIdentityList.push({
          phone: identity.claims.phone_number,
          verified: false,
          addedOn: createdAtStr,
        });
      }

      if (identity.claims.preferred_username != null) {
        usernameIdentityList.push({
          username: identity.claims.preferred_username,
          addedOn: createdAtStr,
        });
      }
    }
    return {
      email: emailIdentityList,
      phone: phoneIdentityList,
      username: usernameIdentityList,
    };
  }, [locale, identities]);

  const onRenderEmailIdentityCell = useCallback(
    (item?: EmailIdentityListItem, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <IdentityListCell
          identityType="email"
          identityName={item.email}
          verified={item.verified}
          connectedOn={item.connectedOn}
          toggleVerified={() => {}}
        />
      );
    },
    []
  );

  const onRenderPhoneIdentityCell = useCallback(
    (item?: PhoneIdentityListItem, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <IdentityListCell
          identityType="phone"
          identityName={item.phone}
          verified={item.verified}
          addedOn={item.addedOn}
          toggleVerified={() => {}}
        />
      );
    },
    []
  );

  const onRenderUsernameIdentityCell = useCallback(
    (item?: UsernameIdentityListItem, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <IdentityListCell
          identityType="username"
          identityName={item.username}
          addedOn={item.addedOn}
        />
      );
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
      <section className={styles.identityLists}>
        {identityLists.email.length > 0 && (
          <>
            <Text as="h3" className={styles.subHeader}>
              <FormattedMessage id="UserDetails.connected-identities.email" />
            </Text>
            <List
              className={styles.list}
              items={identityLists.email}
              onRenderCell={onRenderEmailIdentityCell}
            />
          </>
        )}
        {identityLists.phone.length > 0 && (
          <>
            <Text as="h3" className={styles.subHeader}>
              <FormattedMessage id="UserDetails.connected-identities.phone" />
            </Text>
            <List
              className={styles.list}
              items={identityLists.phone}
              onRenderCell={onRenderPhoneIdentityCell}
            />
          </>
        )}
        {identityLists.username.length > 0 && (
          <>
            <Text as="h3" className={styles.subHeader}>
              <FormattedMessage id="UserDetails.connected-identities.username" />
            </Text>
            <List
              className={styles.list}
              items={identityLists.username}
              onRenderCell={onRenderUsernameIdentityCell}
            />
          </>
        )}
      </section>
    </div>
  );
};

export default UserDetailsConnectedIdentities;
