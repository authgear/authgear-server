import React, { useContext, useState } from "react";
import {
  Button,
  Checkbox,
  Flex,
  IconButton,
  Select,
  Switch,
  Text,
  TextField,
} from "@radix-ui/themes";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ChevronRightIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../../../intl";
import Link from "../../../../Link";
import { Callout } from "../../../../components/v2/Callout/Callout";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
  AuthenticatorPhoneOTPMode,
  CustomAttributesAttributeConfig,
} from "../../../../types";
import {
  ALL_ACCOUNT_RECOVERY_DESTINATIONS,
  AccountRecoveryDestination,
  DEFAULT_FLOW_NAME,
  FlowType,
} from "../model";
import { Tooltip } from "../../../../components/v2/Tooltip/Tooltip";
import { destinationMessageID, optionLabel, stepTitle } from "../labels";
import { formFieldLabel } from "../sketches";
import { EditResult, addOption, moveItem, updateStep } from "./edits";
import {
  Contact,
  TargetUsage,
  codeContact,
  duplicateNames,
  ensureNamed,
  referencesTo,
  renameStep,
  targetCandidates,
} from "./references";
import {
  isLoginOnlyIdentification,
  allowedAuthentications,
  AuthenticationKind,
  authenticationKind,
  sendsCodes,
  allowedIdentifications,
  canHaveOptionForm,
  newOption,
} from "./rules";
import { getAt } from "./pointer";
import { StepCheck, neverFillable } from "./checks";
import { StepProblem } from "./FlowEditorContext";
import styles from "./StepForm.module.css";

type Step = AuthenticationFlowStepConfig;
type Option = AuthenticationFlowOptionConfig;
type Form = NonNullable<Step["form"]>;
type FormField = NonNullable<Form["fields"]>[number];

const NAME_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_]*$/;
// The schema reserves "default" as a form name.
const RESERVED_FORM_NAME = "default";
// Not a valid step name, so it cannot collide with one.
const TARGET_NONE = "-none";
// pkg/lib/config GetOnFailure treats an unset on_failure as ignore.
const ON_FAILURE_DEFAULT = "ignore";

const onFailureMessageIDs: Record<string, string> = {
  error: "AuthFlowsScreen.edit.form.on-failure.error",
  ignore: "AuthFlowsScreen.edit.form.on-failure.ignore",
};

export interface StepFormProps {
  flowType: FlowType;
  flow: AuthenticationFlowConfig;
  stepID: string;
  apply: (result: EditResult, notice?: string) => void;
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined;
  appID: string;
  customAttributes: CustomAttributesAttributeConfig[];
  // Attributes a form can ask for, and every attribute of the project.
  fillablePointers: readonly string[];
  knownPointers: readonly string[];
  loginFlowNames: string[];
  signupFlowNames: string[];
  // Removes an option, after confirming when steps are nested under it.
  requestRemoveOption: (stepID: string, index: number) => void;
  // Why the option cannot be removed, as rendered text; null when it can.
  removeOptionBlockedReason: (stepID: string, index: number) => string | null;
  // Save errors and checks on the step, but not those a form field shows.
  problems: StepProblem[];
  // The editor's checks on the step, for its form fields.
  checks: StepCheck[];
  notice: string | undefined;
}

// A field's label; hovering it shows the config key it writes.
function FieldLabel({
  label,
  configKey,
}: {
  label: string;
  configKey: string;
}) {
  return (
    <Tooltip content={<code>{configKey}</code>}>
      <Text as="div" size="2" weight="medium" className={styles.label}>
        {label}
      </Text>
    </Tooltip>
  );
}

function Field({
  label,
  configKey,
  children,
  hint,
}: {
  label: string;
  configKey: string;
  children: React.ReactNode;
  hint?: React.ReactNode;
}) {
  return (
    <div className={styles.field}>
      <FieldLabel label={label} configKey={configKey} />
      {children}
      {hint}
    </div>
  );
}

function Hint({
  children,
  tone = "gray",
}: {
  children: React.ReactNode;
  tone?: "gray" | "red" | "amber";
}) {
  return (
    <Text as="p" size="1" color={tone}>
      {children}
    </Text>
  );
}

