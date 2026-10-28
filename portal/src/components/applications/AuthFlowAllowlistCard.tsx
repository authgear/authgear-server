import React, { useContext, useMemo } from "react";
import {
  Button,
  DropdownMenu,
  Flex,
  IconButton,
  RadioGroup,
  Text,
} from "@radix-ui/themes";
import { TrashIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import Link from "../../Link";
import ErrorRenderer from "../../ErrorRenderer";
import { useFormField } from "../../form";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { TextField } from "../v2/TextField/TextField";
import { Badge } from "../v2/Badge/Badge";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import cn from "classnames";
import {
  AuthenticationFlowsConfig,
  OAuthClientAuthenticationFlowAllowlist,
  UIAuthenticationFlowGroup,
} from "../../types";
import { DEFAULT_GROUP_NAME } from "../../graphql/portal/authflows/list/groups";
import styles from "./AuthFlowAllowlistCard.module.css";

// The dropdown's value for "no group": not a valid group name, so it never
// collides with one.
const NO_GROUP = "-default";

type SignInPages = "authui" | "custom";

// The application's Custom UI URI; null when its plan or type has none.
// undefined is Auth UI; "" is Custom UI chosen with no URI yet, which the
// server refuses on save, so the URI is required.
export interface CustomUIOption {
  uri: string | undefined;
  onChange: (uri: string | undefined) => void;
}

// What an application may run, from the project's flows and groups.
export interface AuthFlowOptions {
  groups: UIAuthenticationFlowGroup[];
  flows: AuthenticationFlowsConfig | undefined;
}

export interface AuthFlowAllowlistCardProps {
  appID: string;
  clientIndex: number;
  allowlist: OAuthClientAuthenticationFlowAllowlist | undefined;
  options: AuthFlowOptions;
  // The app has its own sign-up and login pages, which start flows itself.
  usesCustomUI: boolean;
  // undefined: no allowlist, so the app runs the default flows.
  onChange: (
    allowlist: OAuthClientAuthenticationFlowAllowlist | undefined
  ) => void;
  className?: string;
  contentClassName?: string;
}

// Binds the server's "not defined" errors on one list of the allowlist.
function ListErrors({
  clientIndex,
  list,
  messageID,
}: {
  clientIndex: number;
  list: "groups" | "flows";
  messageID: string;
}) {
  const parent = `/oauth/clients/${clientIndex}/x_authentication_flow_allowlist`;
  const field = useMemo(
    () => ({
      parentJSONPointer: new RegExp(`^${parent}$`),
      fieldName: list,
      rules: [
        makeValidationErrorCustomMessageIDRule(
          "general",
          new RegExp(`^${parent}/${list}/\\d+$`),
          messageID
        ),
      ],
    }),
    [parent, list, messageID]
  );
  const { errors } = useFormField(field);
  if (errors.length === 0) {
    return null;
  }
  return (
    <Text as="p" size="1" color="red">
      <ErrorRenderer errors={errors} />
    </Text>
  );
}

// Whether the app's sign-up and login pages are Auth UI or its own.
export function SignInPagesCard({
  clientIndex,
  customUI,
  className,
  contentClassName,
}: {
  clientIndex: number;
  customUI: CustomUIOption;
  className?: string;
  contentClassName?: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const pages: SignInPages = customUI.uri != null ? "custom" : "authui";
  // The server's error for a missing URI is a format error on "".
  const uriErrorRules = useMemo(
    () =>
      customUI.uri === ""
        ? [
            makeValidationErrorCustomMessageIDRule(
              "format",
              new RegExp(`^/oauth/clients/${clientIndex}/x_custom_ui_uri$`),
              "AuthFlowAllowlistCard.pages.custom.empty"
            ),
          ]
        : undefined,
    [customUI.uri, clientIndex]
  );
  const option = (value: SignInPages, titleID: string, descID: string) => (
    <Text as="label" size="2">
      <Flex gap="2" align="start">
        <RadioGroup.Item value={value} className={styles.radio} />
        <span className={styles.optionText}>
          <Text as="span" size="2">
            {renderToString(titleID)}
          </Text>
          <Text as="span" size="1" color="gray">
            <FormattedMessage
              id={descID}
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                code: (chunks: React.ReactNode) => <code>{chunks}</code>,
              }}
            />
          </Text>
        </span>
      </Flex>
    </Text>
  );
  return (
    <SettingsSectionCard
      className={className}
      contentClassName={contentClassName}
      title={<FormattedMessage id="AuthFlowAllowlistCard.pages" />}
    >
      <RadioGroup.Root
        value={pages}
        onValueChange={(v) =>
          customUI.onChange((v as SignInPages) === "authui" ? undefined : "")
        }
      >
        <Flex direction="column" gap="3">
          {option(
            "authui",
            "AuthFlowAllowlistCard.pages.authui",
            "AuthFlowAllowlistCard.pages.authui.description"
          )}
          {option(
            "custom",
            "AuthFlowAllowlistCard.pages.custom",
            "AuthFlowAllowlistCard.pages.custom.description"
          )}
        </Flex>
      </RadioGroup.Root>
      {pages === "custom" ? (
        <div className={styles.indent}>
          <TextField
            size="2"
            parentJSONPointer={`/oauth/clients/${clientIndex}`}
            fieldName="x_custom_ui_uri"
            label={renderToString("EditOAuthClientForm.custom-ui-uri.label")}
            required={true}
            errorRules={uriErrorRules}
            hint={renderToString(
              "EditOAuthClientForm.custom-ui-uri.description"
            )}
            value={customUI.uri ?? ""}
            onChange={(e) => customUI.onChange(e.currentTarget.value)}
          />
        </div>
      ) : null}
    </SettingsSectionCard>
  );
}

