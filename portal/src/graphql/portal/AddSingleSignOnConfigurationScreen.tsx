import React, { useCallback, useEffect, useMemo, useState } from "react";

import { FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";

import NavBreadcrumb from "../../NavBreadcrumb";
import ScreenContent from "../../ScreenContent";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
  parseOAuthSSOProviderItemKey,
} from "../../types";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import styles from "./AddSingleSignOnConfigurationScreen.module.css";
import SingleSignOnConfigurationWidget, {
  OAuthClientCard,
  useSingleSignOnConfigurationWidget,
} from "./SingleSignOnConfigurationWidget";
import ScreenContentHeader from "../../ScreenContentHeader";
import {
  OAuthProviderFormModel,
  SSOProviderFormState,
  useOAuthProviderForm,
} from "../../hook/useOAuthProviderForm";
import { useNavigate, useParams } from "react-router-dom";
import FormContainer from "../../FormContainer";

interface OAuthClientMenuProps {
  form: OAuthProviderFormModel;
  onSelect: (providerItemKey: OAuthSSOProviderItemKey) => void;
}

const OAuthClientMenu: React.VFC<OAuthClientMenuProps> =
  function OAuthClientMenu(props) {
    const { form, onSelect } = props;
    const providerPropsList = useMemo(() => {
      const existingProviders = form.state.providers.map((p) =>
        createOAuthSSOProviderItemKey(p.config.type, p.config.app_type)
      );
      return oauthSSOProviderItemKeys.map((providerItemKey) => {
        return {
          providerItemKey,
          isAdded: existingProviders.includes(providerItemKey),
        };
      });
    }, [form.state.providers]);
    return (
      <div className={styles.widget}>
        <div className={styles.providerGrid}>
          {providerPropsList.map((providerProps) => (
            <OAuthClientCard
              key={providerProps.providerItemKey}
              className={styles.providerCard}
              providerItemKey={providerProps.providerItemKey}
              isAdded={providerProps.isAdded}
              onAddClick={onSelect}
            />
          ))}
        </div>
      </div>
    );
  };

function generateNewAlias(
  existingProviders: SSOProviderFormState[],
  providerType: OAuthSSOProviderType,
  appType?: OAuthSSOWeChatAppType
) {
  const aliasPrefix = appType
    ? [providerType, appType].join("_")
    : providerType;
  let alias = aliasPrefix;
  let counter = 0;
  const existingAliases = new Set(
    existingProviders.map((provider) => provider.config.alias)
  );
  while (existingAliases.has(alias)) {
    counter += 1;
    alias = `${aliasPrefix}${counter}`;
  }
  return alias;
}

interface OAuthClientFormProps {
  initialAlias: string;
  providerItemKey: OAuthSSOProviderItemKey;
  form: OAuthProviderFormModel;
}

const OAuthClientForm: React.VFC<OAuthClientFormProps> =
  function OAuthClientForm(props) {
    const { initialAlias, providerItemKey, form } = props;
    const widgetProps = useSingleSignOnConfigurationWidget(
      initialAlias,
      providerItemKey,
      form
    );
    return (
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        {...widgetProps}
      />
    );
  };

const AddSingleSignOnConfigurationContent: React.VFC =
  function AddSingleSignOnConfigurationContent() {
    const navigate = useNavigate();
    const { appID } = useParams() as { appID: string };
    const form = useOAuthProviderForm(appID, null);
    const [selectedProviderKey, setSelectedProviderKey] =
      useState<OAuthSSOProviderItemKey>();
    const [newAlias, setNewAlias] = useState<string | null>(null);

    const navBreadcrumbItems = useMemo(() => {
      return [
        {
          to: "..",
          label: (
            <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: (
            <FormattedMessage id="AddSingleSignOnConfigurationScreen.title" />
          ),
        },
      ];
    }, []);

    const onMenuSelect = useCallback((itemKey: OAuthSSOProviderItemKey) => {
      setSelectedProviderKey(itemKey);
    }, []);

    const onSaveSuccess = useCallback(() => {
      navigate("../");
    }, [navigate]);

    useEffect(() => {
      if (selectedProviderKey == null) {
        setNewAlias(null);
        return;
      }
      const [providerType, appType] =
        parseOAuthSSOProviderItemKey(selectedProviderKey);
      setNewAlias(
        generateNewAlias(form.state.providers, providerType, appType)
      );
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedProviderKey]);

    return (
      <FormContainer
        form={form}
        afterSave={onSaveSuccess}
        hideFooterComponent={selectedProviderKey == null}
      >
        <ScreenContent
          header={
            <ScreenContentHeader
              title={
                <NavBreadcrumb
                  className={cn(styles.widget, styles.breadcrumb)}
                  items={navBreadcrumbItems}
                />
              }
            />
          }
        >
          <ShowOnlyIfSIWEIsDisabled>
            {newAlias != null && selectedProviderKey != null ? (
              <OAuthClientForm
                initialAlias={newAlias}
                form={form}
                providerItemKey={selectedProviderKey}
              />
            ) : (
              <OAuthClientMenu form={form} onSelect={onMenuSelect} />
            )}
          </ShowOnlyIfSIWEIsDisabled>
        </ScreenContent>
      </FormContainer>
    );
  };

const AddSingleSignOnConfigurationScreen: React.VFC = () => {
  return <AddSingleSignOnConfigurationContent />;
};

export default AddSingleSignOnConfigurationScreen;
