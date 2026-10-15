import React, { useContext } from "react";
import cn from "classnames";
import { Handle, Position } from "@xyflow/react";
import { ChevronDownIcon } from "@radix-ui/react-icons";
import { Context } from "../../../intl";
import { Tooltip } from "../../../components/v2/Tooltip/Tooltip";
import {
  AuthenticationFlowFormConfig,
  AuthenticationFlowOptionConfig,
  AuthenticationFlowStepConfig,
  CustomAttributesAttributeConfig,
} from "../../../types";
import { FlowStepItem, isDeviceTokenOption } from "./tree";
import {
  AccountAnswer,
  handoffGroups,
  optionAnswerHandleID,
  optionFlowRefs,
  optionHandleID,
  recoveryDestinationHandleID,
} from "./graph";
import {
  oauthProviderMessageIDs,
  optionLabel,
  phoneCodeSetupMessageID,
  destinationMessageID,
  verifyCodeMessageID,
} from "./labels";
import { useFlowSketchContext } from "./FlowSketchContext";
import { IdentifyGroupGeometry } from "./groupGeometry";
import styles from "./FlowNodes.module.css";

const OTP_BOX_COUNT = 6;
const IDENTIFICATION_SELECT_ACCOUNT = "select_account";
const RECOVERY_CODE_COUNT = 6;

function SketchInput({ label }: { label: string }) {
  return (
    <div className={styles.sketchField}>
      <span className={styles.sketchLabel}>{label}</span>
      <div className={styles.sketchInput} />
    </div>
  );
}

function SketchSelect({ label }: { label: string }) {
  return (
    <div className={styles.sketchField}>
      <span className={styles.sketchLabel}>{label}</span>
      <div className={styles.sketchInput}>
        <ChevronDownIcon className={styles.sketchChevron} />
      </div>
    </div>
  );
}

function SketchButton({ label }: { label: string }) {
  return <div className={styles.sketchButton}>{label}</div>;
}

function SketchText({ label }: { label: string }) {
  return <span className={styles.sketchText}>{label}</span>;
}

function SketchLink({ label }: { label: string }) {
  return <span className={styles.sketchLink}>{label}</span>;
}

function SketchCheckbox({ label }: { label: string }) {
  return (
    <div className={styles.sketchCheckbox}>
      <span className={styles.sketchCheckboxBox} />
      <span>{label}</span>
    </div>
  );
}

// Tabler's qrcode icon, from the icon font the portal already loads.
function SketchQRCode({ label }: { label: string }) {
  return (
    <div className={styles.sketchField}>
      <span className={styles.sketchLabel}>{label}</span>
      <span className={styles.sketchQR}>
        <i className="ti ti-qrcode" aria-hidden={true} />
      </span>
    </div>
  );
}

function SketchRadio({ label, checked }: { label: string; checked: boolean }) {
  return (
    <div className={styles.sketchRadio}>
      <span
        className={cn(
          styles.sketchRadioDot,
          checked ? styles.sketchRadioOn : null
        )}
      />
      <span>{label}</span>
    </div>
  );
}

function SketchOTP({ label }: { label: string }) {
  return (
    <div className={styles.sketchField}>
      <span className={styles.sketchLabel}>{label}</span>
      <div className={styles.sketchOTP}>
        {Array.from({ length: OTP_BOX_COUNT }, (_, i) => (
          <span key={i} className={styles.sketchOTPBox} />
        ))}
      </div>
    </div>
  );
}

type SketchControl = "text" | "select";

function fieldControl(
  pointer: string,
  custom: CustomAttributesAttributeConfig | undefined
): SketchControl {
  if (custom != null) {
    switch (custom.type) {
      case "enum":
      case "country_code":
        return "select";
      default:
        return "text";
    }
  }
  switch (pointer) {
    case "/gender":
    case "/zoneinfo":
    case "/locale":
      return "select";
    default:
      return "text";
  }
}