function NameField({
  flow,
  stepID,
  apply,
}: Pick<StepFormProps, "flow" | "stepID" | "apply">) {
  const { renderToString } = useContext(Context);
  const step = getAt(flow, stepID) as Step;
  // Local, so an invalid name can be typed through without reaching config.
  const [text, setText] = useState(step.name ?? "");
  const [shared, setShared] = useState(false);
  const targeted = referencesTo(flow, stepID).length > 0;
  const invalid = text !== "" && !NAME_PATTERN.test(text);
  const refused = text === "" && targeted;
  const duplicate =
    step.name != null && duplicateNames(flow).has(step.name) && !invalid;

  // Renamed on commit, not per keystroke: a name typed on the way could
  // match another step, and the references would follow it there.
  const commit = () => {
    if (text === (step.name ?? "") || invalid || refused) {
      return;
    }
    const result = renameStep(flow, stepID, text);
    setShared(targeted && !result.rewritten && text !== "");
    apply(result);
  };

  return (
    <Field
      label={renderToString("AuthFlowsScreen.edit.form.name")}
      configKey="name"
      hint={
        <>
          {invalid ? (
            <Hint tone="red">
              {renderToString("AuthFlowsScreen.edit.form.name.invalid")}
            </Hint>
          ) : null}
          {refused ? (
            <Hint tone="red">
              {renderToString("AuthFlowsScreen.edit.form.name.targeted")}
            </Hint>
          ) : null}
          {duplicate ? (
            <Hint tone="amber">
              {renderToString("AuthFlowsScreen.edit.form.name.duplicate")}
            </Hint>
          ) : null}
          {shared ? (
            <Hint>
              {renderToString("AuthFlowsScreen.edit.form.name.shared")}
            </Hint>
          ) : null}
        </>
      }
    >
      <TextField.Root
        size="2"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            commit();
          }
        }}
      />
    </Field>
  );
}

// When a code option names no step, it sends to what the user types.
const targetNoneMessageIDs: Record<Contact, string> = {
  email: "AuthFlowsScreen.edit.form.target.none.email",
  phone: "AuthFlowsScreen.edit.form.target.none.phone",
};

// The target_step select: earlier steps the runtime accepts, by name. An
// unnamed candidate is named when picked.
function TargetStepSelect({
  flow,
  stepID,
  usage,
  contact = null,
  value,
  required,
  onChange,
}: {
  flow: AuthenticationFlowConfig;
  stepID: string;
  usage: TargetUsage;
  // For a code option: only steps that collected this contact on the path.
  contact?: Contact | null;
  value: string | undefined;
  required: boolean;
  // A null name clears target_step.
  onChange: (flow: AuthenticationFlowConfig, name: string | null) => void;
}) {
  const { renderToString } = useContext(Context);
  const candidates = targetCandidates(flow, stepID, usage, contact);
  const known = candidates.some((c) => c.name != null && c.name === value);
  // A step that exists but collected another contact on this path.
  const unusable =
    value != null &&
    !known &&
    targetCandidates(flow, stepID, usage).some((c) => c.name === value);
  const onValueChange = (v: string) => {
    if (v === TARGET_NONE) {
      onChange(flow, null);
      return;
    }
    if (v.startsWith("/")) {
      const target = getAt(flow, v) as Step;
      const named = ensureNamed(flow, v, target.type);
      onChange(named.flow, named.name);
      return;
    }
    onChange(flow, v);
  };
  return (
    <Select.Root
      size="1"
      value={value ?? TARGET_NONE}
      onValueChange={onValueChange}
    >
      <Select.Trigger className={styles.cellControl} />
      <Select.Content>
        {!required ? (
          <Select.Item value={TARGET_NONE}>
            {renderToString(
              contact != null
                ? targetNoneMessageIDs[contact]
                : "AuthFlowsScreen.edit.form.target.none"
            )}
          </Select.Item>
        ) : null}
        {value != null && !known ? (
          <Select.Item value={value}>
            {renderToString(
              unusable
                ? "AuthFlowsScreen.edit.form.target.unusable"
                : "AuthFlowsScreen.edit.form.target.missing",
              { name: value }
            )}
          </Select.Item>
        ) : null}
        {candidates.map((c) =>
          c.name != null ? (
            <Select.Item key={c.pointer} value={c.name}>
              {c.name}
            </Select.Item>
          ) : (
            <Select.Item key={c.pointer} value={c.pointer}>
              {renderToString("AuthFlowsScreen.edit.form.target.unnamed", {
                title: stepTitle(
                  (getAt(flow, c.pointer) as Step).type,
                  renderToString
                ),
              })}
            </Select.Item>
          )
        )}
      </Select.Content>
    </Select.Root>
  );
}

function FlowNameSelect({
  names,
  value,
  onChange,
}: {
  names: string[];
  value: string | undefined;
  onChange: (name: string) => void;
}) {
  const all =
    value != null && !names.includes(value) ? [...names, value] : names;
  return (
    <Select.Root size="1" value={value ?? ""} onValueChange={onChange}>
      <Select.Trigger className={styles.cellControl} />
      <Select.Content>
        {all.map((n) => (
          <Select.Item key={n} value={n}>
            {n}
          </Select.Item>
        ))}
      </Select.Content>
    </Select.Root>
  );
}

