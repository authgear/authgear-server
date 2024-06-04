import React, { useContext, useState, useMemo, useCallback } from "react";
import { Routes, Navigate, Route, useNavigate } from "react-router-dom";
import { useTheme, Label, CompoundButton, IButtonProps } from "@fluentui/react";
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

interface DefaultCompoundButtonProps
  extends Omit<IButtonProps, "children" | "text" | "secondaryText"> {
  text?: React.ReactNode;
  secondaryText?: React.ReactNode;
}

function CompoundChoiceButton(props: DefaultCompoundButtonProps) {
  const theme = useTheme();
  const { checked } = props;
  const overrideStyles = useMemo(() => {
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
      label: {
        "margin-bottom": "10px",
        "font-size": "medium",
      },
      description: {
        "line-height": "18px",
        "font-size": "small",
      },
    };
  }, [theme]);
  return (
    // @ts-expect-error
    <CompoundButton
      {...props}
      toggle={true}
      styles={overrideStyles}
      className={
        (checked ? "ring-2" : "ring-1") + " " + styles.CompoundChoiceButton
      }
    />
  );
}

interface ChoiceButtonGroupProps {
  prefix: string;
  availableChoices: string[];
  selectedChoices: string[];
  setChoice: (newChoices: string[]) => void;
}

function processSingleChoice(
  availableChoices: string[],
  oldSelectedChoices: string[],
  newlySelectedChoice: string
): string[] {
  if (!availableChoices.includes(newlySelectedChoice))
    return oldSelectedChoices;
  else if (oldSelectedChoices.includes(newlySelectedChoice))
    return oldSelectedChoices.filter(
      (choice) => choice !== newlySelectedChoice
    );
  return [newlySelectedChoice];
}

function processMultiChoice(
  availableChoices: string[],
  oldSelectedChoices: string[],
  newlySelectedChoice: string
): string[] {
  if (!availableChoices.includes(newlySelectedChoice))
    return oldSelectedChoices;
  else if (oldSelectedChoices.includes(newlySelectedChoice))
    return oldSelectedChoices.filter(
      (choice) => choice !== newlySelectedChoice
    );
  oldSelectedChoices.push(newlySelectedChoice);
  return oldSelectedChoices;
}

function SingleChoiceButtonGroupVariantCentered(props: ChoiceButtonGroupProps) {
  const { prefix, availableChoices, selectedChoices, setChoice } = props;
  const { renderToString } = useContext(Context);
  const buttons = useMemo(
    () =>
      availableChoices.map((choice) => {
        const key = [prefix, choice].join(".");
        return (
          <ChoiceButton
            key={key}
            text={renderToString(key)}
            checked={selectedChoices.includes(choice)}
            onClick={() => {
              setChoice(
                structuredClone(
                  processSingleChoice(availableChoices, selectedChoices, choice)
                )
              );
            }}
          />
        );
      }),
    [prefix, availableChoices, selectedChoices, setChoice, renderToString]
  );
  return (
    <div className={styles.SingleChoiceButtonGroupVariantCentered}>
      {buttons}
    </div>
  );
}

function SingleChoiceButtonGroupVariantLabeled(props: ChoiceButtonGroupProps) {
  const { prefix, availableChoices, selectedChoices, setChoice } = props;
  const { renderToString } = useContext(Context);
  const buttons = useMemo(
    () =>
      availableChoices.map((choice) => {
        const key = [prefix, choice].join(".");
        return (
          <ChoiceButton
            key={key}
            text={renderToString(key)}
            checked={selectedChoices.includes(choice)}
            onClick={() => {
              setChoice(
                structuredClone(
                  processSingleChoice(availableChoices, selectedChoices, choice)
                )
              );
            }}
          />
        );
      }),
    [prefix, availableChoices, selectedChoices, setChoice, renderToString]
  );
  return (
    <div>
      <Label>{renderToString(prefix + ".label")}</Label>
      <div className={styles.SingleChoiceButtonGroupVariantLabeled}>
        {buttons}
      </div>
    </div>
  );
}

function MultiChoiceButtonGroup(props: ChoiceButtonGroupProps) {
  const { prefix, availableChoices, selectedChoices, setChoice } = props;
  const { renderToString } = useContext(Context);
  const buttons = useMemo(
    () =>
      availableChoices.map((choice) => {
        const key = [prefix, choice].join(".");
        return (
          <CompoundChoiceButton
            key={key}
            text={renderToString(key + ".title")}
            secondaryText={renderToString(key + ".subtitle")}
            checked={selectedChoices.includes(choice)}
            onClick={() => {
              setChoice(
                structuredClone(
                  processMultiChoice(availableChoices, selectedChoices, choice)
                )
              );
            }}
          />
        );
      }),
    [prefix, availableChoices, selectedChoices, setChoice, renderToString]
  );
  return <div className={styles.MultiChoiceButtonGroup}>{buttons}</div>;
}

interface StepProps {}

function Step1(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step1";
  const roleChoiceGroup = "roleChoiceGroup";
  const roleChoices = ["Dev", "IT", "PM", "PD", "Market", "Owner", "Other"];
  const defaultRoleChoicesState: string[] = [];
  const [roleChoicesState, setRoleChoicesState] = useState(
    defaultRoleChoicesState
  );
  const empty = useMemo(() => {
    return roleChoicesState.length === 0;
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
      title={renderToString("OnboardingSurveyScreen.step1.title")}
      subtitle={renderToString("OnboardingSurveyScreen.step1.subtitle")}
      nextButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          disabled={empty}
        />
      }
    >
      <SingleChoiceButtonGroupVariantCentered
        prefix={[prefix, roleChoiceGroup].join(".")}
        availableChoices={roleChoices}
        selectedChoices={roleChoicesState}
        setChoice={setRoleChoicesState}
      />
    </SurveyLayout>
  );
}

