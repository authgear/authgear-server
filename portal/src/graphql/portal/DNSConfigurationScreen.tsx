import React, { useMemo, useContext, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import cn from "classnames";
import produce from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Text,
  DetailsList,
  IColumn,
  Stack,
  SelectionMode,
  IDetailsListProps,
  ActionButton,
  VerticalDivider,
  TextField,
  ITextFieldProps,
} from "@fluentui/react";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { Domain, useDomainsQuery } from "./query/domainsQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { PortalAPIAppConfig } from "../../types";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import FormTextField from "../../FormTextField";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { actionButtonTheme, destructiveTheme } from "../../theme";
import { useTextField } from "../../hook/useInput";
import { FormContext } from "../../error/FormContext";
import { useValidationError } from "../../error/useValidationError";
import { clearEmptyObject } from "../../util/misc";

import styles from "./DNSConfigurationScreen.module.scss";

interface DNSConfigurationProps {
  domains: Domain[];
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
}

interface PublicOriginConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
}

interface DomainListItem {
  id: string;
  domain: string;
  isVerified: boolean;
}

interface DomainListActionButtonsProps {
  id: string;
  isVerified: boolean;
  onDeleteClick: (id: string) => void;
}

const ADD_DOMAIN_TEXT_FIELD_STYLES: ITextFieldProps["styles"] = {
  fieldGroup: {
    borderRadius: "2px 0 0 2px",
    borderRightWidth: "0",
  },
};

const DOMAIN_LIST_STYLES: IDetailsListProps["styles"] = {
  headerWrapper: { marginTop: "-10px" },
};

function makeDomainListColumn(renderToString: (messageID: string) => string) {
  return [
    {
      key: "domain",
      name: renderToString("DNSConfigurationScreen.domain-list.header.domain"),
      minWidth: 250,
      className: styles.domainListColumn,
    },
    {
      key: "isVerified",
      name: renderToString("DNSConfigurationScreen.domain-list.header.status"),
      minWidth: 100,
      className: styles.domainListColumn,
    },
    {
      key: "action",
      name: renderToString("action"),
      minWidth: 150,
      className: styles.domainListColumn,
    },
  ];
}

const PublicOriginConfiguration: React.FC<PublicOriginConfigurationProps> = function PublicOriginConfiguration(
  props: PublicOriginConfigurationProps
) {
  const { rawAppConfig, effectiveAppConfig } = props;
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialPublicOrigin = useMemo(() => {
    return effectiveAppConfig?.http?.public_origin ?? "";
  }, [effectiveAppConfig]);

  const { value: publicOrigin, onChange: onPublicOriginChange } = useTextField(
    initialPublicOrigin
  );

  const isModified = useMemo(() => {
    return initialPublicOrigin !== publicOrigin;
  }, [initialPublicOrigin, publicOrigin]);

  const onSaveClick = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      const newPublicOrigin =
        publicOrigin.trim() !== "" ? publicOrigin : undefined;
      draftConfig.http = draftConfig.http ?? {};
      draftConfig.http.public_origin = newPublicOrigin;

      clearEmptyObject(draftConfig);
    });

    updateAppConfig(newAppConfig).catch(() => {});
  }, [publicOrigin, rawAppConfig, updateAppConfig]);

  const {
    value: formContextValue,
    otherError,
    unhandledCauses,
  } = useValidationError(updateAppConfigError);

  return (
    <FormContext.Provider value={formContextValue}>
      <section className={styles.publicOrigin}>
        {otherError && <ShowError error={otherError} />}
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        <Text
          as="h2"
          className={cn(
            styles.header,
            styles.subHeader,
            styles.publicOriginHeader
          )}
        >
          <FormattedMessage id="DNSConfigurationScreen.public-origin.header" />
        </Text>
        <div className={styles.publicOriginInput}>
          <FormTextField
            jsonPointer="/http/public_origin"
            parentJSONPointer="/http"
            fieldName="public_origin"
            fieldNameMessageID="DNSConfigurationScreen.public-origin.header"
            hideLabel={true}
            className={styles.publicOriginField}
            value={publicOrigin}
            onChange={onPublicOriginChange}
          />
          <ButtonWithLoading
            className={styles.savePublicOriginButton}
            disabled={!isModified}
            labelId="save"
            loadingLabelId="saving"
            loading={updatingAppConfig}
            onClick={onSaveClick}
          />
        </div>
      </section>
    </FormContext.Provider>
  );
};

const AddDomainSection: React.FC = function AddDomainSection() {
  const { renderToString } = useContext(Context);
  const { value: newDomain, onChange: onNewDomainChange } = useTextField("");

  const onAddClick = useCallback(() => {
    // TODO: To be implemented
  }, []);

  return (
    <section className={styles.addDomain}>
      <TextField
        className={styles.addDomainField}
        placeholder={renderToString(
          "DNSConfigurationScreen.domain-list.add-domain.placeholder"
        )}
        styles={ADD_DOMAIN_TEXT_FIELD_STYLES}
        value={newDomain}
        onChange={onNewDomainChange}
      />
      <ButtonWithLoading
        className={styles.addDomainButton}
        iconProps={{ iconName: "CircleAdditionSolid" }}
        loading={false}
        labelId="add"
        onClick={onAddClick}
      />
    </section>
  );
};