type FlowsMode = "default" | "group";

// Which flows the app uses: the default flows, or an ordered list of flow
// groups, the top one being what Auth UI runs. With Custom UI, the same choice limits
// the flows the app's own pages may start. Individual flows stay in
// authgear.yaml: a group already unlocks its flows for Custom UI.
export function AuthFlowAllowlistCard({
  appID,
  clientIndex,
  allowlist,
  options,
  usesCustomUI,
  onChange,
  className,
  contentClassName,
}: AuthFlowAllowlistCardProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const groups = (allowlist?.groups ?? []).map((g) => g.name);
  const flows = allowlist?.flows ?? [];
  const [primary = NO_GROUP] = groups;
  const groupNames = options.groups.map((g) => g.name);
  const mode: FlowsMode = primary === NO_GROUP ? "default" : "group";

  const set = (nextGroups: string[]) => {
    // Auth UI checks the flows it runs against the allowlist too. With
    // individual flows but no group, the default flows would be refused, so
    // the default group is named instead.
    const named =
      nextGroups.length === 0 && flows.length > 0
        ? [DEFAULT_GROUP_NAME]
        : nextGroups;
    onChange(
      named.length === 0 && flows.length === 0
        ? undefined
        : {
            ...(named.length > 0
              ? { groups: named.map((name) => ({ name })) }
              : {}),
            ...(flows.length > 0 ? { flows } : {}),
          }
    );
  };

  const label = (name: string): React.ReactNode =>
    name === DEFAULT_GROUP_NAME && !groupNames.includes(name)
      ? renderToString("AuthFlowAllowlistCard.default-flows")
      : name;
  const known = (name: string) =>
    name === DEFAULT_GROUP_NAME || groupNames.includes(name);
  const addable = [
    ...(groupNames.includes(DEFAULT_GROUP_NAME) ? [] : [DEFAULT_GROUP_NAME]),
    ...groupNames,
  ].filter((n) => !groups.includes(n));
  const flowsPath = `/project/${encodeURIComponent(
    appID
  )}/configuration/authentication/flows`;
  const groupsPath = `/project/${encodeURIComponent(
    appID
  )}/configuration/authentication/flows?tab=groups`;
  const linkTo =
    (to: string) =>
    // eslint-disable-next-line react/no-unstable-nested-components
    (chunks: React.ReactNode) =>
      <Link to={to}>{chunks}</Link>;
  // What being primary means, on its badge and on Make primary.
  const primaryTip = (
    <FormattedMessage
      id="AuthFlowAllowlistCard.group.primary.tooltip"
      values={{
        // eslint-disable-next-line react/no-unstable-nested-components
        code: (chunks: React.ReactNode) => <code>{chunks}</code>,
      }}
    />
  );
  const option = (
    value: FlowsMode,
    titleID: string,
    descID: string,
    disabled: boolean
  ) => (
    <Text as="label" size="2">
      <Flex gap="2" align="start">
        <RadioGroup.Item
          value={value}
          disabled={disabled}
          className={styles.radio}
        />
        <span className={styles.optionText}>
          <Text as="span" size="2">
            {renderToString(titleID)}
          </Text>
          <Text as="span" size="1" color="gray">
            <FormattedMessage
              id={descID}
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                code: (chunks: React.ReactNode) => <code>{chunks}</code>,
                flows: linkTo(flowsPath),
                groups: linkTo(groupsPath),
              }}
            />
          </Text>
        </span>
      </Flex>
    </Text>
  );

  return (
    <SettingsSectionCard
      className={className}
      contentClassName={contentClassName}
      title={<FormattedMessage id="AuthFlowAllowlistCard.title" />}
    >
      <RadioGroup.Root
        value={mode}
        onValueChange={(v) =>
          (v as FlowsMode) === "default"
            ? set([])
            : groupNames.length > 0
            ? set([groupNames[0]])
            : undefined
        }
      >
        <Flex direction="column" gap="3">
          {option(
            "default",
            usesCustomUI
              ? "AuthFlowAllowlistCard.mode.default.custom"
              : "AuthFlowAllowlistCard.mode.default",
            usesCustomUI
              ? "AuthFlowAllowlistCard.mode.default.description.custom"
              : "AuthFlowAllowlistCard.mode.default.description",
            false
          )}
          {option(
            "group",
            "AuthFlowAllowlistCard.mode.group",
            usesCustomUI
              ? "AuthFlowAllowlistCard.mode.group.description.custom"
              : "AuthFlowAllowlistCard.mode.group.description",
            groupNames.length === 0 && mode === "default"
          )}
        </Flex>
      </RadioGroup.Root>
      {groupNames.length === 0 && mode === "default" ? (
        <Text as="p" size="1" color="gray" className={styles.indent}>
          <FormattedMessage
            id="AuthFlowAllowlistCard.mode.group.none"
            values={{ link: linkTo(groupsPath) }}
          />
        </Text>
      ) : null}
      {mode === "group" ? (
        <div className={cn(styles.field, styles.indent)}>
          {/* The top group is the one Auth UI runs; the others are what the
              app may ask for with x_authentication_flow_group. */}
          <div className={styles.groupList}>
            {groups.map((name, i) => (
              <div key={name} className={styles.groupRow}>
                <div className={styles.rowName}>
                  <Text size="2">{label(name)}</Text>
                  {!known(name) ? (
                    <Text size="1" color="red">
                      {renderToString("AuthFlowAllowlistCard.group.unknown")}
                    </Text>
                  ) : null}
                </div>
                {i === 0 ? (
                  <Tooltip content={primaryTip}>
                    <span>
                      <Badge
                        size="1"
                        variant="info"
                        text={renderToString(
                          "AuthFlowAllowlistCard.group.default"
                        )}
                      />
                    </span>
                  </Tooltip>
                ) : (
                  <Tooltip content={primaryTip}>
                    <Button
                      type="button"
                      size="1"
                      variant="ghost"
                      onClick={() =>
                        set([name, ...groups.filter((g) => g !== name)])
                      }
                    >
                      {renderToString(
                        "AuthFlowAllowlistCard.group.make-default"
                      )}
                    </Button>
                  </Tooltip>
                )}
                {/* The last group stays: Standard flows is the radio above. */}
                <IconButton
                  type="button"
                  size="1"
                  variant="ghost"
                  color="gray"
                  disabled={groups.length <= 1}
                  aria-label={renderToString("AuthFlowAllowlistCard.remove")}
                  onClick={() => set(groups.filter((g) => g !== name))}
                >
                  <TrashIcon />
                </IconButton>
              </div>
            ))}
            {addable.length > 0 ? (
              <div className={styles.groupRow}>
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger>
                    <Button type="button" size="1" variant="ghost">
                      {renderToString("AuthFlowAllowlistCard.add-group")}
                    </Button>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content size="2">
                    {addable.map((name) => (
                      <DropdownMenu.Item
                        key={name}
                        onSelect={() => set([...groups, name])}
                      >
                        {label(name)}
                      </DropdownMenu.Item>
                    ))}
                  </DropdownMenu.Content>
                </DropdownMenu.Root>
              </div>
            ) : null}
          </div>
          <ListErrors
            clientIndex={clientIndex}
            list="groups"
            messageID="AuthFlowAllowlistCard.error.unknown-group"
          />
        </div>
      ) : null}
      {flows.length > 0 ? (
        <Text as="p" size="1" color="gray">
          {renderToString("AuthFlowAllowlistCard.yaml-flows", {
            count: flows.length,
          })}
        </Text>
      ) : null}
      <ListErrors
        clientIndex={clientIndex}
        list="flows"
        messageID="AuthFlowAllowlistCard.error.unknown-flow"
      />
    </SettingsSectionCard>
  );
}