function formFieldLabel(
  pointer: string,
  customAttributes: CustomAttributesAttributeConfig[],
  renderToString: (id: string) => string
): string {
  const name = pointer.replace(/^\//, "");
  return customAttributes.some((a) => a.pointer === pointer)
    ? name
    : renderToString(`standard-attribute.${name}`);
}

function FormFields({ form }: { form: AuthenticationFlowFormConfig }) {
  const { renderToString } = useContext(Context);
  const { customAttributes } = useFlowSketchContext();
  const fields = form.fields ?? [];
  if (fields.length === 0) {
    return (
      <SketchLink label={renderToString("AuthFlowsScreen.sketch.no-fields")} />
    );
  }
  return (
    <>
      {fields.map((field, i) => {
        const pointer = field.user_profile?.pointer ?? "";
        const required = field.user_profile?.required === true;
        const custom = customAttributes.find((a) => a.pointer === pointer);
        const label =
          formFieldLabel(pointer, customAttributes, renderToString) +
          (required ? " *" : "");
        switch (fieldControl(pointer, custom)) {
          case "select":
            return <SketchSelect key={i} label={label} />;
          default:
            return <SketchInput key={i} label={label} />;
        }
      })}
    </>
  );
}

function legacyProfileForm(
  step: AuthenticationFlowStepConfig
): AuthenticationFlowFormConfig {
  return {
    fields: (step.user_profile ?? []).map((up) => ({ user_profile: up })),
  };
}

function IdentifyOption({
  option,
}: {
  option: AuthenticationFlowOptionConfig;
}) {
  const { renderToString } = useContext(Context);
  const { oauthProviders } = useFlowSketchContext();
  const label = optionLabel(option, renderToString);
  switch (option.identification) {
    case "email":
    case "phone":
    case "username":
      return (
        <>
          <SketchInput label={label} />
          {option.form != null ? <FormFields form={option.form} /> : null}
        </>
      );
    case "oauth":
      if (oauthProviders.length === 0) {
        return (
          <SketchLink
            label={renderToString("AuthFlowsScreen.sketch.no-oauth-provider")}
          />
        );
      }
      return (
        <>
          {oauthProviders.map((p) => (
            <SketchButton
              key={p.alias}
              label={renderToString("AuthFlowsScreen.sketch.continue-with", {
                provider: renderToString(oauthProviderMessageIDs[p.type]),
              })}
            />
          ))}
        </>
      );
    case "passkey":
      return (
        <SketchButton
          label={renderToString("AuthFlowsScreen.sketch.passkey-button")}
        />
      );
    case IDENTIFICATION_SELECT_ACCOUNT:
      return (
        <SketchButton
          label={renderToString("AuthFlowsScreen.sketch.select-account")}
        />
      );
    case "ldap":
      return (
        <>
          <SketchInput
            label={renderToString("AuthFlowsScreen.sketch.ldap-username")}
          />
          <SketchInput
            label={renderToString("AuthFlowsScreen.sketch.ldap-password")}
          />
        </>
      );
    default:
      return <SketchLink label={label} />;
  }
}

// Setting up a method is drawn as one line describing it.
const setupMessageIDs: Partial<Record<string, string>> = {
  primary_password: "AuthFlowsScreen.setup.password",
  secondary_password: "AuthFlowsScreen.setup.password",
  primary_passkey: "AuthFlowsScreen.setup.passkey",
  primary_oob_otp_email: "AuthFlowsScreen.setup.oob_otp_email",
  secondary_oob_otp_email: "AuthFlowsScreen.setup.oob_otp_email",
  secondary_totp: "AuthFlowsScreen.setup.totp",
};

// A method being set up: its one-line description as the heading, then the
// fields the user fills in. A code method without target_step asks for the
// email or phone first; with target_step the code goes to that step's
// login ID, so only the code is entered.
function CreateAuthenticatorOption({
  option,
}: {
  option: AuthenticationFlowOptionConfig;
}) {
  const { renderToString } = useContext(Context);
  const { phoneOTPMode } = useFlowSketchContext();
  const auth = option.authentication;
  const isPhoneCode =
    auth === "primary_oob_otp_sms" || auth === "secondary_oob_otp_sms";
  const id = isPhoneCode
    ? phoneCodeSetupMessageID(phoneOTPMode)
    : auth != null
    ? setupMessageIDs[auth]
    : undefined;
  const heading =
    id != null ? renderToString(id) : optionLabel(option, renderToString);
  const codeBoxes = (
    <SketchOTP label={renderToString("AuthFlowsScreen.sketch.enter-code")} />
  );
  const askFirst = (messageID: string) =>
    option.target_step == null ? (
      <SketchInput label={renderToString(messageID)} />
    ) : null;
  let fields: React.ReactNode;
  switch (auth) {
    case "primary_password":
    case "secondary_password":
      fields = (
        <SketchInput
          label={renderToString("AuthFlowsScreen.sketch.new-password")}
        />
      );
      break;
    case "primary_passkey":
      fields = (
        <SketchButton
          label={renderToString("AuthFlowsScreen.sketch.create-passkey")}
        />
      );
      break;
    case "primary_oob_otp_email":
    case "secondary_oob_otp_email":
      fields = (
        <>
          {askFirst("AuthFlowsScreen.identification.email")}
          {codeBoxes}
        </>
      );
      break;
    case "primary_oob_otp_sms":
    case "secondary_oob_otp_sms":
      fields = (
        <>
          {askFirst("AuthFlowsScreen.sketch.phone-number")}
          {codeBoxes}
        </>
      );
      break;
    case "secondary_totp":
      fields = (
        <>
          <SketchQRCode
            label={renderToString("AuthFlowsScreen.sketch.scan-qr")}
          />
          <SketchOTP
            label={renderToString("AuthFlowsScreen.sketch.enter-app-code")}
          />
        </>
      );
      break;
    default:
      fields = null;
  }
  return (
    <>
      <span className={styles.sketchOption}>{heading}</span>
      {fields}
    </>
  );
}

function AuthenticatorOption({
  option,
}: {
  option: AuthenticationFlowOptionConfig;
}) {
  const { renderToString } = useContext(Context);
  const { phoneOTPMode } = useFlowSketchContext();
  const label = optionLabel(option, renderToString, phoneOTPMode);
  switch (option.authentication) {
    case "primary_password":
    case "secondary_password":
      return <SketchInput label={label} />;
    case "primary_passkey":
      return (
        <SketchButton
          label={renderToString("AuthFlowsScreen.sketch.passkey-button")}
        />
      );
    case "primary_oob_otp_email":
    case "secondary_oob_otp_email":
    case "primary_oob_otp_sms":
    case "secondary_oob_otp_sms":
    case "secondary_totp":
      return <SketchOTP label={label} />;
    case "device_token":
      return <SketchCheckbox label={label} />;
    default:
      return <SketchLink label={label} />;
  }
}

// The two outputs of an option that can lead to both flows, above and below
// the row's middle.
const ACCOUNT_ANSWERS: {
  answer: AccountAnswer;
  top: string;
  messageID: string;
}[] = [
  {
    answer: "has_account",
    top: "calc(50% - 10px)",
    messageID: "AuthFlowsScreen.answer.has_account",
  },
  {
    answer: "no_account",
    top: "calc(50% + 10px)",
    messageID: "AuthFlowsScreen.answer.no_account",
  },
];

function OptionHandle({
  item,
  optionIndex,
  className,
}: {
  item: FlowStepItem;
  optionIndex: number;
  className: string;
}) {
  const { renderToString } = useContext(Context);
  if (item.branches == null) {
    return null;
  }
  const option = item.step.one_of?.[optionIndex];
  // An option that can lead to both flows answers whether the account
  // exists, with one labelled output per answer.
  if (option != null && optionFlowRefs(option).length > 1) {
    return (
      <>
        {ACCOUNT_ANSWERS.map(({ answer, top, messageID }) => (
          <Handle
            key={answer}
            type="source"
            position={Position.Right}
            id={optionAnswerHandleID(optionIndex, answer)}
            className={className}
            style={{ top }}
          >
            <span className={styles.portLabel}>
              {renderToString(messageID)}
            </span>
          </Handle>
        ))}
      </>
    );
  }
  return (
    <Handle
      type="source"
      position={Position.Right}
      id={optionHandleID(optionIndex)}
      className={className}
    />
  );
}

// One section per destination the flow sends to: AuthUI shows a "link sent"
// page for a link and code boxes for a code. Several are drawn as sub-nodes.
function RecoveryCodeFields() {
  const { renderToString } = useContext(Context);
  const { recoveryDestinations } = useFlowSketchContext();
  const codeBoxes = (
    <SketchOTP label={renderToString("AuthFlowsScreen.sketch.enter-code")} />
  );
  if (recoveryDestinations.length === 0) {
    return codeBoxes;
  }
  const subnode = recoveryDestinations.length > 1;
  return (
    <>
      {recoveryDestinations.map((d) => {
        const heading = destinationMessageID(d.channel, d.otp_form);
        return (
          <div
            key={`${d.channel}.${d.otp_form}`}
            className={cn(
              styles.optionRow,
              subnode ? styles.optionSubnode : null
            )}
          >
            <Handle
              type="target"
              position={Position.Left}
              id={recoveryDestinationHandleID(d)}
              className={styles.destinationHandle}
            />
            <span className={styles.sketchOption}>
              {heading != null ? renderToString(heading) : d.channel}
            </span>
            {d.otp_form === "link" ? (
              <>
                <SketchText
                  label={renderToString("AuthFlowsScreen.recovery.link-sent")}
                />
                <SketchLink
                  label={renderToString("AuthFlowsScreen.recovery.resend")}
                />
              </>
            ) : (
              codeBoxes
            )}
          </div>
        );
      })}
    </>
  );
}

function OptionRow({
  item,
  optionIndex,
  subnode,
  children,
}: {
  item: FlowStepItem;
  optionIndex: number;
  // One of several alternatives: drawn as its own tinted sub-node.
  subnode: boolean;
  children: React.ReactNode;
}) {
  return (
    <div
      className={cn(styles.optionRow, subnode ? styles.optionSubnode : null)}
    >
      {children}
      <OptionHandle
        item={item}
        optionIndex={optionIndex}
        className={styles.optionHandle}
      />
    </div>
  );
}

// The index of the option AuthUI shows on a page of its own before the login
// page, or -1. Only an identify step with other options is drawn as a group.
export function continuePageOptionIndex(item: FlowStepItem): number {
  const options = item.step.one_of ?? [];
  if (item.step.type !== "identify" || options.length < 2) {
    return -1;
  }
  return options.findIndex(
    (o) => o.identification === IDENTIFICATION_SELECT_ACCOUNT
  );
}

// The "Continue as" page AuthUI shows before the login page when the user
// already has a session. Rows have fixed heights from the group geometry.
export function ContinueAsPage({
  geometry,
}: {
  geometry: IdentifyGroupGeometry;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { rows } = geometry;
  return (
    <div
      className={styles.continuePage}
      style={{
        left: geometry.continueLeft,
        top: geometry.continueTop,
        width: geometry.continueWidth,
        height: geometry.continueHeight,
        padding: `${rows.pagePadding}px 10px`,
      }}
    >
      <Tooltip
        content={renderToString("AuthFlowsScreen.sketch.select-account-note")}
      >
        <span className={styles.pageCaption} style={{ height: rows.caption }}>
          {renderToString("AuthFlowsScreen.sketch.select-account-tag")}
        </span>
      </Tooltip>
      <span className={styles.pageTitle} style={{ height: rows.title }}>
        {renderToString("AuthFlowsScreen.sketch.select-account")}
      </span>
      <div className={styles.fixedRow} style={{ height: rows.button }}>
        <SketchButton
          label={renderToString("AuthFlowsScreen.sketch.continue-as-user")}
        />
      </div>
    </div>
  );
}

function OptionRows({
  item,
  hiddenIndex,
}: {
  item: FlowStepItem;
  hiddenIndex: number;
}) {
  const { renderToString } = useContext(Context);
  const isIdentify = item.step.type === "identify";
  const isCreate = item.step.type === "create_authenticator";
  const allOptions = item.step.one_of ?? [];
  const options = allOptions.map((option, i) => ({ option, i }));
  const shown = options.filter(
    ({ option, i }) => i !== hiddenIndex && !isDeviceTokenOption(option)
  );
  const deviceToken = options.find(({ option }) => isDeviceTokenOption(option));
  // Several alternatives are drawn as sub-nodes.
  const subnodes = shown.length > 1;

  // Options that hand off to the same two flows share a group box with one
  // has-account / no-account pair, at the position of the group's first.
  const groupOf = new Map<number, number[]>();
  for (const group of handoffGroups(allOptions)) {
    if (group.length > 1) {
      for (const i of group) {
        groupOf.set(i, group);
      }
    }
  }
  const entries: { key: number; members: typeof shown }[] = [];
  for (const entry of shown) {
    const group = groupOf.get(entry.i);
    if (group == null) {
      entries.push({ key: entry.i, members: [entry] });
    } else if (group[0] === entry.i) {
      entries.push({
        key: entry.i,
        members: shown.filter(({ i }) => group.includes(i)),
      });
    }
  }

  const sketch = (option: AuthenticationFlowOptionConfig) =>
    isIdentify ? (
      <IdentifyOption option={option} />
    ) : isCreate ? (
      <CreateAuthenticatorOption option={option} />
    ) : (
      <AuthenticatorOption option={option} />
    );
  const or = (
    <div className={styles.orDivider}>
      {renderToString("AuthFlowsScreen.sketch.or")}
    </div>
  );

  return (
    <>
      {entries.map(({ key, members }, k) => (
        <React.Fragment key={key}>
          {k > 0 ? or : null}
          {members.length === 1 ? (
            <OptionRow item={item} optionIndex={key} subnode={subnodes}>
              {sketch(members[0].option)}
            </OptionRow>
          ) : (
            <div className={cn(styles.optionRow, styles.optionGroup)}>
              {members.map(({ option, i }, m) => (
                <React.Fragment key={i}>
                  {m > 0 ? or : null}
                  {sketch(option)}
                </React.Fragment>
              ))}
              <OptionHandle
                item={item}
                optionIndex={key}
                className={styles.optionHandle}
              />
            </div>
          )}
        </React.Fragment>
      ))}
      {deviceToken != null ? (
        // Part of the page, not an alternative: it lets this device skip the
        // step next time.
        <div className={styles.rememberDevice}>
          <AuthenticatorOption option={deviceToken.option} />
          <span className={styles.rememberDeviceNote}>
            {renderToString("AuthFlowsScreen.sketch.remember-device-note")}
          </span>
        </div>
      ) : null}
    </>
  );
}

export function StepSketch({
  item,
  hiddenIndex,
}: {
  item: FlowStepItem;
  // An option drawn elsewhere (the stacked sheet), left out of the card.
  hiddenIndex: number;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { phoneOTPMode } = useFlowSketchContext();
  const { step } = item;
  switch (step.type) {
    case "identify":
    case "authenticate":
    case "create_authenticator":
      return <OptionRows item={item} hiddenIndex={hiddenIndex} />;
    case "verify":
      return (
        <SketchOTP
          label={renderToString(
            verifyCodeMessageID(item.identification, phoneOTPMode)
          )}
        />
      );
    case "verify_account_recovery_code":
      return <RecoveryCodeFields />;
    case "fill_form":
      return <FormFields form={step.form ?? {}} />;
    case "fill_in_user_profile":
      return <FormFields form={legacyProfileForm(step)} />;
    case "change_password":
    case "reset_password":
      return (
        <>
          <SketchInput
            label={renderToString("AuthFlowsScreen.sketch.new-password")}
          />
          <SketchInput
            label={renderToString("AuthFlowsScreen.sketch.confirm-password")}
          />
        </>
      );
    case "select_destination":
      // Only drawn as a page with enumerate_destinations: the user picks one
      // of the account's emails and phones, shown masked.
      return (
        <>
          <SketchText
            label={renderToString("AuthFlowsScreen.destination.choose")}
          />
          <SketchRadio
            label={renderToString("AuthFlowsScreen.destination.sample-email")}
            checked={true}
          />
          <SketchRadio
            label={renderToString("AuthFlowsScreen.destination.sample-phone")}
            checked={false}
          />
          <SketchButton label={renderToString("AuthFlowsScreen.sketch.send")} />
        </>
      );
    case "view_recovery_code":
      return (
        <div className={styles.sketchCodes}>
          {Array.from({ length: RECOVERY_CODE_COUNT }, (_, i) => (
            <span key={i}>••••-••••</span>
          ))}
        </div>
      );
    case "terminate_other_sessions":
      return (
        <>
          <SketchText
            label={renderToString("AuthFlowsScreen.sketch.other-device")}
          />
          <SketchButton
            label={renderToString("AuthFlowsScreen.sketch.continue")}
          />
          <SketchLink
            label={renderToString("AuthFlowsScreen.sketch.dont-log-out")}
          />
        </>
      );
    case "prompt_create_passkey":
      return (
        <>
          <SketchButton
            label={renderToString("AuthFlowsScreen.sketch.create-passkey")}
          />
          <SketchLink label={renderToString("AuthFlowsScreen.sketch.skip")} />
        </>
      );
    default:
      return item.branches != null ? (
        <OptionRows item={item} hiddenIndex={hiddenIndex} />
      ) : (
        <SketchLink
          label={renderToString("AuthFlowsScreen.sketch.unknown-step")}
        />
      );
  }
}
