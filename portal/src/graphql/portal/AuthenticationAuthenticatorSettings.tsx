import React from "react";
import {
  IColumn,
  Checkbox,
  SelectionMode,
  ICheckboxProps,
  PrimaryButton,
  DefaultEffects,
} from "@fluentui/react";
import produce from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import DetailsListWithOrdering, { swap } from "../../DetailsListWithOrdering";
import {
  PortalAPIAppConfig,
  primaryAuthenticatorTypes,
  secondaryAuthenticatorTypes,
  PrimaryAuthenticatorType,
  SecondaryAuthenticatorType,
} from "../../types";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";

interface Props {
  appConfig: PortalAPIAppConfig | null;
}

interface AuthenticatorCheckboxProps extends ICheckboxProps {
  authenticatorKey: string;
  onAuthticatorCheckboxChange: (key: string, checked: boolean) => void;
}

interface AuthenticatorListItem<KeyType> {
  activated: boolean;
  key: KeyType;
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

function useRenderItemColumn<KeyType extends string>(
  onCheckboxClicked: (key: string, checked: boolean) => void
) {
  const renderItemColumn = React.useCallback(
    (
      item: AuthenticatorListItem<KeyType>,
      _index?: number,
      column?: IColumn
    ) => {
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

function useOnActivateClicked<KeyType extends string>(
  state: AuthenticatorListItem<KeyType>[],
  setState: React.Dispatch<
    React.SetStateAction<AuthenticatorListItem<KeyType>[]>
  >
) {
  const onActivateClicked = React.useCallback(
    (key: string, checked: boolean) => {
      const itemIndex = state.findIndex(
        (authenticator) => authenticator.key === key
      );
      if (itemIndex < 0) {
        return;
      }
      setState((prev: AuthenticatorListItem<KeyType>[]) => {
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
  primaryAuthenticators: AuthenticatorListItem<PrimaryAuthenticatorType>[];
  secondaryAuthenticators: AuthenticatorListItem<SecondaryAuthenticatorType>[];
} => {
  const authentication = appConfig?.authentication;
  const primaryAuthenticatorKeys = new Set(
    authentication?.primary_authenticators
  );
  const secondaryAuthenticatorKeys = new Set(
    authentication?.secondary_authenticators
  );

  const primaryAuthenticators = primaryAuthenticatorTypes.map((key) => {
    return {
      activated: primaryAuthenticatorKeys.has(key),
      key,
    };
  });
  const secondaryAuthenticators = secondaryAuthenticatorTypes.map((key) => {
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

function getActivatedKeyListFromState<KeyType>(
  state: AuthenticatorListItem<KeyType>[]
) {
  return state
    .filter((authenticator) => authenticator.activated)
    .map((authenticator) => authenticator.key);
}

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
      minWidth: 120,
      maxWidth: 120,
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

  const onPrimarySwapClicked = React.useCallback(
    (index1: number, index2: number) => {
      setPrimaryAuthenticatorState(
        swap(primaryAuthenticatorState, index1, index2)
      );
    },
    [primaryAuthenticatorState]
  );
  const onSecondarySwapClicked = React.useCallback(
    (index1: number, index2: number) => {
      setSecondaryAuthenticatorState(
        swap(secondaryAuthenticatorState, index1, index2)
      );
    },
    [secondaryAuthenticatorState]
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

  const onSaveButtonClicked = React.useCallback(() => {
    if (props.appConfig == null) {
      return;
    }

    const activatedPrimaryKeyList = getActivatedKeyListFromState(
      primaryAuthenticatorState
    );
    const activatedSecondaryKeyList = getActivatedKeyListFromState(
      secondaryAuthenticatorState
    );

    const newAppConfig = produce(props.appConfig, (draftConfig) => {
      const authentication = draftConfig.authentication;
      authentication.primary_authenticators = activatedPrimaryKeyList;
      authentication.secondary_authenticators = activatedSecondaryKeyList;
    });

    // TODO: call mutation to save config
    console.log(newAppConfig);
  }, [props.appConfig, primaryAuthenticatorState, secondaryAuthenticatorState]);

  return (
    <div className={styles.root}>
      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <h2 className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticator.widgetHeader.primary" />
        </h2>
        <DetailsListWithOrdering
          items={primaryAuthenticatorState}
          columns={authenticatorColumns}
          onRenderItemColumn={renderPrimaryItemColumn}
          onSwapClicked={onPrimarySwapClicked}
          selectionMode={SelectionMode.none}
        />
      </div>

      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <h2 className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticator.widgetHeader.secondary" />
        </h2>
        <DetailsListWithOrdering
          items={secondaryAuthenticatorState}
          columns={authenticatorColumns}
          onRenderItemColumn={renderSecondaryItemColumn}
          onSwapClicked={onSecondarySwapClicked}
          selectionMode={SelectionMode.none}
        />
      </div>

      <div className={styles.saveButtonContainer}>
        <PrimaryButton onClick={onSaveButtonClicked}>
          <FormattedMessage id="save" />
        </PrimaryButton>
      </div>
    </div>
  );
};

export default AuthenticationAuthenticatorSettings;
