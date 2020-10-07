import React, { useMemo } from "react";
import {
  IColumn,
  Checkbox,
  SelectionMode,
  ICheckboxProps,
  DefaultEffects,
  Text,
} from "@fluentui/react";
import produce from "immer";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import DetailsListWithOrdering from "../../DetailsListWithOrdering";
import { swap } from "../../OrderButtons";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  PortalAPIAppConfig,
  primaryAuthenticatorTypes,
  secondaryAuthenticatorTypes,
  PrimaryAuthenticatorType,
  SecondaryAuthenticatorType,
  PortalAPIApp,
} from "../../types";
import { isArrayEqualInOrder, clearEmptyObject } from "../../util/misc";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";

interface Props {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

interface AuthenticatorCheckboxProps extends ICheckboxProps {
  authenticatorKey: string;
  onAuthticatorCheckboxChange: (key: string, checked: boolean) => void;
}

interface AuthenticatorListItem<KeyType> {
  activated: boolean;
  key: KeyType;
}

interface AuthenticationAuthenticatorScreenState {
  primaryAuthenticators: AuthenticatorListItem<PrimaryAuthenticatorType>[];
  secondaryAuthenticators: AuthenticatorListItem<SecondaryAuthenticatorType>[];
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
              ariaLabel={item.key}
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
        const newState = produce(prev, (draftState) => {
          draftState[itemIndex].activated = checked;
        });
        return newState;
      });
    },
    [state, setState]
  );
  return onActivateClicked;
}

// return list with all keys, active key from config in order
function makeAuthenticatorKeys<KeyType>(
  activeKeys: KeyType[],
  availableKeys: KeyType[]
) {
  const activeKeySet = new Set(activeKeys);
  const inactiveKeys = availableKeys.filter((key) => !activeKeySet.has(key));
  return [...activeKeys, ...inactiveKeys].map((key) => {
    return {
      activated: activeKeySet.has(key),
      key,
    };
  });
}

const constructListData = (
  appConfig: PortalAPIAppConfig | null
): AuthenticationAuthenticatorScreenState => {
  const authentication = appConfig?.authentication;

  const primaryAuthenticators = makeAuthenticatorKeys(
    authentication?.primary_authenticators ?? [],
    [...primaryAuthenticatorTypes]
  );
  const secondaryAuthenticators = makeAuthenticatorKeys(
    authentication?.secondary_authenticators ?? [],
    [...secondaryAuthenticatorTypes]
  );

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
  const {
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
  } = props;
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

  const initialState = useMemo(() => {
    return constructListData(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [
    primaryAuthenticatorState,
    setPrimaryAuthenticatorState,
  ] = React.useState(initialState.primaryAuthenticators);
  const [
    secondaryAuthenticatorState,
    setSecondaryAuthenticatorState,
  ] = React.useState(initialState.secondaryAuthenticators);

  const isFormModified = useMemo(() => {
    const screenState: AuthenticationAuthenticatorScreenState = {
      primaryAuthenticators: primaryAuthenticatorState,
      secondaryAuthenticators: secondaryAuthenticatorState,
    };
    return !deepEqual(initialState, screenState, { strict: true });
  }, [initialState, primaryAuthenticatorState, secondaryAuthenticatorState]);

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

  const renderPrimaryAriaLabel = React.useCallback(
    (index?: number): string => {
      return index != null ? primaryAuthenticatorState[index].key : "";
    },
    [primaryAuthenticatorState]
  );
  const renderSecondaryAriaLabel = React.useCallback(
    (index?: number): string => {
      return index != null ? secondaryAuthenticatorState[index].key : "";
    },
    [secondaryAuthenticatorState]
  );

  const onSaveButtonClicked = React.useCallback(() => {
    if (effectiveAppConfig == null || rawAppConfig == null) {
      return;
    }

    const initialActivatedPrimaryKeyList =
      effectiveAppConfig.authentication?.primary_authenticators ?? [];
    const initialActivatedSecondaryKeyList =
      effectiveAppConfig.authentication?.secondary_authenticators ?? [];

    const activatedPrimaryKeyList = getActivatedKeyListFromState(
      primaryAuthenticatorState
    );
    const activatedSecondaryKeyList = getActivatedKeyListFromState(
      secondaryAuthenticatorState
    );

    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.authentication = draftConfig.authentication ?? {};
      const { authentication } = draftConfig;
      if (
        !isArrayEqualInOrder(
          initialActivatedPrimaryKeyList,
          activatedPrimaryKeyList
        )
      ) {
        authentication.primary_authenticators = activatedPrimaryKeyList;
      }
      if (
        !isArrayEqualInOrder(
          initialActivatedSecondaryKeyList,
          activatedSecondaryKeyList
        )
      ) {
        authentication.secondary_authenticators = activatedSecondaryKeyList;
      }

      clearEmptyObject(draftConfig);
    });

    // TODO: handle error
    updateAppConfig(newAppConfig).catch(() => {});
  }, [
    rawAppConfig,
    effectiveAppConfig,
    updateAppConfig,
    primaryAuthenticatorState,
    secondaryAuthenticatorState,
  ]);

  return (
    <div className={styles.root}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <Text as="h2" className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticator.widgetHeader.primary" />
        </Text>
        <DetailsListWithOrdering
          items={primaryAuthenticatorState}
          columns={authenticatorColumns}
          onRenderItemColumn={renderPrimaryItemColumn}
          onSwapClicked={onPrimarySwapClicked}
          selectionMode={SelectionMode.none}
          renderAriaLabel={renderPrimaryAriaLabel}
        />
      </div>

      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <Text as="h2" className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticator.widgetHeader.secondary" />
        </Text>
        <DetailsListWithOrdering
          items={secondaryAuthenticatorState}
          columns={authenticatorColumns}
          onRenderItemColumn={renderSecondaryItemColumn}
          onSwapClicked={onSecondarySwapClicked}
          selectionMode={SelectionMode.none}
          renderAriaLabel={renderSecondaryAriaLabel}
        />
      </div>

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          disabled={!isFormModified}
          onClick={onSaveButtonClicked}
          loading={updatingAppConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
    </div>
  );
};

export default AuthenticationAuthenticatorSettings;
