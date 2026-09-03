import React, { useCallback, useContext, useMemo, useState } from "react";
import { Text } from "@radix-ui/themes";
import { QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import type {
  CIMDDomainMode,
  FormState as ApplicationsFormState,
} from "../../graphql/portal/ApplicationsConfigurationScreen";
import { AppConfigFormModel } from "../../hook/useAppConfigForm";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { Toggle } from "../v2/Toggle/Toggle";
import { TextField } from "../v2/TextField/TextField";
import { TextFieldList } from "../v2/TextFieldList/TextFieldList";
import { RadioCards, RadioCardOption } from "../v2/RadioCards/RadioCards";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import { MechanismSectionHeader } from "./MechanismSectionHeader";
import styles from "./CIMDSection.module.css";

// JSON pointers of the objects holding the fields, so the config schema's
// validation errors -- "minimum" on a lifetime, "pattern" on a domain --
// bind to the field that caused them instead of only reaching the generic
// error bar. TextFieldList appends the item index itself.
const CIMD_JSON_POINTER = "/oauth/client_id_metadata_document";
const CIMD_CLIENT_CONFIG_JSON_POINTER =
  "/oauth/client_id_metadata_document/client_config";

export interface CIMDSectionProps {
  form: AppConfigFormModel<ApplicationsFormState>;
}

export const CIMDSection: React.VFC<CIMDSectionProps> = function CIMDSection({
  form,
}) {
  const { state, setState, isUpdating, effectiveConfig } = form;
  const { renderToString } = useContext(Context);

  // Every client_config field is resolved server-side by
  // OAuthDynamicClientTokenLifetimesConfig.SetDefaults(), so the effective
  // config carries concrete values even when authgear.yaml omits the whole
  // section. Read the placeholders from there rather than restating the Go
  // defaults here, where they would drift silently.
  const effectiveClientConfig =
    effectiveConfig.oauth?.client_id_metadata_document?.client_config;

  const [isAnyDomainConfirmationVisible, setIsAnyDomainConfirmationVisible] =
    useState(false);

  const enabled = state.cimdEnabled;

  const onEnabledChange = useCallback(
    (checked: boolean) => {
      // Deferred like every other control on this tab: nothing is written
      // until the admin presses Save. Saving from here would commit the
      // whole form state, including edits elsewhere on the tab and on the
      // sibling Applications tab, which share this form model.
      setState((prev) => ({
        ...prev,
        cimdEnabled: checked,
      }));
    },
    [setState]
  );

  const setDomainMode = useCallback(
    (mode: CIMDDomainMode) => {
      setState((prev) => ({
        ...prev,
        cimdDomainMode: mode,
        // Entering "Only these domains" with nothing listed would describe a
        // restriction that does not exist, so seed one empty row: the admin
        // sees where the domain goes, and an unfilled row fails validation
        // on save instead of quietly meaning "any domain".
        cimdAllowedDomains:
          mode === "list" && prev.cimdAllowedDomains.length === 0
            ? [""]
            : prev.cimdAllowedDomains,
      }));
    },
    [setState]
  );

  const onDomainModeChange = useCallback(
    (mode: CIMDDomainMode) => {
      if (mode === "any" && state.cimdDomainMode === "list") {
        // Widening from a restricted allowlist to "any domain" is the
        // permissive act on this card -- confirm it before reflecting it in
        // the form state.
        setIsAnyDomainConfirmationVisible(true);
        return;
      }
      setDomainMode(mode);
    },
    [state.cimdDomainMode, setDomainMode]
  );

  const onConfirmAnyDomain = useCallback(() => {
    setIsAnyDomainConfirmationVisible(false);
    setDomainMode("any");
  }, [setDomainMode]);

  const onCancelAnyDomain = useCallback(() => {
    setIsAnyDomainConfirmationVisible(false);
  }, []);

  const setAllowedDomains = useCallback(
    (domains: string[]) => {
      setState((prev) => ({
        ...prev,
        cimdAllowedDomains: domains,
      }));
    },
    [setState]
  );

  // TextFieldList hands back the CURRENT list, so each callback computes the
  // updated one itself.
  const onDomainAdd = useCallback(
    (list: string[], item: string) => {
      setAllowedDomains([...list, item]);
    },
    [setAllowedDomains]
  );

  const onDomainChange = useCallback(
    (list: string[], index: number, item: string) => {
      setAllowedDomains(list.map((value, i) => (i === index ? item : value)));
    },
    [setAllowedDomains]
  );

  const onDomainDelete = useCallback(
    (list: string[], index: number) => {
      setAllowedDomains(list.filter((_, i) => i !== index));
    },
    [setAllowedDomains]
  );

  const onAccessTokenLifetimeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({
        ...prev,
        cimdAccessTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
      }));
    },
    [setState]
  );

  const onRefreshTokenLifetimeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({
        ...prev,
        cimdRefreshTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
      }));
    },
    [setState]
  );

  const onRefreshTokenIdleTimeoutEnabledChange = useCallback(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        cimdRefreshTokenIdleTimeoutEnabled: checked,
      }));
    },
    [setState]
  );

  const onRefreshTokenIdleTimeoutChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({
        ...prev,
        cimdRefreshTokenIdleTimeoutSeconds:
          parseIntegerAllowLeadingZeros(value),
      }));
    },
    [setState]
  );

  const domainModeOptions = useMemo<RadioCardOption<CIMDDomainMode>[]>(
    () => [
      {
        value: "any",
        disabled: isUpdating,
        title: <FormattedMessage id="CIMDSection.trusted-domains.any.title" />,
        subtitle: (
          <FormattedMessage id="CIMDSection.trusted-domains.any.subtitle" />
        ),
      },
      {
        value: "list",
        disabled: isUpdating,
        title: <FormattedMessage id="CIMDSection.trusted-domains.list.title" />,
        subtitle: (
          <FormattedMessage id="CIMDSection.trusted-domains.list.subtitle" />
        ),
      },
    ],
    [isUpdating]
  );

  return (
    <>
      <MechanismSectionHeader
        title={<FormattedMessage id="CIMDSection.title" />}
        description={
          <FormattedMessage
            id="CIMDSection.enable.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              mcpLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/get-started/auth-for-mcp">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        }
        toggleLabel={<FormattedMessage id="CIMDSection.enable.toggle.label" />}
        checked={enabled}
        disabled={isUpdating}
        onCheckedChange={onEnabledChange}
      />

      {enabled ? (
        <SettingsSectionCard
          contentClassName="gap-4"
          title={<FormattedMessage id="CIMDSection.trusted-domains.title" />}
          description={
            <FormattedMessage id="CIMDSection.trusted-domains.note" />
          }
        >
          <RadioCards
            size="1"
            highContrast={true}
            itemMinWidth={220}
            itemFillSpaces={true}
            value={state.cimdDomainMode}
            onValueChange={onDomainModeChange}
            options={domainModeOptions}
          />
          {state.cimdDomainMode === "list" ? (
            <TextFieldList
              parentJSONPointer={CIMD_JSON_POINTER}
              fieldName="allowed_domains"
              list={state.cimdAllowedDomains}
              onListItemAdd={onDomainAdd}
              onListItemChange={onDomainChange}
              onListItemDelete={onDomainDelete}
              addButtonLabelMessageID="CIMDSection.trusted-domains.add"
              deleteButtonAriaLabel={renderToString(
                "CIMDSection.trusted-domains.delete"
              )}
              placeholder={renderToString(
                "CIMDSection.trusted-domains.placeholder"
              )}
              itemErrorMessageID="CIMDSection.trusted-domains.invalid"
              label={
                <span className={styles.labelWithHint}>
                  <FormattedMessage id="CIMDSection.trusted-domains.label" />
                  <Tooltip
                    content={renderToString(
                      "CIMDSection.trusted-domains.wildcard-hint"
                    )}
                  >
                    <QuestionMarkCircledIcon
                      className={styles.hintIcon}
                      width="1rem"
                      height="1rem"
                    />
                  </Tooltip>
                </span>
              }
              disabled={isUpdating}
              minItem={1}
            />
          ) : null}
        </SettingsSectionCard>
      ) : null}

      {enabled ? (
        <SettingsSectionCard
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientsTab.client-config.title" />
          }
          description={
            <FormattedMessage id="CIMDSection.client-config.description" />
          }
        >
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="DynamicClientsTab.access-token-lifetime.label" />
            }
            parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
            fieldName="access_token_lifetime_seconds"
            placeholder={effectiveClientConfig?.access_token_lifetime_seconds?.toFixed(
              0
            )}
            value={state.cimdAccessTokenLifetimeSeconds?.toFixed(0) ?? ""}
            onChange={onAccessTokenLifetimeChange}
          />
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="DynamicClientsTab.refresh-token-lifetime.label" />
            }
            parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
            fieldName="refresh_token_lifetime_seconds"
            placeholder={effectiveClientConfig?.refresh_token_lifetime_seconds?.toFixed(
              0
            )}
            value={state.cimdRefreshTokenLifetimeSeconds?.toFixed(0) ?? ""}
            onChange={onRefreshTokenLifetimeChange}
          />
          <div className="flex flex-col gap-1">
            <Toggle
              checked={state.cimdRefreshTokenIdleTimeoutEnabled}
              disabled={isUpdating}
              onCheckedChange={onRefreshTokenIdleTimeoutEnabledChange}
              text={
                <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.description" />
            </Text>
          </div>
          <TextField
            size="2"
            labelSize="2"
            type="text"
            disabled={!state.cimdRefreshTokenIdleTimeoutEnabled}
            label={
              <FormattedMessage id="DynamicClientsTab.refresh-token-idle-timeout.label" />
            }
            parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
            fieldName="refresh_token_idle_timeout_seconds"
            placeholder={effectiveClientConfig?.refresh_token_idle_timeout_seconds?.toFixed(
              0
            )}
            value={state.cimdRefreshTokenIdleTimeoutSeconds?.toFixed(0) ?? ""}
            onChange={onRefreshTokenIdleTimeoutChange}
          />
        </SettingsSectionCard>
      ) : null}

      <ConfirmationDialog
        open={isAnyDomainConfirmationVisible}
        onOpenChange={setIsAnyDomainConfirmationVisible}
        title={<FormattedMessage id="CIMDSection.any-domain.confirm.title" />}
        description={
          <FormattedMessage id="CIMDSection.any-domain.confirm.description" />
        }
        confirmText={
          <FormattedMessage id="CIMDSection.any-domain.confirm.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmAnyDomain}
        onCancel={onCancelAnyDomain}
      />
    </>
  );
};
