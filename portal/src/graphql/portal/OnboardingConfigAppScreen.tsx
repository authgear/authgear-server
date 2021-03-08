import React, { useCallback, useContext, useMemo } from "react";
import { Navigate, useParams } from "react-router-dom";
import produce from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Checkbox, FontIcon, Label, Text } from "@fluentui/react";

import {
  IdentityType,
  LoginIDKeyType,
  PortalAPIAppConfig,
  SecondaryAuthenticationMode,
  SecondaryAuthenticatorType,
  VerificationClaimsConfig,
} from "../../types";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenHeader from "../../ScreenHeader";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useAppConfigQuery } from "./query/appConfigQuery";
import OnboardingFormContainer from "./OnboardingFormContainer";
import styles from "./OnboardingConfigAppScreen.module.scss";

interface PendingFormState {
  identities: Set<IdentityType>;
  loginIDKeys: Set<LoginIDKeyType>;
}

interface FormState {
  pendingForm: PendingFormState;
}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {
    pendingForm: {
      identities: new Set<IdentityType>(),
      loginIDKeys: new Set<LoginIDKeyType>(),
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {});
}

interface IdentitiesButton {
  labelId: string;
  iconName: string;
  // button should have either identityType or loginIDType
  // has loginIDType implicitly means identityType == login_id
  identityType?: IdentityType;
  loginIDType?: LoginIDKeyType;
}

interface IdentitiesItemContentProps {
  form: AppConfigFormModel<FormState>;
  btnItem: IdentitiesButton;
}

const IdentitiesItemContent: React.FC<IdentitiesItemContentProps> = function IdentitiesItemContent(
  props
) {
  const {
    form: { state, setState },
    btnItem,
  } = props;

  const getCheckedState = useCallback(() => {
    // check only if button item has loginIDType (e.g. email, phone and username)
    if (btnItem.loginIDType) {
      return state.pendingForm.loginIDKeys.has(btnItem.loginIDType);
    } else if (btnItem.identityType) {
      return state.pendingForm.identities.has(btnItem.identityType);
    }
    console.error(
      "IdentitiesButton should have either identityType or loginIDType"
    );
    return false;
  }, [state, btnItem]);

  const onCheckedChange = useCallback(
    (checked?: boolean) => {
      const identities = new Set(state.pendingForm.identities);
      const loginIDKeys = new Set(state.pendingForm.loginIDKeys);
      if (btnItem.loginIDType) {
        if (checked) {
          loginIDKeys.add(btnItem.loginIDType);
        } else {
          loginIDKeys.delete(btnItem.loginIDType);
        }
      } else if (btnItem.identityType) {
        if (checked) {
          identities.add(btnItem.identityType);
        } else {
          identities.delete(btnItem.identityType);
        }
      } else {
        console.error(
          "IdentitiesButton should have either identityType or loginIDType"
        );
        return;
      }

      // check if there is any login id enabled
      // and update login_id in identities list
      if (loginIDKeys.size > 0) {
        identities.add("login_id");
      } else {
        identities.delete("login_id");
      }

      setState((prev) => ({
        ...prev,
        pendingForm: {
          ...prev.pendingForm,
          identities,
          loginIDKeys,
        },
      }));
    },
    [state, setState, btnItem]
  );

  const onItemClick = useCallback(
    (event: React.FormEvent) => {
      event.preventDefault();
      event.stopPropagation();
      const currentChecked = getCheckedState();
      onCheckedChange(!currentChecked);
    },
    [getCheckedState, onCheckedChange]
  );

  const onCheckboxChange = useCallback(
    (event?: React.FormEvent, checked?: boolean) => {
      event?.preventDefault();
      event?.stopPropagation();
      onCheckedChange(checked);
    },
    [onCheckedChange]
  );

  return (
    <div className={styles.identityListItem} onClick={onItemClick}>
      <div className={styles.label}>
        <FontIcon iconName={btnItem.iconName} className={styles.icon} />
        <Text block={true} variant="medium">
          <FormattedMessage id={btnItem.labelId} />
        </Text>
      </div>
      <Checkbox
        className={styles.checkbox}
        checked={getCheckedState()}
        onChange={onCheckboxChange}
      />
    </div>
  );
};

interface IdentitiesListContentProps {
  form: AppConfigFormModel<FormState>;
}

const IdentitiesListContent: React.FC<IdentitiesListContentProps> = function IdentitiesListContent(
  props
) {
  const { form } = props;

  const identitiesButtonItems: IdentitiesButton[] = useMemo(
    () => [
      {
        labelId: "Onboarding.identities.email",
        iconName: "Mail",
        loginIDType: "email",
      },
      {
        labelId: "Onboarding.identities.phone",
        iconName: "CellPhone",
        loginIDType: "phone",
      },
      {
        labelId: "Onboarding.identities.username",
        iconName: "ContactCard",
        loginIDType: "username",
      },
      {
        labelId: "Onboarding.identities.sso",
        iconName: "Globe",
        identityType: "oauth",
      },
      {
        labelId: "Onboarding.identities.anonymous-user",
        iconName: "FollowUser",
        identityType: "anonymous",
      },
    ],
    []
  );

  const showUsernameOnlyAlert = useMemo(
    () =>
      form.state.pendingForm.loginIDKeys.size === 1 &&
      form.state.pendingForm.loginIDKeys.has("username"),
    [form.state.pendingForm.loginIDKeys]
  );

  return (
    <section className={styles.sections}>
      <Label className={styles.fieldLabel}>
        <FontIcon iconName="Contact" className={styles.icon} />
        <FormattedMessage id="Onboarding.identities.label" />
      </Label>
      <div className={styles.identityList}>
        {identitiesButtonItems.map((btn, idx) => {
          return (
            <IdentitiesItemContent
              form={form}
              btnItem={btn}
              key={`identity-item-${idx}`}
            />
          );
        })}
      </div>
      {showUsernameOnlyAlert && (
        <Text className={styles.alertText} block={true} variant="small">
          <FontIcon iconName="AlertSolid" className={styles.icon} />
          <FormattedMessage id="Onboarding.identities.username-only-alert" />
        </Text>
      )}
    </section>
  );
};

interface OnboardingConfigAppScreenFormProps {
  form: AppConfigFormModel<FormState>;
}

const OnboardingConfigAppScreenForm: React.FC<OnboardingConfigAppScreenFormProps> = function OnboardingConfigAppScreenForm(
  props
) {
  const { form } = props;
  return (
    <div>
      <Text className={styles.pageTitle} block={true} variant="xLarge">
        <FormattedMessage id="Onboarding.title" />
      </Text>
      <Text className={styles.pageDesc} block={true} variant="small">
        <FormattedMessage id="Onboarding.desc" />
      </Text>
      <IdentitiesListContent form={form} />
    </div>
  );
};

const OnboardingConfigAppScreenContent: React.FC = function OnboardingConfigAppScreenContent() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <OnboardingFormContainer form={form}>
      <OnboardingConfigAppScreenForm form={form} />
    </OnboardingFormContainer>
  );
};

const OnboardingConfigAppScreen: React.FC = function OnboardingConfigAppScreen() {
  const { appID } = useParams();

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, loading, error } = useAppConfigQuery(appID);
  if (loading) {
    return <ShowLoading />;
  }
  const isInvalidAppID = error == null && effectiveAppConfig == null;
  if (isInvalidAppID) {
    return <Navigate to="/apps" replace={true} />;
  }

  return (
    <div className={styles.root}>
      <ScreenHeader />
      <OnboardingConfigAppScreenContent />
    </div>
  );
};

export default OnboardingConfigAppScreen;
