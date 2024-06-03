import React, { ComponentType, useContext, useState, useMemo, useCallback } from "react";
import { Routes, Navigate, Route, useNavigate } from "react-router-dom";
import { useTheme } from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton, { DefaultButtonProps } from "./DefaultButton";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import SurveyLayout from "./OnboardingSurveyLayout";
import styles from "./OnboardingSurveyScreen.module.css";

function ChoiceButton(props: DefaultButtonProps) {
  const theme = useTheme();
  const { checked } = props;
  const styles = useMemo(() => {
    return {
      root: {
        border: "none",
        "--tw-ring-color": theme.semanticColors.inputBorder,
      },
      rootChecked: {
        "--tw-ring-color": theme.palette.themePrimary,
        color: theme.palette.themePrimary,
        backgroundColor: theme.semanticColors.buttonBackground,
      },
      rootCheckedHovered: {
        color: theme.palette.themePrimary,
      },
    };
  }, [theme]);
  return (
    <DefaultButton
      {...props}
      toggle={true}
      styles={styles}
      className={checked ? "ring-2" : "ring-1"}
    />
  );
}
/*
function CompoundChoiceButton (props: ButtonProps) {
  const { buttonKey, pressed, onClick } = props;
  const { renderToString } = useContext(Context);
  return (
    <CompoundButton
      text={renderToString(buttonKey)}
      secondaryText={renderToString(buttonKey + ".secondaryText")}
      toggle
      checked={pressed}
      className={styles.compoundChoiceButton}
      onClick={onClick}
    />
  )
}
*/

interface ChoiceButtonGroupProps {
  prefix: string;
  choices: readonly string[];
  state: Record<string, boolean>;
  setChoice: (choice: Record<string, boolean>) => void;
  processChoice: (
    oldChoice: Record<string, boolean>,
    selectedChoice: Readonly<string>
  ) => Record<string, boolean>;
  Button: ComponentType<DefaultButtonProps>;
}

function ChoiceButtonGroup(props: ChoiceButtonGroupProps) {
  const { prefix, choices, state, setChoice, processChoice, Button } = props;
  const { renderToString } = useContext(Context);
  const buttons = useMemo(
    () =>
      choices.map((choice) => {
        const key = [prefix, choice].join(".");
        return (
          <Button
            key={key}
            text={renderToString(key)}
            checked={state[choice]}
            onClick={() => {
              setChoice(structuredClone(processChoice(state, choice)));
            }}
          />
        );
      }),
    [state, Button, prefix, choices, processChoice, setChoice, renderToString]
  );
  return <div className={styles.singleChoiceButtonGroup}>{buttons}</div>;
}

function processSingleChoice(
  oldChoices: Record<string, boolean>,
  choice: Readonly<string>
): Record<string, boolean> {
  if (oldChoices[choice]) oldChoices[choice] = false;
  else
    Object.entries(oldChoices).forEach(([k, _]) => {
      oldChoices[k] = k === choice;
    });
  return oldChoices;
}

function allFalse(
  choices: readonly string[]
): Record<typeof choices[number], boolean> {
  const result = {} as Record<typeof choices[number], boolean>;
  choices.forEach((c) => {
    result[c] = false;
  });
  return result;
}

interface StepProps {}

function Step1(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step1";
  const roleChoiceGroup = "roleChoiceGroup";
  const roleChoices = [
    "Dev",
    "IT",
    "PM",
    "PD",
    "Market",
    "Owner",
    "Other",
  ] as const;
  const defaultRoleChoices: Record<typeof roleChoices[number], boolean> =
    allFalse(roleChoices);
  const [roleChoicesState, setRoleChoicesState] = useState(defaultRoleChoices);
  const empty = useMemo(() => {
    return (
      Object.entries(roleChoicesState).reduce(
        (acc, [_, v]) => acc + (v ? 1 : 0),
        0
      ) === 0
    );
  }, [roleChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const onClickNext = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../2");
  }, [navigate]);
  return (
    <SurveyLayout
      title={renderToString(prefix + ".title")}
      subtitle={renderToString(prefix + ".subtitle")}
      backButtonDisabled={true}
      primaryButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          disabled={empty}
        />
      }
      secondaryButton={
        <DefaultButton
          onClick={() => {}}
          text={<FormattedMessage id="back" />}
          className={styles.backButton}
        />
      }
    >
      <ChoiceButtonGroup
        prefix={[prefix, roleChoiceGroup].join(".")}
        choices={roleChoices}
        state={roleChoicesState}
        setChoice={setRoleChoicesState}
        processChoice={processSingleChoice}
        Button={ChoiceButton}
      />
    </SurveyLayout>
  );
}

function Step2(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step2";
  const toriChoiceGroup = "teamOrIndividualChoiceGroup";
  const toriChoices = [
    "Team",
    "Individual",
  ] as const;
  const defaultToriChoices: Record<typeof toriChoices[number], boolean> =
    allFalse(toriChoices);
  const [toriChoicesState, setToriChoicesState] = useState(defaultToriChoices);
  const empty = useMemo(() => {
    return (
      Object.entries(toriChoicesState).reduce(
        (acc, [_, v]) => acc + (v ? 1 : 0),
        0
      ) === 0
    );
  }, [toriChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const onClickNext = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      if (toriChoicesState["Team"]) navigate("./../3-team")
      if (toriChoicesState["Individual"]) navigate("./../3-individual")
  }, [navigate, toriChoicesState]);
  const onClickBack = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../1");
  }, [navigate]);
  return (
    <SurveyLayout
      title={renderToString(prefix + ".title")}
      subtitle={renderToString(prefix + ".subtitle")}
      backButtonDisabled={false}
      primaryButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          className={styles.nextButton}
          disabled={empty}
        />
      }
      secondaryButton={
        <DefaultButton
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
          className={styles.backButton}
        />
      }
    >
      <ChoiceButtonGroup
        prefix={[prefix, toriChoiceGroup].join(".")}
        choices={toriChoices}
        state={toriChoicesState}
        setChoice={setToriChoicesState}
        processChoice={processSingleChoice}
        Button={ChoiceButton}
      />
    </SurveyLayout>
  );
}


export const OnboardingSurveyScreen: React.VFC =
  function OnboardingSurveyScreen() {
    return (
      <Routes>
        <Route path="/1" element={<Step1 />} />
        <Route path="/2" element={<Step2 />} />
        <Route path="*" element={<Navigate to="1" replace={true} />} />
      </Routes>
    );
  };

export default OnboardingSurveyScreen;
