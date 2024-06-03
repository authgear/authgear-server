import React, {
  ComponentType,
  useContext,
  useState,
  useMemo,
  useCallback,
} from "react";
import { Routes, Navigate, Route, useNavigate } from "react-router-dom";
import { useTheme, Label } from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton, { DefaultButtonProps } from "./DefaultButton";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import FormTextField from "./FormTextField";
import PhoneTextField, { PhoneTextFieldValues } from "./PhoneTextField";
import { FormProvider } from "./form";
import SurveyLayout from "./OnboardingSurveyLayout";
import styles from "./OnboardingSurveyScreen.module.css";

function ChoiceButton(props: DefaultButtonProps) {
  const theme = useTheme();
  const { checked } = props;
  const styles = useMemo(() => {
    return {
      root: {
        border: "none",
        "--tw-ring-color": theme.semanticColors.variantBorder,
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
  isIndividualComponent: boolean;
}

function ChoiceButtonGroup(props: ChoiceButtonGroupProps) {
  const {
    prefix,
    choices,
    state,
    setChoice,
    processChoice,
    Button,
    isIndividualComponent,
  } = props;
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
  return (
    <div>
      {isIndividualComponent ? null : (
        <Label>{renderToString(prefix + ".label")}</Label>
      )}
      <div
        className={
          isIndividualComponent
            ? styles.individualSingleChoiceButtonGroup
            : styles.singleChoiceButtonGroup
        }
      >
        {buttons}
      </div>
    </div>
  );
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
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../2");
    },
    [navigate]
  );
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
      secondaryButton={<DefaultButton />}
    >
      <ChoiceButtonGroup
        prefix={[prefix, roleChoiceGroup].join(".")}
        choices={roleChoices}
        state={roleChoicesState}
        setChoice={setRoleChoicesState}
        processChoice={processSingleChoice}
        Button={ChoiceButton}
        isIndividualComponent={true}
      />
    </SurveyLayout>
  );
}

function Step2(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step2";
  const toriChoiceGroup = "teamOrIndividualChoiceGroup";
  const toriChoices = ["Team", "Individual"] as const;
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
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (toriChoicesState["Team"]) navigate("./../3-team");
      if (toriChoicesState["Individual"]) navigate("./../3-individual");
    },
    [navigate, toriChoicesState]
  );
  const onClickBack = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../1");
    },
    [navigate]
  );
  const theme = useTheme();
  const backButtonStyles = useMemo(() => {
    return {
      root: {
        border: "none",
        "background-color": theme.semanticColors.bodyStandoutBackground,
      },
    };
  }, [theme]);
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
          styles={backButtonStyles}
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
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
        isIndividualComponent={true}
      />
    </SurveyLayout>
  );
}

function Step3Team(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step3-team";
  const [companyName, setCompanyName] = useState("");
  const defaultPhone: PhoneTextFieldValues = { rawInputValue: "" };
  const [companyPhone, setCompanyPhone] = useState(defaultPhone);
  const companySizeChoiceGroup = "companySizeChoiceGroup";
  const companySizeChoices = [
    "1-49",
    "50-99",
    "100-499",
    "500-1999",
    "2000+",
  ] as const;
  const defaultCompanySizeChoices: Record<
    typeof companySizeChoices[number],
    boolean
  > = allFalse(companySizeChoices);
  const [companySizeChoicesState, setCompanySizeChoicesState] = useState(
    defaultCompanySizeChoices
  );
  const companySizeEmpty = useMemo(() => {
    return (
      Object.entries(companySizeChoicesState).reduce(
        (acc, [_, v]) => acc + (v ? 1 : 0),
        0
      ) === 0
    );
  }, [companySizeChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../4");
    },
    [navigate]
  );
  const onClickBack = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../2");
    },
    [navigate]
  );
  const theme = useTheme();
  const inputStyles = useMemo(() => {
    return {
      fieldGroup: {
        "border-color": theme.semanticColors.variantBorder,
      },
    };
  }, [theme]);
  const backButtonStyles = useMemo(() => {
    return {
      root: {
        border: "none",
        "background-color": theme.semanticColors.bodyStandoutBackground,
      },
    };
  }, [theme]);
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
          disabled={companySizeEmpty || companyName === ""}
        />
      }
      secondaryButton={
        <DefaultButton
          styles={backButtonStyles}
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
        />
      }
    >
      <div className={styles.formBox}>
        <FormProvider loading={false}>
          <FormTextField
            parentJSONPointer={""}
            fieldName="companyName"
            styles={inputStyles}
            label={renderToString(prefix + ".companyName.label")}
            value={companyName}
            onChange={(_, v) => setCompanyName(v!)}
          />
          <ChoiceButtonGroup
            prefix={[prefix, companySizeChoiceGroup].join(".")}
            choices={companySizeChoices}
            state={companySizeChoicesState}
            setChoice={setCompanySizeChoicesState}
            processChoice={processSingleChoice}
            Button={ChoiceButton}
            isIndividualComponent={false}
          />
          <PhoneTextField
            label={renderToString(prefix + ".phone.label")}
            inputValue={companyPhone.rawInputValue}
            onChange={(v) => setCompanyPhone(v)}
          />
        </FormProvider>
      </div>
    </SurveyLayout>
  );
}

function Step3Individual(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step3-individual";
  const [individualWebsite, setIndividualWebsite] = useState("");
  const defaultPhone: PhoneTextFieldValues = { rawInputValue: "" };
  const [individualPhone, setIndividualPhone] = useState(defaultPhone);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../4");
    },
    [navigate]
  );
  const onClickBack = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../2");
    },
    [navigate]
  );
  const theme = useTheme();
  const inputStyles = useMemo(() => {
    return {
      fieldGroup: {
        "border-color": theme.semanticColors.variantBorder,
      },
    };
  }, [theme]);
  const backButtonStyles = useMemo(() => {
    return {
      root: {
        border: "none",
        "background-color": theme.semanticColors.bodyStandoutBackground,
      },
    };
  }, [theme]);
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
          disabled={false}
        />
      }
      secondaryButton={
        <DefaultButton
          styles={backButtonStyles}
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
        />
      }
    >
      <div className={styles.formBox}>
        <FormProvider loading={false}>
          <FormTextField
            parentJSONPointer={""}
            fieldName={"projectWebsite"}
            styles={inputStyles}
            label={renderToString(prefix + ".projectWebsite.label")}
            value={individualWebsite}
            onChange={(_, v) => setIndividualWebsite(v!)}
          />
          <PhoneTextField
            label={renderToString(prefix + ".phone.label")}
            inputValue={individualPhone.rawInputValue}
            onChange={(v) => setIndividualPhone(v)}
          />
        </FormProvider>
      </div>
    </SurveyLayout>
  );
}

export const OnboardingSurveyScreen: React.VFC =
  function OnboardingSurveyScreen() {
    return (
      <Routes>
        <Route path="/1" element={<Step1 />} />
        <Route path="/2" element={<Step2 />} />
        <Route path="/3-team" element={<Step3Team />} />
        <Route path="/3-individual" element={<Step3Individual />} />
        <Route path="*" element={<Navigate to="1" replace={true} />} />
      </Routes>
    );
  };

export default OnboardingSurveyScreen;
