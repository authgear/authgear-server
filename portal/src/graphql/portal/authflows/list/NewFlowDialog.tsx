import React, { useContext, useMemo, useState } from "react";
import { Button, Dialog, Text, TextField } from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import Link from "../../../../Link";
import { Context } from "../../../../intl";
import {
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  RadioCardOption,
  RadioCards,
} from "../../../../components/v2/RadioCards/RadioCards";
import { DEFAULT_FLOW_NAME, flowPath } from "../model";
import { stepTitle } from "../labels";
import { CREATABLE_TYPES, CreatableType, bundleTypes } from "./rows";
import { freeName, nameError } from "./actions";
import {
  MissingMethod,
  NewFlows,
  STARTER_IDS,
  StarterID,
  projectMethods,
  scratchFlows,
  starterFlows,
  starterMissing,
  stepSummary,
} from "./starters";
import { nameErrorMessageIDs } from "./FlowNameDialog";
import styles from "./NewFlowDialog.module.css";

type Start = "scratch" | "template" | "import";
type Step = "type" | "start" | "template" | "name";

const typeTitleMessageIDs: Record<CreatableType, string> = {
  signup_login: "AuthFlowsScreen.new.type.signup_login",
  signup: "AuthFlowsScreen.new.type.signup",
  login: "AuthFlowsScreen.new.type.login",
};

const typeDescriptionMessageIDs: Record<CreatableType, string> = {
  signup_login: "AuthFlowsScreen.new.type.signup_login.description",
  signup: "AuthFlowsScreen.new.type.signup.description",
  login: "AuthFlowsScreen.new.type.login.description",
};

const startMessageIDs: Record<Start, string> = {
  scratch: "AuthFlowsScreen.new.start.scratch",
  template: "AuthFlowsScreen.new.start.template",
  import: "AuthFlowsScreen.new.start.import",
};

const startDescriptionMessageIDs: Record<Start, string> = {
  scratch: "AuthFlowsScreen.new.start.scratch.description",
  template: "AuthFlowsScreen.new.start.template.description",
  import: "AuthFlowsScreen.new.start.import.description",
};

const starterMessageIDs: Record<StarterID, string> = {
  current: "AuthFlowsScreen.new.starter.current",
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
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // authgear.yaml's flows, as written.
  flows: AuthenticationFlowsConfig | undefined;
  generated: AuthenticationFlowsConfig;
  config: PortalAPIAppConfig;
}

// Type, then how to start, then (for a template) which one, then the name.
// Create opens the editor on the new flows as an unsaved draft.
export function NewFlowDialog({
  appID,
  open,
  onOpenChange,
  flows,
  generated,
  config,
}: NewFlowDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const methods = useMemo(() => projectMethods(config), [config]);
  const [step, setStep] = useState<Step>("type");
  const [type, setType] = useState<CreatableType>("signup_login");
  const [start, setStart] = useState<Start>("scratch");
  const [starter, setStarter] = useState<StarterID>("current");
  const [name, setName] = useState("");

  const types = type === "signup_login" ? bundleTypes() : [type];
  // default is free only while no custom flow of these types has it.
  const allowDefault = nameError(flows, types, DEFAULT_FLOW_NAME, true) == null;
  const error = nameError(flows, types, name, allowDefault);

  const toName = () => {
    setName(freeName(flows, types, type));
    setStep("name");
  };
  const create = () => {
    if (error != null) {
      return;
    }
    const newFlows: NewFlows =
      start === "template"
        ? starterFlows(starter, type, name, generated)
        : scratchFlows(type, name, methods);
    onOpenChange(false);
    navigate(flowPath(appID, type, name), { state: { newFlows } });
  };

  const typeOptions: RadioCardOption<CreatableType>[] = CREATABLE_TYPES.map(
    (t) => ({
      value: t,
      title: renderToString(typeTitleMessageIDs[t]),
      subtitle: renderToString(typeDescriptionMessageIDs[t]),
    })
  );
  const startOptions: RadioCardOption<Start>[] = (
    ["scratch", "template", "import"] as const
  ).map((s) => ({
    value: s,
    title: renderToString(startMessageIDs[s]),
    subtitle: renderToString(startDescriptionMessageIDs[s]),
    disabled: s === "import",
  }));
  // The flow a template card summarises: the login flow for Login, else the
  // signup flow.
  const summaryType = type === "login" ? "login" : "signup";
  const starterOptions: RadioCardOption<StarterID>[] = STARTER_IDS.map((id) => {
    const missing = starterMissing(id, methods);
    const flow = starterFlows(id, type, "preview", generated).find(
      (f) => f.type === summaryType
    )?.flow;
    return {
      value: id,
      title: renderToString(starterMessageIDs[id]),
      subtitle:
        missing.length > 0
          ? missing.map((m) => renderToString(missingMessageIDs[m])).join(" ")
          : flow != null
          ? stepSummary(flow)
              .map((t) => stepTitle(t, renderToString))
              .join(" → ")
          : undefined,
      disabled: missing.length > 0,
    };
  });
  const anyMissing = STARTER_IDS.some(
    (id) => starterMissing(id, methods).length > 0
  );

  const back: Partial<Record<Step, Step>> = {
    start: "type",
    template: "start",
    name: start === "template" ? "template" : "start",
  };

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(o) => {
        onOpenChange(o);
        if (!o) {
          setStep("type");
        }
      }}
    >
      <Dialog.Content maxWidth="560px" size="3">
        <Dialog.Title>
          {renderToString("AuthFlowsScreen.new.title")}
        </Dialog.Title>
        {step === "type" ? (
          <RadioCards
            size="2"
            numberOfColumns={1}
            options={typeOptions}
            value={type}
            onValueChange={setType}
          />
        ) : null}
        {step === "start" ? (
          <RadioCards
            size="2"
            numberOfColumns={1}
            options={startOptions}
            value={start}
            onValueChange={setStart}
          />
        ) : null}
        {step === "template" ? (
          <>
            <RadioCards
              size="2"
              numberOfColumns={1}
              options={starterOptions}
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
            ) : name === DEFAULT_FLOW_NAME ? (
              <Text as="p" size="1" color="gray" className={styles.note}>
                {renderToString("AuthFlowsScreen.new.name.default")}
              </Text>
            ) : null}
          </form>
        ) : null}
        <div className={styles.actions}>
          {back[step] != null ? (
            <SecondaryButton
              size="2"
              text={renderToString("AuthFlowsScreen.new.back")}
              onClick={() => setStep(back[step] ?? "type")}
            />
          ) : null}
          <span className={styles.spacer} />
          {step === "type" ? (
            <Button size="2" onClick={() => setStep("start")}>
              {renderToString("AuthFlowsScreen.new.next")}
            </Button>
          ) : null}
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