// A column of the options table, beside the option itself; its key is the
// config key it writes.
type OptionColumnKey =
  | "login_flow"
  | "signup_flow"
  | "on_failure"
  | "verification_required";

const optionColumnMessageIDs: Record<OptionColumnKey, string> = {
  login_flow: "AuthFlowsScreen.edit.form.column.has-account",
  signup_flow: "AuthFlowsScreen.edit.form.column.no-account",
  on_failure: "AuthFlowsScreen.edit.form.column.on-failure",
  verification_required: "AuthFlowsScreen.edit.form.column.verify",
};

function NotApplicable() {
  return <span className={styles.notApplicable}>—</span>;
}

// One row per option, with each setting as a column, so labels appear once.
const authenticationKindMessageIDs: Record<AuthenticationKind, string> = {
  primary: "AuthFlowsScreen.edit.form.authentication-kind.primary",
  secondary: "AuthFlowsScreen.edit.form.authentication-kind.secondary",
};

// The identifications or authenticators an option can take. Authenticators
// are grouped into primary and 2FA, since both have a Password and a Code
// sent to email; a 2FA one is tagged, which the closed select shows too.
function OptionValueItems({
  isIdentify,
  values,
  phoneOTPMode,
}: {
  isIdentify: boolean;
  values: string[];
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  if (isIdentify) {
    return (
      <>
        {values.map((v) => (
          <Select.Item key={v} value={v}>
            {optionLabel({ identification: v }, renderToString, phoneOTPMode)}
          </Select.Item>
        ))}
      </>
    );
  }
  const kinds: AuthenticationKind[] = ["primary", "secondary"];
  return (
    <>
      {kinds.map((kind, k) => {
        const group = values.filter((v) => authenticationKind(v) === kind);
        if (group.length === 0) {
          return null;
        }
        return (
          <React.Fragment key={kind}>
            {k > 0 ? <Select.Separator /> : null}
            <Select.Group>
              <Select.Label>
                {renderToString(authenticationKindMessageIDs[kind])}
              </Select.Label>
              {group.map((v) => (
                <Select.Item key={v} value={v}>
                  {optionLabel(
                    { authentication: v },
                    renderToString,
                    phoneOTPMode
                  )}
                  {kind === "secondary" ? (
                    <span className={styles.kindTag}>
                      {renderToString(authenticationKindMessageIDs.secondary)}
                    </span>
                  ) : null}
                </Select.Item>
              ))}
            </Select.Group>
          </React.Fragment>
        );
      })}
    </>
  );
}

function OptionsField(props: StepFormProps) {
  const { renderToString } = useContext(Context);
  const { flowType, flow, stepID, apply, phoneOTPMode } = props;
  const step = getAt(flow, stepID) as Step;
  const options = step.one_of ?? [];
  const isIdentify = step.type === "identify";
  const values = isIdentify
    ? allowedIdentifications(flowType)
    : allowedAuthentications(flowType, step.type);
  const nextOption = newOption(flowType, step);
  const valueOf = (o: Option) =>
    isIdentify ? o.identification : o.authentication;

  const updateOption = (
    j: number,
    recipe: (o: Option) => void,
    base: AuthenticationFlowConfig = flow
  ) =>
    apply(
      updateStep(base, stepID, (s) => {
        const o = s.one_of?.[j];
        if (o != null) {
          recipe(o);
        }
      })
    );

  const setValue = (j: number, v: string) =>
    updateOption(j, (o) => {
      if (!isIdentify) {
        o.authentication = v;
        // Only code options send to a step's email or phone, or verify it.
        if (!sendsCodes(v)) {
          delete o.verification_required;
          delete o.target_step;
        }
        return;
      }
      o.identification = v;
      if (!canHaveOptionForm(flowType, v)) {
        delete o.form;
      }
      // Options that only log in take no signup_flow; every other
      // signup_login option needs both flows.
      if (flowType === "signup_login") {
        if (isLoginOnlyIdentification(v)) {
          delete o.signup_flow;
        } else {
          o.signup_flow ??= DEFAULT_FLOW_NAME;
        }
      }
    });

  const columns: OptionColumnKey[] = [];
  if (flowType === "signup_login") {
    columns.push("login_flow", "signup_flow");
  }
  if (flowType === "account_recovery") {
    columns.push("on_failure");
  }
  // Where an option's codes go is an advanced setting, under the table.
  const hasSendCodesTo =
    (step.type === "authenticate" || step.type === "create_authenticator") &&
    flowType !== "reauth" &&
    options.some((o) => sendsCodes(o.authentication));
  if (step.type === "create_authenticator") {
    columns.push("verification_required");
  }

  const renderCell = (
    column: OptionColumnKey,
    o: Option,
    j: number
  ): React.ReactNode => {
    switch (column) {
      case "login_flow":
        return (
          <FlowNameSelect
            names={props.loginFlowNames}
            value={o.login_flow}
            onChange={(v) =>
              updateOption(j, (d) => {
                d.login_flow = v;
              })
            }
          />
        );
      case "signup_flow":
        // Options that always find an account never sign up.
        return isLoginOnlyIdentification(o.identification) ? (
          <NotApplicable />
        ) : (
          <FlowNameSelect
            names={props.signupFlowNames}
            value={o.signup_flow}
            onChange={(v) =>
              updateOption(j, (d) => {
                d.signup_flow = v;
              })
            }
          />
        );
      case "on_failure":
        return (
          <Select.Root
            size="1"
            value={o.on_failure ?? ON_FAILURE_DEFAULT}
            onValueChange={(v) =>
              updateOption(j, (d) => {
                d.on_failure = v;
              })
            }
          >
            <Select.Trigger className={styles.cellControl} />
            <Select.Content>
              {Object.entries(onFailureMessageIDs).map(([v, id]) => (
                <Select.Item key={v} value={v}>
                  {renderToString(id)}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        );
      case "verification_required":
        // Only codes sent to an email or phone verify it.
        return sendsCodes(o.authentication) ? (
          <Checkbox
            size="1"
            checked={o.verification_required ?? true}
            onCheckedChange={(checked) =>
              updateOption(j, (d) => {
                d.verification_required = checked === true;
              })
            }
          />
        ) : (
          <NotApplicable />
        );
    }
  };

  return (
    <div className={styles.options}>
      <FieldLabel
        label={renderToString("AuthFlowsScreen.edit.form.options")}
        configKey="one_of"
      />
      <div
        className={styles.optionTable}
        style={{
          gridTemplateColumns: `minmax(0, 1.6fr) ${columns
            .map((c) =>
              c === "verification_required" ? "auto" : "minmax(0, 1fr)"
            )
            .join(" ")} auto`,
        }}
      >
        <span className={styles.columnHeader}>
          {renderToString("AuthFlowsScreen.edit.form.column.option")}
        </span>
        {columns.map((c) => (
          <Tooltip key={c} content={<code>{c}</code>}>
            <span className={styles.columnHeader}>
              {renderToString(optionColumnMessageIDs[c])}
            </span>
          </Tooltip>
        ))}
        <span />
        {options.map((o, j) => {
          const used = new Set(options.filter((_, k) => k !== j).map(valueOf));
          const current = valueOf(o);
          const removeBlocked = props.removeOptionBlockedReason(stepID, j);
          return (
            <React.Fragment key={j}>
              <Select.Root
                size="1"
                value={current ?? ""}
                onValueChange={(v) => setValue(j, v)}
              >
                <Select.Trigger className={styles.cellControl} />
                <Select.Content>
                  <OptionValueItems
                    isIdentify={isIdentify}
                    values={(current != null && !values.includes(current)
                      ? [...values, current]
                      : values
                    ).filter((v) => v === current || !used.has(v))}
                    phoneOTPMode={phoneOTPMode}
                  />
                </Select.Content>
              </Select.Root>
              {columns.map((c) => (
                <div key={c} className={styles.cell}>
                  {renderCell(c, o, j)}
                </div>
              ))}
              <Tooltip
                content={
                  removeBlocked ??
                  renderToString("AuthFlowsScreen.edit.form.option.remove")
                }
              >
                <span>
                  <IconButton
                    type="button"
                    size="1"
                    variant="ghost"
                    color="gray"
                    disabled={removeBlocked != null}
                    aria-label={renderToString(
                      "AuthFlowsScreen.edit.form.option.remove"
                    )}
                    onClick={() => props.requestRemoveOption(stepID, j)}
                  >
                    <TrashIcon />
                  </IconButton>
                </span>
              </Tooltip>
            </React.Fragment>
          );
        })}
      </div>
      <Button
        type="button"
        size="1"
        variant="soft"
        disabled={nextOption == null}
        onClick={() => {
          if (nextOption != null) {
            apply(addOption(flow, stepID, nextOption));
          }
        }}
      >
        {renderToString("AuthFlowsScreen.edit.add-option")}
      </Button>
      {hasSendCodesTo ? (
        <AdvancedSettings
          setCount={
            options.filter(
              (o) => sendsCodes(o.authentication) && o.target_step != null
            ).length
          }
        >
          <FieldLabel
            label={renderToString("AuthFlowsScreen.edit.form.send-codes-to")}
            configKey="target_step"
          />
          <Text as="p" size="1" color="gray">
            {renderToString(
              "AuthFlowsScreen.edit.form.send-codes-to.description"
            )}
          </Text>
          <div className={styles.advancedTable}>
            {options.map((o, j) =>
              sendsCodes(o.authentication) ? (
                <React.Fragment key={j}>
                  <Text size="2" className={styles.advancedOption}>
                    {optionLabel(o, renderToString, phoneOTPMode)}
                  </Text>
                  <TargetStepSelect
                    flow={flow}
                    stepID={stepID}
                    usage={
                      step.type === "create_authenticator"
                        ? step.type
                        : "authenticate"
                    }
                    contact={codeContact(o.authentication)}
                    value={o.target_step}
                    required={false}
                    onChange={(base, name) =>
                      updateOption(
                        j,
                        (d) => {
                          if (name == null) {
                            delete d.target_step;
                          } else {
                            d.target_step = name;
                          }
                        },
                        base
                      )
                    }
                  />
                </React.Fragment>
              ) : null
            )}
          </div>
        </AdvancedSettings>
      ) : null}
    </div>
  );
}

// Settings most flows never change, collapsed until opened. The summary
// counts the ones that are set, so a value is never out of sight entirely.
function AdvancedSettings({
  setCount,
  children,
}: {
  setCount: number;
  children: React.ReactNode;
}) {
  const { renderToString } = useContext(Context);
  const [open, setOpen] = useState(false);
  return (
    <details
      className={styles.advanced}
      open={open}
      onToggle={(e) => setOpen(e.currentTarget.open)}
    >
      <summary className={styles.advancedSummary}>
        <ChevronRightIcon
          className={styles.advancedChevron}
          aria-hidden={true}
        />
        {renderToString("AuthFlowsScreen.edit.form.advanced")}
        {setCount > 0 ? (
          <Text size="1" color="gray">
            {renderToString("AuthFlowsScreen.edit.form.advanced.set", {
              count: setCount,
            })}
          </Text>
        ) : null}
      </summary>
      <div className={styles.advancedBody}>{children}</div>
    </details>
  );
}

// Local text, so an invalid name never reaches the draft.
function FormNameField({
  value,
  onCommit,
  hint,
}: {
  value: string;
  onCommit: (name: string) => void;
  hint?: React.ReactNode;
}) {
  const { renderToString } = useContext(Context);
  const [text, setText] = useState(value);
  const invalid =
    text !== "" && (!NAME_PATTERN.test(text) || text === RESERVED_FORM_NAME);
  const commit = () => {
    if (text !== value && !invalid) {
      onCommit(text);
    }
  };
  return (
    <Field
      label={renderToString("AuthFlowsScreen.edit.form.form-name")}
      configKey="form.name"
      hint={
        invalid ? (
          <Hint tone="red">
            {renderToString("AuthFlowsScreen.edit.form.form-name.invalid")}
          </Hint>
        ) : (
          hint
        )
      }
    >
      <TextField.Root
        size="2"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            commit();
          }
        }}
      />
    </Field>
  );
}

// Why an attribute of the project can't be picked, for the field picker.
const notFillableReasonIDs = {
  identity: "AuthFlowsScreen.edit.form.field.reason.identity",
  address: "AuthFlowsScreen.edit.form.field.reason.address",
  access: "AuthFlowsScreen.edit.form.field.reason.access",
} as const;

function notFillableReason(pointer: string): keyof typeof notFillableReasonIDs {
  if (pointer.startsWith("/address/")) {
    return "address";
  }
  return neverFillable(pointer) ? "identity" : "access";
}

interface FormEditorProps
  extends Pick<
    StepFormProps,
    | "stepID"
    | "appID"
    | "customAttributes"
    | "fillablePointers"
    | "knownPointers"
    | "checks"
  > {
  form: Form | undefined;
  updateForm: (recipe: (form: Form) => void) => void;
  // The identify option the form belongs to; absent for a form step.
  option?: number;
  nameHint?: React.ReactNode;
}

// A form's name and fields, on a form step or an identify option.
function FormEditor(props: FormEditorProps) {
  const { renderToString } = useContext(Context);
  const { appID, customAttributes, form, updateForm, option } = props;
  const fields = form?.fields ?? [];
  const pointers = props.fillablePointers;
  const unavailable = props.knownPointers.filter((p) => !pointers.includes(p));
  const label = (p: string) =>
    formFieldLabel(p, customAttributes, renderToString);
  const settingsPath = (p: string) =>
    `/project/${encodeURIComponent(appID)}/configuration/user-profile/${
      customAttributes.some((a) => a.pointer === p)
        ? "custom-attributes"
        : "standard-attributes"
    }`;
  const notFillable = new Set(
    props.checks.flatMap((c) =>
      c.kind === "field_not_fillable" && c.option === option ? [c.pointer] : []
    )
  );
  return (
    <>
      <FormNameField
        key={`${props.stepID}:${option ?? ""}:${form?.name ?? ""}`}
        value={form?.name ?? ""}
        hint={props.nameHint}
        onCommit={(name) =>
          updateForm((f) => {
            if (name === "") {
              delete f.name;
            } else {
              f.name = name;
            }
          })
        }
      />
      <div className={styles.options}>
        <FieldLabel
          label={renderToString("AuthFlowsScreen.edit.form.fields")}
          configKey="form.fields"
        />
        {fields.map((f, j) => {
          const pointer = f.user_profile?.pointer ?? "";
          // A field cannot ask for what another field already asks for.
          const others = new Set(
            fields.filter((_, k) => k !== j).map((g) => g.user_profile?.pointer)
          );
          const repeated = others.has(pointer);
          return (
            <div key={j} className={styles.option}>
              <Flex gap="2" align="center">
                <Select.Root
                  size="1"
                  value={pointer}
                  onValueChange={(v) =>
                    updateForm((d) => {
                      const field = d.fields?.[j];
                      if (field != null) {
                        field.user_profile = {
                          ...field.user_profile,
                          pointer: v,
                        };
                      }
                    })
                  }
                >
                  <Select.Trigger className={styles.grow} />
                  <Select.Content>
                    {(pointers.includes(pointer) || pointer === ""
                      ? pointers
                      : [...pointers, pointer]
                    ).map((p) => (
                      <Select.Item
                        key={p}
                        value={p}
                        disabled={p === pointer ? false : others.has(p)}
                      >
                        {label(p)}
                      </Select.Item>
                    ))}
                    {unavailable.some((p) => p !== pointer) ? (
                      <>
                        <Select.Separator />
                        <Select.Group>
                          <Select.Label>
                            {renderToString(
                              "AuthFlowsScreen.edit.form.field.unavailable"
                            )}
                          </Select.Label>
                          {unavailable
                            .filter((p) => p !== pointer)
                            .map((p) => (
                              <Select.Item key={p} value={p} disabled={true}>
                                {renderToString(
                                  notFillableReasonIDs[notFillableReason(p)],
                                  { attribute: label(p) }
                                )}
                              </Select.Item>
                            ))}
                        </Select.Group>
                      </>
                    ) : null}
                  </Select.Content>
                </Select.Root>
                <IconButton
                  type="button"
                  size="1"
                  variant="ghost"
                  color="gray"
                  disabled={j === 0}
                  aria-label={renderToString("OrderButtons.move-up")}
                  onClick={() =>
                    updateForm((d) => {
                      d.fields = moveItem(d.fields ?? [], j, j - 1);
                    })
                  }
                >
                  <ArrowUpIcon />
                </IconButton>
                <IconButton
                  type="button"
                  size="1"
                  variant="ghost"
                  color="gray"
                  disabled={j === fields.length - 1}
                  aria-label={renderToString("OrderButtons.move-down")}
                  onClick={() =>
                    updateForm((d) => {
                      d.fields = moveItem(d.fields ?? [], j, j + 1);
                    })
                  }
                >
                  <ArrowDownIcon />
                </IconButton>
                <Button
                  type="button"
                  size="1"
                  variant="ghost"
                  color="red"
                  disabled={fields.length <= 1}
                  onClick={() =>
                    updateForm((d) => {
                      d.fields?.splice(j, 1);
                    })
                  }
                >
                  {renderToString("AuthFlowsScreen.edit.form.field.remove")}
                </Button>
              </Flex>
              <Text as="label" size="1">
                <Flex gap="2" align="center">
                  <Checkbox
                    size="1"
                    checked={f.user_profile?.required === true}
                    onCheckedChange={(checked) =>
                      updateForm((d) => {
                        const field = d.fields?.[j];
                        if (field?.user_profile != null) {
                          field.user_profile.required = checked === true;
                        }
                      })
                    }
                  />
                  {renderToString("AuthFlowsScreen.edit.form.field.required")}
                </Flex>
              </Text>
              {notFillable.has(pointer) ? (
                <Hint tone="red">
                  {!props.knownPointers.includes(pointer) ? (
                    <FormattedMessage
                      id="AuthFlowsScreen.check.field-unknown"
                      values={{ attribute: label(pointer) }}
                    />
                  ) : neverFillable(pointer) ? (
                    <FormattedMessage
                      id="AuthFlowsScreen.check.field-never-fillable"
                      values={{ attribute: label(pointer) }}
                    />
                  ) : (
                    <FormattedMessage
                      id="AuthFlowsScreen.check.field-not-fillable"
                      values={{
                        attribute: label(pointer),
                        // eslint-disable-next-line react/no-unstable-nested-components
                        link: (chunks: React.ReactNode) => (
                          <Link to={settingsPath(pointer)}>{chunks}</Link>
                        ),
                      }}
                    />
                  )}
                </Hint>
              ) : null}
              {repeated ? (
                <Hint tone="red">
                  {renderToString("AuthFlowsScreen.check.field-duplicate", {
                    attribute: label(pointer),
                  })}
                </Hint>
              ) : null}
            </div>
          );
        })}
        <Button
          type="button"
          size="1"
          variant="soft"
          disabled={pointers.every((p) =>
            fields.some((f) => f.user_profile?.pointer === p)
          )}
          onClick={() =>
            updateForm((d) => {
              const used = new Set(
                (d.fields ?? []).map((f) => f.user_profile?.pointer)
              );
              const next = pointers.find((p) => !used.has(p));
              if (next != null) {
                d.fields = [...(d.fields ?? []), newField(next)];
              }
            })
          }
        >
          {renderToString("AuthFlowsScreen.edit.form.field.add")}
        </Button>
        {unavailable.some((p) => notFillableReason(p) === "access") ? (
          <Hint>
            <FormattedMessage
              id="AuthFlowsScreen.edit.form.field.unavailable.note"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                standard: (chunks: React.ReactNode) => (
                  <Link
                    to={`/project/${encodeURIComponent(
                      appID
                    )}/configuration/user-profile/standard-attributes`}
                  >
                    {chunks}
                  </Link>
                ),
                // eslint-disable-next-line react/no-unstable-nested-components
                custom: (chunks: React.ReactNode) => (
                  <Link
                    to={`/project/${encodeURIComponent(
                      appID
                    )}/configuration/user-profile/custom-attributes`}
                  >
                    {chunks}
                  </Link>
                ),
              }}
            />
          </Hint>
        ) : null}
      </div>
    </>
  );
}

