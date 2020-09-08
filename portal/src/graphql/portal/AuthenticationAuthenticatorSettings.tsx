import React from "react";
import {
  IColumn,
  Checkbox,
  SelectionMode,
  ICheckboxProps,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import DetailsListWithOrdering, {
  useOnSwapClicked,
} from "../../DetailsListWithOrdering";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";

interface Props {
  appConfig: PortalAPIAppConfig | null;
}

interface AuthenticatorCheckboxProps extends ICheckboxProps {
  authenticatorKey: string;
  onAuthticatorCheckboxChange: (key: string, checked: boolean) => void;
}

interface AuthenticatorListItem {
  activated: boolean;
  key: string;
}

const AuthenticatorCheckbox: React.FC<AuthenticatorCheckboxProps> = function AuthenticatorCheckbox(
  props: AuthenticatorCheckboxProps
) {
  const onChange = React.useCallback(
    (_event, checked?: boolean) => {
      props.onAuthticatorCheckboxChange(props.authenticatorKey, !!checked);
    },
    [props]
  );

  return <Checkbox {...props} onChange={onChange} />;
};

function useRenderItemColumn(
  onCheckboxClicked: (key: string, checked: boolean) => void
) {
  const renderItemColumn = React.useCallback(
    (item: AuthenticatorListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "activated":
          return (
            <AuthenticatorCheckbox
              authenticatorKey={item.key}
              checked={item.activated}
              onAuthticatorCheckboxChange={onCheckboxClicked}
            />
          );

        case "key":
          return <span>{item.key}</span>;

        default:
          return <span>{item.key}</span>;
      }
    },
    [onCheckboxClicked]
  );
  return renderItemColumn;
}

function useOnActivateClicked(
  state: AuthenticatorListItem[],
  setState: React.Dispatch<React.SetStateAction<AuthenticatorListItem[]>>
) {
  const onActivateClicked = React.useCallback(
    (key: string, checked: boolean) => {
      const itemIndex = state.findIndex(
        (authenticator) => authenticator.key === key
      );
      if (itemIndex < 0) {
        return;
      }
      setState((prev: AuthenticatorListItem[]) => {
        prev[itemIndex].activated = checked;
        return [...prev];
      });
    },
    [state, setState]
  );
  return onActivateClicked;
}

const constructListData = (
  appConfig: PortalAPIAppConfig | null
): {
  primaryAuthenticators: AuthenticatorListItem[];
  secondaryAuthenticators: AuthenticatorListItem[];
} => {
  const authenticators = appConfig?.authenticator ?? {};
  const authentication = appConfig?.authentication;
  const primaryAuthenticatorKeys = new Set(
    authentication?.primary_authenticators
  );
  const secondaryAuthenticatorKeys = new Set(
    authentication?.secondary_authenticators
  );
  const availableAuthenticatorKeys = Object.keys(authenticators);

  const primaryAuthenticators = availableAuthenticatorKeys.map((key) => {
    return {
      activated: primaryAuthenticatorKeys.has(key),
      key,
    };
  });
  const secondaryAuthenticators = availableAuthenticatorKeys.map((key) => {
    return {
      activated: secondaryAuthenticatorKeys.has(key),
      key,
    };
  });

  return {
    primaryAuthenticators,
    secondaryAuthenticators,
  };
};

const AuthenticationAuthenticatorSettings: React.FC<Props> = function AuthenticationAuthenticatorSettings(
  props: Props
) {
  const { renderToString } = React.useContext(Context);

  const authenticatorColumns: IColumn[] = [
    {
      key: "activated",
      fieldName: "activated",
      name: renderToString("AuthenticationAuthenticator.activateHeader"),
      className: styles.authenticatorColumn,
      minWidth: 150,
      maxWidth: 150,
    },
    {
      key: "key",
      fieldName: "key",
      name: renderToString("AuthenticationAuthenticator.authenticatorHeader"),
      className: styles.authenticatorColumn,
      minWidth: 300,
      maxWidth: 300,
    },
  ];

  const { primaryAuthenticators, secondaryAuthenticators } = constructListData(
    props.appConfig
  );

  const [
    primaryAuthenticatorState,
    setPrimaryAuthenticatorState,
  ] = React.useState(primaryAuthenticators);
  const [
    secondaryAuthenticatorState,
    setSecondaryAuthenticatorState,
  ] = React.useState(secondaryAuthenticators);

  const onPrimarySwapClicked = useOnSwapClicked(
    primaryAuthenticators,
    setPrimaryAuthenticatorState
  );
  const onSecondarySwapClicked = useOnSwapClicked(
    secondaryAuthenticators,
    setSecondaryAuthenticatorState
  );

  const onPrimaryActivateClicked = useOnActivateClicked(
    primaryAuthenticatorState,
    setPrimaryAuthenticatorState
  );
  const onSecondaryActivateClicked = useOnActivateClicked(
    secondaryAuthenticatorState,
    setSecondaryAuthenticatorState
  );

  const renderPrimaryItemColumn = useRenderItemColumn(onPrimaryActivateClicked);
  const renderSecondaryItemColumn = useRenderItemColumn(
    onSecondaryActivateClicked
  );

  return (
    <div className={styles.root}>
      <DetailsListWithOrdering
        items={primaryAuthenticatorState}
        columns={authenticatorColumns}
        onRenderItemColumn={renderPrimaryItemColumn}
        onSwapClicked={onPrimarySwapClicked}
        selectionMode={SelectionMode.none}
      />
      <DetailsListWithOrdering
        items={secondaryAuthenticatorState}
        columns={authenticatorColumns}
        onRenderItemColumn={renderSecondaryItemColumn}
        onSwapClicked={onSecondarySwapClicked}
        selectionMode={SelectionMode.none}
      />
    </div>
  );
};

export default AuthenticationAuthenticatorSettings;