const DomainListActionButtons: React.FC<DomainListActionButtonsProps> = function DomainListActionButtons(
  props: DomainListActionButtonsProps
) {
  const { id, isVerified, onDeleteClick: onDeleteClickProps } = props;

  const navigate = useNavigate();

  const onVerifyClicked = useCallback(() => {
    navigate(`./${id}/verify`);
  }, [id, navigate]);

  const onDeleteClick = useCallback(() => {
    onDeleteClickProps(id);
  }, [id, onDeleteClickProps]);

  return (
    <section className={styles.actionButtonContainer}>
      {!isVerified && (
        <>
          <ActionButton
            className={styles.actionButton}
            theme={actionButtonTheme}
            onClick={onVerifyClicked}
          >
            <FormattedMessage id="verify" />
          </ActionButton>
          <VerticalDivider className={styles.divider} />
        </>
      )}
      <ActionButton
        className={styles.actionButton}
        theme={destructiveTheme}
        onClick={onDeleteClick}
      >
        <FormattedMessage id="delete" />
      </ActionButton>
    </section>
  );
};

const DNSConfiguration: React.FC<DNSConfigurationProps> = function DNSConfiguration(
  props: DNSConfigurationProps
) {
  const { domains, rawAppConfig, effectiveAppConfig } = props;

  const { renderToString } = useContext(Context);

  const domainListColumns: IColumn[] = useMemo(() => {
    return makeDomainListColumn(renderToString);
  }, [renderToString]);

  const domainListItems: DomainListItem[] = useMemo(() => {
    return domains.map((domain) => ({
      id: domain.id,
      domain: domain.domain,
      isVerified: domain.isVerified,
    }));
  }, [domains]);

  const onDeleteClick = useCallback((_id: string) => {
    // TODO: to be implemented
  }, []);

  const renderDomainListColumn = useCallback<
    Required<IDetailsListProps>["onRenderItemColumn"]
  >(
    (item: DomainListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "domain":
          return <span>{item.domain}</span>;
        case "isVerified": {
          if (item.isVerified) {
            return (
              <span>
                <FormattedMessage id="DNSConfigurationScreen.domain-list.status.verified" />
              </span>
            );
          }
          return (
            <span>
              <FormattedMessage id="DNSConfigurationScreen.domain-list.status.not-verified" />
            </span>
          );
        }
        case "action":
          return (
            <DomainListActionButtons
              id={item.id}
              isVerified={item.isVerified}
              onDeleteClick={onDeleteClick}
            />
          );
        default:
          return null;
      }
    },
    [onDeleteClick]
  );

  const renderDomainListHeader = useCallback<
    Required<IDetailsListProps>["onRenderDetailsHeader"]
  >((props, defaultRenderer) => {
    const defaultHeaderNode = defaultRenderer?.(props) ?? null;
    return (
      <>
        {defaultHeaderNode}
        <AddDomainSection />
      </>
    );
  }, []);

  return (
    <section className={styles.content}>
      <PublicOriginConfiguration
        rawAppConfig={rawAppConfig}
        effectiveAppConfig={effectiveAppConfig}
      />
      <Text as="h2" className={cn(styles.header, styles.subHeader)}>
        <FormattedMessage id="DNSConfigurationScreen.domain-list.title" />
      </Text>
      <DetailsList
        columns={domainListColumns}
        items={domainListItems}
        styles={DOMAIN_LIST_STYLES}
        selectionMode={SelectionMode.none}
        onRenderItemColumn={renderDomainListColumn}
        onRenderDetailsHeader={renderDomainListHeader}
      />
    </section>
  );
};

const DNSConfigurationScreen: React.FC = function DNSConfigurationScreen() {
  const { appID } = useParams();
  const {
    effectiveAppConfig,
    rawAppConfig,
    loading: fetchingAppConfig,
    error: fetchAppConfigError,
    refetch: refetchAppConfig,
  } = useAppConfigQuery(appID);
  const {
    domains,
    loading: fetchingDomains,
    error: fetchDomainsError,
    refetch: refetchDomains,
  } = useDomainsQuery(appID);

  if (fetchingAppConfig || fetchingDomains) {
    return <ShowLoading />;
  }

  if (fetchAppConfigError != null || fetchDomainsError != null) {
    return (
      <Stack>
        {fetchAppConfigError && (
          <ShowError error={fetchAppConfigError} onRetry={refetchAppConfig} />
        )}
        {fetchDomainsError && (
          <ShowError error={fetchDomainsError} onRetry={refetchDomains} />
        )}
      </Stack>
    );
  }
  return (
    <main className={styles.root}>
      <Text className={cn(styles.header, styles.mainHeader)} as="h1">
        <FormattedMessage id="DNSConfigurationScreen.title" />
      </Text>
      <Text className={styles.desc}>
        <FormattedMessage id="DNSConfigurationScreen.desc" />
      </Text>
      <DNSConfiguration
        effectiveAppConfig={effectiveAppConfig}
        rawAppConfig={rawAppConfig}
        domains={domains ?? []}
      />
    </main>
  );
};

export default DNSConfigurationScreen;