function newField(pointer: string): FormField {
  return { user_profile: { pointer } };
}

function FormFieldsField(props: StepFormProps) {
  const { flow, stepID, apply } = props;
  const step = getAt(flow, stepID) as Step;
  return (
    <FormEditor
      {...props}
      form={step.form}
      updateForm={(recipe) =>
        apply(
          updateStep(flow, stepID, (s) => {
            s.form ??= {};
            recipe(s.form);
          })
        )
      }
      nameHint={
        <Hint>
          <FormattedMessage
            id="AuthFlowsScreen.edit.form.form-name.text-keys"
            values={{
              name: step.form?.name ?? RESERVED_FORM_NAME,
              // eslint-disable-next-line react/no-unstable-nested-components
              code: (chunks: React.ReactNode) => <code>{chunks}</code>,
            }}
          />
        </Hint>
      }
    />
  );
}

// Fields an identify option asks for on the same page as its login ID.
function OptionForms(props: StepFormProps) {
  const { renderToString } = useContext(Context);
  const { flowType, flow, stepID, apply, phoneOTPMode } = props;
  const step = getAt(flow, stepID) as Step;
  const options = step.one_of ?? [];
  if (!options.some((o) => canHaveOptionForm(flowType, o.identification))) {
    return null;
  }
  const updateOption = (j: number, recipe: (o: Option) => void) =>
    apply(
      updateStep(flow, stepID, (s) => {
        const o = s.one_of?.[j];
        if (o != null) {
          recipe(o);
        }
      })
    );
  return (
    <div className={styles.options}>
      <FieldLabel
        label={renderToString("AuthFlowsScreen.edit.form.option-forms")}
        configKey="one_of.form"
      />
      <Hint>
        {renderToString("AuthFlowsScreen.edit.form.option-forms.description")}
      </Hint>
      {options.map((o, j) => {
        if (!canHaveOptionForm(flowType, o.identification)) {
          return null;
        }
        const name = optionLabel(o, renderToString, phoneOTPMode);
        if (o.form == null) {
          return (
            <Button
              key={j}
              type="button"
              size="1"
              variant="soft"
              disabled={props.fillablePointers.length === 0}
              onClick={() =>
                updateOption(j, (d) => {
                  d.form = { fields: [newField(props.fillablePointers[0])] };
                })
              }
            >
              {renderToString("AuthFlowsScreen.edit.form.option-forms.add", {
                option: name,
              })}
            </Button>
          );
        }
        return (
          <div key={j} className={styles.option}>
            <Flex gap="2" align="center" justify="between">
              <Text size="2" weight="medium">
                {name}
              </Text>
              <Button
                type="button"
                size="1"
                variant="ghost"
                color="red"
                onClick={() =>
                  updateOption(j, (d) => {
                    delete d.form;
                  })
                }
              >
                {renderToString(
                  "AuthFlowsScreen.edit.form.option-forms.remove"
                )}
              </Button>
            </Flex>
            <FormEditor
              {...props}
              form={o.form}
              option={j}
              updateForm={(recipe) =>
                updateOption(j, (d) => {
                  d.form ??= {};
                  recipe(d.form);
                })
              }
            />
          </div>
        );
      })}
    </div>
  );
}