function Step2(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step2";
  const toriChoiceGroup = "teamOrIndividualChoiceGroup";
  const toriChoices = ["Team", "Individual"];
  const defaultToriChoices: string[] = [];
  const [toriChoicesState, setToriChoicesState] = useState(defaultToriChoices);
  const empty = useMemo(() => {
    return toriChoicesState.length === 0;
  }, [toriChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (toriChoicesState.includes("Team")) navigate("./../3-team");
      if (toriChoicesState.includes("Individual"))
        navigate("./../3-individual");
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
      title={renderToString("OnboardingSurveyScreen.step2.title")}
      subtitle={renderToString("OnboardingSurveyScreen.step2.subtitle")}
      nextButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          className={styles.nextButton}
          disabled={empty}
        />
      }
      backButton={
        <DefaultButton
          styles={backButtonStyles}
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
        />
      }
    >
      <SingleChoiceButtonGroupVariantCentered
        prefix={[prefix, toriChoiceGroup].join(".")}
        availableChoices={toriChoices}
        selectedChoices={toriChoicesState}
        setChoice={setToriChoicesState}
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
  const companySizeChoices = ["1-49", "50-99", "100-499", "500-1999", "2000+"];
  const defaultCompanySizeChoices: string[] = [];
  const [companySizeChoicesState, setCompanySizeChoicesState] = useState(
    defaultCompanySizeChoices
  );
  const companySizeEmpty = useMemo(() => {
    return companySizeChoicesState.length === 0;
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
      title={renderToString("OnboardingSurveyScreen.step3-team.title")}
      subtitle={renderToString("OnboardingSurveyScreen.step3-team.subtitle")}
      nextButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          className={styles.nextButton}
          disabled={companySizeEmpty || companyName === ""}
        />
      }
      backButton={
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
            label={renderToString(
              "OnboardingSurveyScreen.step3-team.companyName.label"
            )}
            value={companyName}
            onChange={(_, v) => setCompanyName(v!)}
          />
          <SingleChoiceButtonGroupVariantLabeled
            prefix={[prefix, companySizeChoiceGroup].join(".")}
            availableChoices={companySizeChoices}
            selectedChoices={companySizeChoicesState}
            setChoice={setCompanySizeChoicesState}
          />
          <PhoneTextField
            label={renderToString(
              "OnboardingSurveyScreen.step3-team.phone.label"
            )}
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
      title={renderToString("OnboardingSurveyScreen.step3-individual.title")}
      subtitle={renderToString(
        "OnboardingSurveyScreen.step3-individual.subtitle"
      )}
      nextButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
          className={styles.nextButton}
          disabled={false}
        />
      }
      backButton={
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
            label={renderToString(
              "OnboardingSurveyScreen.step3-individual.projectWebsite.label"
            )}
            value={individualWebsite}
            onChange={(_, v) => setIndividualWebsite(v!)}
          />
          <PhoneTextField
            label={renderToString(
              "OnboardingSurveyScreen.step3-individual.phone.label"
            )}
            inputValue={individualPhone.rawInputValue}
            onChange={(v) => setIndividualPhone(v)}
          />
        </FormProvider>
      </div>
    </SurveyLayout>
  );
}

function Step4(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step4";
  const reasonChoiceGroup = "reasonChoiceGroup";
  const reasonChoices = ["Auth", "SSO", "Security", "Portal", "Other"];
  const defaultReasonChoices: string[] = [];
  const [reasonChoicesState, setReasonChoicesState] =
    useState(defaultReasonChoices);
  const empty = useMemo(() => {
    return reasonChoicesState.length === 0;
  }, [reasonChoicesState]);
  const { renderToString } = useContext(Context);
  const [otherReason, setOtherReason] = useState("");
  const navigate = useNavigate();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../../projects/create");
    },
    [navigate]
  );
  const onClickBack = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      //TODO: change to 3-team or 3-individual depending on localStorage
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
      title={renderToString("OnboardingSurveyScreen.step4.title")}
      subtitle={renderToString("OnboardingSurveyScreen.step4.subtitle")}
      nextButton={
        <PrimaryButton
          onClick={onClickNext}
          text={renderToString("OnboardingSurveyScreen.step4.finish")}
          className={styles.nextButton}
          disabled={empty}
        />
      }
      backButton={
        <DefaultButton
          styles={backButtonStyles}
          onClick={onClickBack}
          text={<FormattedMessage id="back" />}
        />
      }
    >
      <MultiChoiceButtonGroup
        prefix={[prefix, reasonChoiceGroup].join(".")}
        availableChoices={reasonChoices}
        selectedChoices={reasonChoicesState}
        setChoice={setReasonChoicesState}
      />
      {reasonChoicesState.includes("Other") ? (
        <FormProvider loading={false}>
          <FormTextField
            parentJSONPointer={""}
            fieldName={"otherReason"}
            styles={inputStyles}
            className={styles.otherReasonInput}
            label={renderToString(
              "OnboardingSurveyScreen.step4.otherReason.label"
            )}
            value={otherReason}
            onChange={(_, v) => setOtherReason(v!)}
          />
        </FormProvider>
      ) : null}
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
        <Route path="/4" element={<Step4 />} />
        <Route path="*" element={<Navigate to="1" replace={true} />} />
      </Routes>
    );
  };

export default OnboardingSurveyScreen;
