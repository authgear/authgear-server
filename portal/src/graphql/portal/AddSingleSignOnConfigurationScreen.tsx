import React, { useCallback, useMemo, useState } from "react";

import { FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";

import NavBreadcrumb from "../../NavBreadcrumb";
import ScreenContent from "../../ScreenContent";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
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

interface OAuthClientFormProps {
  alias: string;
  providerItemKey: OAuthSSOProviderItemKey;
  form: OAuthProviderFormModel;
}

const OAuthClientForm: React.VFC<OAuthClientFormProps> =
  function OAuthClientForm(props) {
    const { alias, providerItemKey, form } = props;
    const widgetProps = useSingleSignOnConfigurationWidget(
      alias,
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
    const [selectedProvider, setSelectedProvider] =
      useState<OAuthSSOProviderItemKey>();

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
      setSelectedProvider(itemKey);
    }, []);

    const onSaveSuccess = useCallback(() => {
      navigate("../");
    }, [navigate]);

    const newAlias = useMemo(() => {
      if (selectedProvider == null) {
        return null;
      }
      // FIXME
      return selectedProvider;
    }, [selectedProvider]);

    return (
      <FormContainer
        form={form}
        afterSave={onSaveSuccess}
        hideFooterComponent={selectedProvider == null}
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
            {newAlias != null && selectedProvider != null ? (
              <OAuthClientForm
                alias={newAlias}
                form={form}
                providerItemKey={selectedProvider}
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