function sameDestination(
  a: { channel: string; otp_form?: string },
  b: AccountRecoveryDestination
) {
  return a.channel === b.channel && a.otp_form === b.otp_form;
}

function SelectDestinationFields(props: StepFormProps) {
  const { renderToString } = useContext(Context);
  const { flow, stepID, apply } = props;
  const step = getAt(flow, stepID) as Step;
  // No allowed_channels means every channel at runtime.
  const allowed =
    step.allowed_channels != null && step.allowed_channels.length > 0
      ? step.allowed_channels
      : ALL_ACCOUNT_RECOVERY_DESTINATIONS;
  const update = (recipe: (s: Step) => void) =>
    apply(updateStep(flow, stepID, recipe));
  return (
    <>
      <Field
        label={renderToString("AuthFlowsScreen.edit.form.enumerate")}
        configKey="enumerate_destinations"
        hint={
          <Hint>
            {renderToString("AuthFlowsScreen.edit.form.enumerate.note")}
          </Hint>
        }
      >
        <Switch
          size="1"
          checked={step.enumerate_destinations === true}
          onCheckedChange={(checked) =>
            update((s) => {
              if (checked) {
                s.enumerate_destinations = true;
              } else {
                delete s.enumerate_destinations;
              }
            })
          }
        />
      </Field>
      <Field
        label={renderToString("AuthFlowsScreen.edit.form.channels")}
        configKey="allowed_channels"
      >
        {ALL_ACCOUNT_RECOVERY_DESTINATIONS.map((d) => {
          const checked = allowed.some((a) => sameDestination(a, d));
          const id = destinationMessageID(d.channel, d.otp_form);
          return (
            <Text as="label" size="2" key={`${d.channel}.${d.otp_form}`}>
              <Flex gap="2" align="center">
                <Checkbox
                  size="1"
                  checked={checked}
                  // At least one must stay: none would mean every channel.
                  disabled={checked ? allowed.length <= 1 : false}
                  onCheckedChange={(on) =>
                    update((s) => {
                      const next = allowed.filter(
                        (a) => !sameDestination(a, d)
                      );
                      s.allowed_channels = (
                        on === true ? [...allowed, d] : next
                      ).map((a) => ({
                        channel: a.channel,
                        otp_form: a.otp_form,
                      }));
                    })
                  }
                />
                {id != null ? renderToString(id) : d.channel}
              </Flex>
            </Text>
          );
        })}
      </Field>
    </>
  );
}

