import React, { useContext, useMemo } from "react";
import yaml from "js-yaml";
import { Flex, Heading, IconButton, Text } from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../../intl";
import { AuthenticatorPhoneOTPMode } from "../../../types";
import { FlowStepItem } from "./tree";
import { optionLabel, stepDescriptionMessageID, stepTitle } from "./labels";
import { StepForm, StepFormProps } from "./edit/StepForm";
import { StepProblem } from "./edit/FlowEditorContext";
import { Callout } from "../../../components/v2/Callout/Callout";
import { StepYAMLEditor } from "./edit/StepYAMLEditor";
import { replaceStep } from "./edit/edits";
import { FormLanguageSelect, FormTextFields } from "./edit/FormTextFields";
import { FlowType } from "./model";
import styles from "./StepDetailsPanel.module.css";

export interface StepDetailsPanelProps {
  item: FlowStepItem;
  flowType: FlowType;
  // How phone codes are sent, to name the channel in the options list.
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined;
  onClose: () => void;
  // The step's save errors, checks and warnings; while editing, the form
  // shows them.
  problems: StepProblem[];
  // Present in edit mode: the panel becomes a form for the step.
  edit?: Omit<StepFormProps, "stepID">;
}

export function StepDetailsPanel({
  item,
  flowType,
  phoneOTPMode,
  onClose,
  problems,
  edit,
}: StepDetailsPanelProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { step } = item;
  const description = stepDescriptionMessageID(step.type);
  const stepYAML = useMemo(() => yaml.dump(step), [step]);
  return (
    <aside className={styles.panel}>
      <Flex justify="between" align="center">
        <Heading size="3" as="h2">
          {stepTitle(step.type, renderToString)}
        </Heading>
        <IconButton
          variant="ghost"
          color="gray"
          aria-label={renderToString("AuthFlowsScreen.panel.close")}
          onClick={onClose}
        >
          <Cross2Icon />
        </IconButton>
      </Flex>
      {edit != null ? (
        <>
          {description != null ? (
            <Text as="p" size="2" color="gray">
              <FormattedMessage id={description} />
            </Text>
          ) : null}
          <StepForm {...edit} stepID={item.id} />
          <Heading size="2" as="h3">
            <FormattedMessage id="AuthFlowsScreen.edit.yaml.title" />
          </Heading>
          <StepYAMLEditor
            key={`${item.id}:${stepYAML}`}
            step={step}
            flowType={edit.flowType}
            onApply={(next) =>
              edit.apply(replaceStep(edit.flow, item.id, next))
            }
          />
        </>
      ) : (
        <>
          {problems.map((p, i) => (
            <Callout
              key={i}
              type={p.tone === "error" ? "error" : "warning"}
              size="1"
              text={p.text}
              showCloseButton={false}
            />
          ))}
          {/* A form step's text, to read in each language. */}
          {step.type === "fill_form" ? (
            <>
              <FormLanguageSelect formName={step.form?.name} />
              <FormTextFields
                formName={step.form?.name}
                flowType={flowType}
                ensureName={() => step.form?.name ?? ""}
              />
            </>
          ) : null}
          <ReadOnlyDetails
            item={item}
            phoneOTPMode={phoneOTPMode}
            stepYAML={stepYAML}
          />
        </>
      )}
    </aside>
  );
}

function ReadOnlyDetails({
  item,
  phoneOTPMode,
  stepYAML,
}: {
  item: FlowStepItem;
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined;
  stepYAML: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { step } = item;
  const description = stepDescriptionMessageID(step.type);
  const options = step.one_of ?? [];
  return (
    <>
      {step.name != null && step.type !== "fill_form" ? (
        <dl className={styles.nameList}>
          <dt>
            <FormattedMessage id="AuthFlowsScreen.panel.name" />
          </dt>
          <dd>
            <code>{step.name}</code>
          </dd>
        </dl>
      ) : null}
      {description != null ? (
        <Text as="p" size="2">
          <FormattedMessage id={description} />
        </Text>
      ) : null}
      {item.optional ? (
        <Text as="p" size="2">
          <FormattedMessage id="AuthFlowsScreen.panel.optional" />
        </Text>
      ) : null}
      {step.target_step != null ? (
        <Text as="p" size="2">
          <FormattedMessage
            id="AuthFlowsScreen.panel.target-step"
            values={{ name: step.target_step }}
          />
        </Text>
      ) : null}
      {options.length > 0 ? (
        <>
          <Heading size="2" as="h3">
            <FormattedMessage id="AuthFlowsScreen.panel.options" />
          </Heading>
          <ul className={styles.options}>
            {options.map((o, i) => (
              <li key={i}>{optionLabel(o, renderToString, phoneOTPMode)}</li>
            ))}
          </ul>
        </>
      ) : null}
      <Heading size="2" as="h3">
        <FormattedMessage id="AuthFlowsScreen.panel.yaml" />
      </Heading>
      <pre className={styles.yaml}>{stepYAML}</pre>
    </>
  );
}
