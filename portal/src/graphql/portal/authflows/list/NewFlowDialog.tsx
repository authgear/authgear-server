import React, { useContext, useMemo, useState } from "react";
import { Button, Dialog, RadioCards, Text, TextField } from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import Link from "../../../../Link";
import { Context } from "../../../../intl";
import {
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import { flowPath } from "../model";
import { stepTitle } from "../labels";
import { CreatableType } from "./rows";
import { freeName, nameError } from "./actions";
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
  stepSummary,
} from "./starters";
import { nameErrorMessageIDs } from "./FlowNameDialog";
import styles from "./NewFlowDialog.module.css";

type Start = "current" | "scratch" | "template" | "import";
// The templates step lists the common setups; "current" is its own start.
type TemplateID = Exclude<StarterID, "current">;

interface Choice<T extends string> {
  value: T;
  title: string;
  subtitle?: string;
  disabled?: boolean;
}

// One full-width card per choice, text left-aligned, so the steps read as a
// list rather than a grid of tiles.
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
        </RadioCards.Item>
      ))}
    </RadioCards.Root>
  );
}
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
  email_password: "AuthFlowsScreen.new.starter.email_password",
  email_code: "AuthFlowsScreen.new.starter.email_code",
  phone_code: "AuthFlowsScreen.new.starter.phone_code",
  email_password_2fa: "AuthFlowsScreen.new.starter.email_password_2fa",
};

const missingMessageIDs: Record<MissingMethod, string> = {
  email_login: "AuthFlowsScreen.new.missing.email_login",
  phone_login: "AuthFlowsScreen.new.missing.phone_login",
  password: "AuthFlowsScreen.new.missing.password",
  email_code: "AuthFlowsScreen.new.missing.email_code",
  sms_code: "AuthFlowsScreen.new.missing.sms_code",
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
  const [starter, setStarter] = useState<TemplateID>("email_password");
  const [name, setName] = useState("");

  const error = nameError(flows, type, name);

  const toName = () => {
    setName(freeName(flows, type, type));
    setStep("name");
  };
  const create = () => {
    if (error != null) {
      return;
    }
    const newFlows: NewFlows =
      start === "current"
        ? starterFlows("current", type, name, generated)
        : start === "template"
        ? starterFlows(starter, type, name, generated)
        : scratchFlows(type, name, methods);
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
  }));
  const templateIDs = STARTER_IDS.filter(
    (id): id is TemplateID => id !== "current"
  );
  const starterOptions: Choice<TemplateID>[] = templateIDs.map((id) => {
    const missing = starterMissing(id, methods);
    // A template always makes exactly one flow.
    const [{ flow }] = starterFlows(id, type, "preview", generated);
    return {
      value: id,
      title: renderToString(starterMessageIDs[id]),
      subtitle:
        missing.length > 0
          ? missing.map((m) => renderToString(missingMessageIDs[m])).join(" ")
          : stepSummary(flow)
              .map((t) => stepTitle(t, renderToString))
              .join(" → "),
      disabled: missing.length > 0,
    };
  });
  const anyMissing = templateIDs.some(
    (id) => starterMissing(id, methods).length > 0
  );

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
        }
      }}
    >
      <Dialog.Content maxWidth="560px" size="3">
        <Dialog.Title>{renderToString(titleMessageIDs[type])}</Dialog.Title>
        {step === "start" ? (
          <ChoiceCards
            choices={startOptions}
            value={start}
            onValueChange={setStart}
          />
        ) : null}
        {step === "template" ? (
          <>
            <ChoiceCards
              choices={starterOptions}
              value={starter}
              onValueChange={setStarter}
            />
            {anyMissing ? (
              <Text as="p" size="1" color="gray" className={styles.note}>
                {renderToString("AuthFlowsScreen.new.missing.note")}{" "}
                <Link
                  to={`/project/${encodeURIComponent(
                    appID
                  )}/configuration/authentication/login-methods`}
                >
                  {renderToString("AuthFlowsScreen.new.missing.link")}
                </Link>
              </Text>
            ) : null}
          </>
        ) : null}
        {step === "name" ? (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              create();
            }}
          >
            <TextField.Root
              value={name}
              onChange={(e) => setName(e.target.value)}
              aria-label={renderToString("AuthFlowsScreen.list.name")}
              autoFocus={true}
            />
            {error != null ? (
              <Text as="p" size="1" color="red" className={styles.note}>
                {renderToString(nameErrorMessageIDs[error])}
              </Text>
            ) : null}
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
              onClick={() =>
                start === "template" ? setStep("template") : toName()
              }
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
            <Button size="2" disabled={error != null} onClick={create}>
              {renderToString("AuthFlowsScreen.new.create")}
            </Button>
          ) : null}
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
