import React, { useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  Button,
  Dialog,
  RadioCards,
  RadioGroup,
  Text,
  TextField,
} from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import Link from "../../../../Link";
import { Context } from "../../../../intl";
import {
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  ChatBubbleIcon,
  CheckCircledIcon,
  CopyIcon,
  EnvelopeClosedIcon,
  GlobeIcon,
  IdCardIcon,
  LayoutIcon,
  LockClosedIcon,
  MobileIcon,
  PlusIcon,
  TimerIcon,
  UploadIcon,
} from "@radix-ui/react-icons";
import { flowPath } from "../model";
import { CreatableType } from "./rows";
import { apiNameFrom, freeName, nameError } from "./actions";
import {
  MissingMethod,
  NewFlows,
  STARTER_IDS,
  StarterID,
  projectMethods,
  scratchFlows,
  scratchMissing,
  starterFlows,
  starterMissing,
  STARTER_STRIPS,
  Strip,
  StripItem,
} from "./starters";
import { DISPLAY_NAME_MAX, nameErrorMessageIDs } from "./FlowNameDialog";
import styles from "./NewFlowDialog.module.css";

type Start = "current" | "scratch" | "template" | "import";
// The templates step lists the common setups; "current" is its own start.
type TemplateID = Exclude<StarterID, "current">;

interface Choice<T extends string> {
  value: T;
  title: string;
  subtitle?: string;
  disabled?: boolean;
  // A tile on the left of the card.
  icon?: React.ReactNode;
}

// One full-width card per choice, text left-aligned, so the choices read as
// a list.
function ChoiceCards<T extends string>({
  choices,
  value,
  onValueChange,
}: {
  choices: Choice<T>[];
  value: T;
  onValueChange: (value: T) => void;
}): React.ReactElement {
  return (
    <RadioCards.Root
      columns="1"
      size="2"
      value={value}
      onValueChange={(v) => onValueChange(v as T)}
    >
      {choices.map((c) => (
        <RadioCards.Item key={c.value} value={c.value} disabled={c.disabled}>
          <div className={styles.choiceRow}>
            {c.icon != null ? (
              <span className={styles.choiceIcon} aria-hidden={true}>
                {c.icon}
              </span>
            ) : null}
            <div className={styles.choice}>
              <Text size="2" weight="medium">
                {c.title}
              </Text>
              {c.subtitle != null ? (
                <Text size="2" color="gray">
                  {c.subtitle}
                </Text>
              ) : null}
            </div>
          </div>
        </RadioCards.Item>
      ))}
    </RadioCards.Root>
  );
}

const startIcons: Record<Start, React.ReactNode> = {
  current: <CopyIcon width={18} height={18} />,
  scratch: <PlusIcon width={18} height={18} />,
  template: <LayoutIcon width={18} height={18} />,
  import: <UploadIcon width={18} height={18} />,
};

const stripIcons: Record<StripItem, React.ReactNode> = {
  email: <EnvelopeClosedIcon />,
  phone: <MobileIcon />,
  email_or_phone: <IdCardIcon />,
  otp: <ChatBubbleIcon />,
  verify: <CheckCircledIcon />,
  password: <LockClosedIcon />,
  social: <GlobeIcon />,
  authenticator: <TimerIcon />,
};

const stripMessageIDs: Record<StripItem, string> = {
  email: "AuthFlowsScreen.new.strip.email",
  phone: "AuthFlowsScreen.new.strip.phone",
  email_or_phone: "AuthFlowsScreen.new.strip.email_or_phone",
  otp: "AuthFlowsScreen.new.strip.otp",
  verify: "AuthFlowsScreen.new.strip.verify",
  password: "AuthFlowsScreen.new.strip.password",
  social: "AuthFlowsScreen.new.strip.social",
  authenticator: "AuthFlowsScreen.new.strip.authenticator",
};

