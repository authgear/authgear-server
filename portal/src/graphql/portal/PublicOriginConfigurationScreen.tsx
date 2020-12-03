import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import produce from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Dropdown, Text } from "@fluentui/react";
import { Domain, useDomainsQuery } from "./query/domainsQuery";
import { PortalAPIAppConfig } from "../../types";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useDropdown } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FormContainer from "../../FormContainer";

import styles from "./PublicOriginConfigurationScreen.module.scss";

function getOriginFromDomain(domain: Domain): string {
  // assume domain has no scheme
  // use https scheme
  return `https://${domain.domain}`;
}

interface FormState {
  publicOrigin: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.http ??= {};
    if (currentState.publicOrigin !== initialState.publicOrigin) {
      config.http.public_origin = currentState.publicOrigin;
    }
    clearEmptyObject(config);
  });
}

interface PublicOriginConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  domains: Domain[];
}

const PublicOriginConfigurationContent: React.FC<PublicOriginConfigurationContentProps> = function PublicOriginConfigurationContent(
  props
) {
  const {
    domains,
    form: { state, setState },
  } = props;

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="PublicOriginConfigurationScreen.title" />,
      },
    ];
  }, []);

  const availableOrigins = useMemo(() => {
    const verifiedDomains = domains.filter((domain) => domain.isVerified);
    const origins = new Set(verifiedDomains.map(getOriginFromDomain));
    origins.add(state.publicOrigin);
    // eslint-disable-next-line @typescript-eslint/require-array-sort-compare
    return Array.from(origins).sort();
  }, [domains, state.publicOrigin]);

  const {
    options: publicOriginOptions,
    onChange: onPublicOriginChange,
  } = useDropdown(
    availableOrigins,
    (origin) => {
      setState((state) => ({ ...state, publicOrigin: origin }));
    },
    state.publicOrigin
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Text className={styles.description}>
        <FormattedMessage id="PublicOriginConfigurationScreen.desc" />
      </Text>
      <Dropdown
        className={styles.field}
        options={publicOriginOptions}
        selectedKey={state.publicOrigin}
        onChange={onPublicOriginChange}
      />
    </div>
  );
};

const PublicOriginConfigurationScreen: React.FC = function PublicOriginConfigurationScreen() {
  const { appID } = useParams();
  const {
    domains,
    loading: isLoadingDomains,
    error: loadDomainsError,
    refetch: reloadDomains,
  } = useDomainsQuery(appID);

  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  const reload = useCallback(() => {
    reloadDomains().catch(() => {});
    form.reload();
  }, [reloadDomains, form]);

  if (form.isLoading || isLoadingDomains) {
    return <ShowLoading />;
  }

  if (form.loadError || loadDomainsError) {
    return (
      <ShowError error={form.loadError ?? loadDomainsError} onRetry={reload} />
    );
  }

  return (
    <FormContainer form={form}>
      <PublicOriginConfigurationContent form={form} domains={domains ?? []} />
    </FormContainer>
  );
};

export default PublicOriginConfigurationScreen;
