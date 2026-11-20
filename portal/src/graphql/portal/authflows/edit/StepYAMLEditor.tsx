import React, { useContext, useMemo, useState } from "react";
import yaml from "js-yaml";
import { Button, Flex, Text } from "@radix-ui/themes";
import { Context } from "../../../../intl";
import CodeEditor from "../../../../CodeEditor";
import { AuthenticationFlowStepConfig } from "../../../../types";
import { FlowType } from "../model";
import { ParseStepYAMLError, parseStepYAML } from "./parseStepYAML";
import styles from "./StepYAMLEditor.module.css";

const errorMessageIDs: Record<ParseStepYAMLError, string> = {
  parse: "AuthFlowsScreen.edit.yaml.error.parse",
  not_step: "AuthFlowsScreen.edit.yaml.error.not-step",
  type: "AuthFlowsScreen.edit.yaml.error.type",
  form: "AuthFlowsScreen.edit.yaml.error.form",
};

// The selected step, with everything nested under it, as YAML. Apply
// replaces the step in the draft; nothing changes until then.
export function StepYAMLEditor({
  step,
  flowType,
  onApply,
}: {
  step: AuthenticationFlowStepConfig;
  flowType: FlowType;
  onApply: (step: AuthenticationFlowStepConfig) => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const original = useMemo(() => yaml.dump(step), [step]);
  // Edits stay local until applied; a change to the step from elsewhere
  // (the form, the canvas) resets them through the key on this component.
  const [text, setText] = useState(original);
  const [error, setError] = useState<ParseStepYAMLError | null>(null);

  const onClickApply = () => {
    const result = parseStepYAML(text, flowType);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setError(null);
    onApply(result.step);
  };

  return (
    <div className={styles.root}>
      <div className={styles.editor}>
        <CodeEditor
          className={styles.codeEditor}
          language="yaml"
          value={text}
          onChange={(v) => {
            setText(v ?? "");
            setError(null);
          }}
          options={{ minimap: { enabled: false }, lineNumbers: "off" }}
        />
      </div>
      {error != null ? (
        <Text as="p" size="1" color="red">
          {renderToString(errorMessageIDs[error])}
        </Text>
      ) : null}
      <Flex justify="end">
        <Button
          type="button"
          size="1"
          disabled={text === original}
          onClick={onClickApply}
        >
          {renderToString("AuthFlowsScreen.edit.yaml.apply")}
        </Button>
      </Flex>
    </div>
  );
}