// A flow's steps, top to bottom: an icon, then what the user does, with
// alternatives joined by "or".
function StepList({
  title,
  strip,
}: {
  title: string;
  strip: Strip;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.stepList}>
      <Text
        size="1"
        weight="medium"
        color="gray"
        className={styles.stepListTitle}
      >
        {title}
      </Text>
      <ol className={styles.steps}>
        {strip.map((choices, i) => (
          <li key={i} className={styles.stepItem}>
            <span className={styles.stepIcon} aria-hidden={true}>
              {stripIcons[choices[0]]}
            </span>
            <Text size="2">
              {choices
                .map((item) => renderToString(stripMessageIDs[item]))
                .join(` ${renderToString("AuthFlowsScreen.new.strip.or")} `)}
            </Text>
          </li>
        ))}
      </ol>
    </div>
  );
}

const starterIcons: Record<TemplateID, React.ReactNode> = {
  otp: <ChatBubbleIcon width={16} height={16} />,
  password: <LockClosedIcon width={16} height={16} />,
  verified_email_phone: <CheckCircledIcon width={16} height={16} />,
  social: <GlobeIcon width={16} height={16} />,
  password_2fa: <TimerIcon width={16} height={16} />,
};

const starterDescriptionMessageIDs: Record<TemplateID, string> = {
  otp: "AuthFlowsScreen.new.starter.otp.description",
  password: "AuthFlowsScreen.new.starter.password.description",
  verified_email_phone:
    "AuthFlowsScreen.new.starter.verified_email_phone.description",
  social: "AuthFlowsScreen.new.starter.social.description",
  password_2fa: "AuthFlowsScreen.new.starter.password_2fa.description",
};

type Step = "start" | "template" | "name";

// The section's New flow button picks the type.
const titleMessageIDs: Record<CreatableType, string> = {
  signup_login: "AuthFlowsScreen.new.title.signup_login",
  signup: "AuthFlowsScreen.new.title.signup",
  login: "AuthFlowsScreen.new.title.login",
};

const startMessageIDs: Record<Start, string> = {
  current: "AuthFlowsScreen.new.start.current",
  scratch: "AuthFlowsScreen.new.start.scratch",
  template: "AuthFlowsScreen.new.start.template",
  import: "AuthFlowsScreen.new.start.import",
};

const startDescriptionMessageIDs: Record<Start, string> = {
  current: "AuthFlowsScreen.new.start.current.description",
  scratch: "AuthFlowsScreen.new.start.scratch.description",
  template: "AuthFlowsScreen.new.start.template.description",
  import: "AuthFlowsScreen.new.start.import.description",
};

const starterMessageIDs: Record<TemplateID, string> = {
  otp: "AuthFlowsScreen.new.starter.otp",
  password: "AuthFlowsScreen.new.starter.password",
  verified_email_phone: "AuthFlowsScreen.new.starter.verified_email_phone",
  social: "AuthFlowsScreen.new.starter.social",
  password_2fa: "AuthFlowsScreen.new.starter.password_2fa",
};

const missingMessageIDs: Record<MissingMethod, string> = {
  email_login: "AuthFlowsScreen.new.missing.email_login",
  phone_login: "AuthFlowsScreen.new.missing.phone_login",
  email_or_phone: "AuthFlowsScreen.new.missing.email_or_phone",
  otp_login: "AuthFlowsScreen.new.missing.otp_login",
  password: "AuthFlowsScreen.new.missing.password",
  email_code: "AuthFlowsScreen.new.missing.email_code",
  sms_code: "AuthFlowsScreen.new.missing.sms_code",
  social: "AuthFlowsScreen.new.missing.social",
  totp: "AuthFlowsScreen.new.missing.totp",
};

export interface NewFlowDialogProps {
  appID: string;
  type: CreatableType;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // authgear.yaml's flows, as written.
  flows: AuthenticationFlowsConfig | undefined;
  generated: AuthenticationFlowsConfig;
  config: PortalAPIAppConfig;
}