export function StepForm(props: StepFormProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { flowType, flow, stepID, apply, problems, notice } = props;
  const step = getAt(flow, stepID) as Step;
  const update = (recipe: (s: Step) => void) =>
    apply(updateStep(flow, stepID, recipe));

  return (
    <div className={styles.form}>
      {problems.map((p, i) => (
        <Callout
          key={i}
          type={p.tone === "error" ? "error" : "warning"}
          size="1"
          text={p.text}
        />
      ))}
      {notice != null ? <Callout type="info" size="1" text={notice} /> : null}
      <NameField
        key={`${stepID}:${step.name ?? ""}`}
        flow={flow}
        stepID={stepID}
        apply={apply}
      />
      {step.type === "authenticate" && flowType === "login" ? (
        <Field
          label={renderToString("AuthFlowsScreen.edit.form.optional")}
          configKey="optional"
        >
          <Switch
            size="1"
            checked={step.optional === true}
            onCheckedChange={(checked) =>
              update((s) => {
                if (checked) {
                  s.optional = true;
                } else {
                  delete s.optional;
                }
              })
            }
          />
        </Field>
      ) : null}
      {step.type === "verify" || step.type === "change_password" ? (
        <Field
          label={renderToString("AuthFlowsScreen.edit.form.target-step")}
          configKey="target_step"
        >
          <TargetStepSelect
            flow={flow}
            stepID={stepID}
            usage={step.type}
            value={step.target_step}
            required={true}
            onChange={(base, name) =>
              apply(
                updateStep(base, stepID, (s) => {
                  if (name != null) {
                    s.target_step = name;
                  }
                })
              )
            }
          />
        </Field>
      ) : null}
      {step.one_of != null &&
      (step.type === "identify" ||
        step.type === "authenticate" ||
        step.type === "create_authenticator") ? (
        <OptionsField {...props} />
      ) : null}
      {step.type === "identify" ? <OptionForms {...props} /> : null}
      {step.type === "fill_form" ? <FormFieldsField {...props} /> : null}
      {step.type === "select_destination" ? (
        <SelectDestinationFields {...props} />
      ) : null}
    </div>
  );
}