// How to start, then (for a template) which one, then the name.
// Create opens the editor on the new flows as an unsaved draft.
export function NewFlowDialog({
  appID,
  type,
  open,
  onOpenChange,
  flows,
  generated,
  config,
}: NewFlowDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const methods = useMemo(() => projectMethods(config), [config]);
  const [step, setStep] = useState<Step>("start");
  const [start, setStart] = useState<Start>("current");
  const [starter, setStarter] = useState<TemplateID>("otp");
  const [displayName, setDisplayName] = useState("");
  // The name follows the display name until it is typed in.
  const [typedName, setTypedName] = useState<string | null>(null);
  const [fallbackName, setFallbackName] = useState("");
  const name = typedName ?? (apiNameFrom(displayName) || fallbackName);

  // A signup & login template also makes signup and login flows of the
  // same name, so the name must be free for those too.
  const makesAll = type === "signup_login" && start === "template";
  const error =
    nameError(flows, type, name) ??
    (makesAll
      ? nameError(flows, "signup", name) ?? nameError(flows, "login", name)
      : null);
  const displayNameTooLong = displayName.trim().length > DISPLAY_NAME_MAX;

  const templateIDs = STARTER_IDS.filter(
    (id): id is TemplateID => id !== "current"
  );
  const missingOf = (id: TemplateID) => starterMissing(id, methods);
  const toTemplates = () => {
    const first = templateIDs.find((id) => missingOf(id).length === 0);
    if (first != null) {
      setStarter(first);
    }
    setStep("template");
  };
  const toName = () => {
    setFallbackName(freeName(flows, type, type));
    setStep("name");
  };
  const create = () => {
    if (error != null || displayNameTooLong) {
      return;
    }
    const made: NewFlows =
      start === "current"
        ? starterFlows("current", type, name, generated, methods)
        : start === "template"
        ? starterFlows(starter, type, name, generated, methods)
        : scratchFlows(type, name, methods);
    const label = displayName.trim();
    // The flow being made gets the display name; flows made with it keep
    // theirs.
    const newFlows: NewFlows = made.map((f) =>
      f.type === type && f.flow.name === name && label !== ""
        ? { ...f, flow: { ...f.flow, display_name: label } }
        : f
    );
    onOpenChange(false);
    navigate(flowPath(appID, type, name), { state: { newFlows } });
  };

  const startOptions: Choice<Start>[] = (
    ["current", "scratch", "template", "import"] as const
  ).map((s) => ({
    value: s,
    title: renderToString(startMessageIDs[s]),
    subtitle:
      s === "scratch" && scratchMissing(methods)
        ? renderToString("AuthFlowsScreen.new.start.scratch.missing")
        : renderToString(startDescriptionMessageIDs[s]),
    disabled: s === "import" || (s === "scratch" && scratchMissing(methods)),
    icon: startIcons[s],
  }));
  const selectedMissing = missingOf(starter);
  const strips = STARTER_STRIPS[starter];

  const back: Partial<Record<Step, Step>> = {
    template: "start",
    name: start === "template" ? "template" : "start",
  };

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(o) => {
        onOpenChange(o);
        if (!o) {
          setStep("start");
          setDisplayName("");
          setTypedName(null);
        }
      }}
    >
      <Dialog.Content
        maxWidth={step === "template" ? "760px" : "560px"}
        size="3"
      >
        <Dialog.Title>{renderToString(titleMessageIDs[type])}</Dialog.Title>
        {step === "start" ? (
          <ChoiceCards
            choices={startOptions}
            value={start}
            onValueChange={setStart}
          />
        ) : null}
        {step === "template" ? (
          <div className={styles.picker}>
            <RadioGroup.Root
              className={styles.templateList}
              value={starter}
              onValueChange={(v) => setStarter(v as TemplateID)}
            >
              {templateIDs.map((id) => {
                const missing = missingOf(id).length > 0;
                return (
                  <label
                    key={id}
                    className={cn(
                      styles.templateItem,
                      id === starter ? styles.templateItemSelected : null,
                      missing ? styles.templateItemMissing : null
                    )}
                  >
                    <RadioGroup.Item value={id} className={styles.srOnly} />
                    <span className={styles.templateIcon} aria-hidden={true}>
                      {starterIcons[id]}
                    </span>
                    <span className={styles.choice}>
                      <Text size="2" weight="medium">
                        {renderToString(starterMessageIDs[id])}
                      </Text>
                      <Text size="1" color="gray">
                        {missing
                          ? renderToString("AuthFlowsScreen.new.unavailable")
                          : renderToString(starterDescriptionMessageIDs[id])}
                      </Text>
                    </span>
                  </label>
                );
              })}
            </RadioGroup.Root>
            <div className={styles.preview}>
              {selectedMissing.length > 0 ? (
                <div className={styles.previewMissing}>
                  <Text size="2" weight="medium">
                    {renderToString("AuthFlowsScreen.new.missing.title")}
                  </Text>
                  <ul className={styles.missingList}>
                    {selectedMissing.map((m) => (
                      <li key={m}>
                        <Text size="2" color="gray">
                          {renderToString(missingMessageIDs[m])}
                        </Text>
                      </li>
                    ))}
                  </ul>
                  <Text size="2">
                    <Link
                      to={`/project/${encodeURIComponent(
                        appID
                      )}/configuration/authentication/login-methods`}
                    >
                      {renderToString("AuthFlowsScreen.new.missing.link")}
                    </Link>
                  </Text>
                </div>
              ) : (
                <div className={styles.previewSteps}>
                  {type !== "login" ? (
                    <StepList
                      title={renderToString("AuthFlowsScreen.new.strip.signup")}
                      strip={strips.signup}
                    />
                  ) : null}
                  {type !== "signup" ? (
                    <StepList
                      title={renderToString("AuthFlowsScreen.new.strip.login")}
                      strip={strips.login}
                    />
                  ) : null}
                </div>
              )}
            </div>
          </div>
        ) : null}
        {step === "name" ? (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              create();
            }}
          >
            <label className={styles.field}>
              <Text size="2" weight="medium">
                {renderToString("AuthFlowsScreen.list.display-name")}
              </Text>
              <TextField.Root
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                autoFocus={true}
              />
              {displayNameTooLong ? (
                <Text as="p" size="1" color="red">
                  {renderToString(
                    "AuthFlowsScreen.list.display-name.too-long",
                    { max: DISPLAY_NAME_MAX }
                  )}
                </Text>
              ) : null}
            </label>
            <label className={styles.field}>
              <Text size="2" weight="medium">
                {renderToString("AuthFlowsScreen.list.flow-name")}
              </Text>
              <TextField.Root
                value={name}
                onChange={(e) => setTypedName(e.target.value)}
              />
              {error != null ? (
                <Text as="p" size="1" color="red">
                  {renderToString(nameErrorMessageIDs[error])}
                </Text>
              ) : (
                <Text as="p" size="1" color="gray">
                  {renderToString("AuthFlowsScreen.list.flow-name.hint")}
                </Text>
              )}
            </label>
          </form>
        ) : null}
        <div className={styles.actions}>
          {back[step] != null ? (
            <SecondaryButton
              size="2"
              text={renderToString("AuthFlowsScreen.new.back")}
              onClick={() => setStep(back[step] ?? "start")}
            />
          ) : null}
          <span className={styles.spacer} />
          {step === "start" ? (
            <Button
              size="2"
              onClick={() => (start === "template" ? toTemplates() : toName())}
            >
              {renderToString("AuthFlowsScreen.new.next")}
            </Button>
          ) : null}
          {step === "template" ? (
            <Button
              size="2"
              disabled={starterMissing(starter, methods).length > 0}
              onClick={toName}
            >
              {renderToString("AuthFlowsScreen.new.next")}
            </Button>
          ) : null}
          {step === "name" ? (
            <Button
              size="2"
              disabled={error != null || displayNameTooLong}
              onClick={create}
            >
              {renderToString("AuthFlowsScreen.new.create")}
            </Button>
          ) : null}
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
