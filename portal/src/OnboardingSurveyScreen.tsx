import cn from "classnames";
import React, {
  useContext,
  useState,
  useMemo,
  useCallback,
  useEffect,
} from "react";
import {
  Routes,
  Navigate,
  Route,
  useNavigate,
  NavigateFunction,
} from "react-router-dom";
import { useTheme, Label, CompoundButton, IButtonProps } from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton, { DefaultButtonProps } from "./DefaultButton";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import FormTextField from "./FormTextField";
import PhoneTextField, { PhoneTextFieldValues } from "./PhoneTextField";
import { FormProvider } from "./form";
import SurveyLayout from "./OnboardingSurveyLayout";
import styles from "./OnboardingSurveyScreen.module.css";

const buttonTranslationKeys = {
  step1_roleChoiceGroup_Dev: "OnboardingSurveyScreen.step1.roleChoiceGroup.Dev",
  step1_roleChoiceGroup_IT: "OnboardingSurveyScreen.step1.roleChoiceGroup.IT",
  step1_roleChoiceGroup_PM: "OnboardingSurveyScreen.step1.roleChoiceGroup.PM",
  step1_roleChoiceGroup_PD: "OnboardingSurveyScreen.step1.roleChoiceGroup.PD",
  step1_roleChoiceGroup_Market:
    "OnboardingSurveyScreen.step1.roleChoiceGroup.Market",
  step1_roleChoiceGroup_Owner:
    "OnboardingSurveyScreen.step1.roleChoiceGroup.Owner",
  step1_roleChoiceGroup_Other:
    "OnboardingSurveyScreen.step1.roleChoiceGroup.Other",
  step2_teamOrIndividualChoiceGroup_Team:
    "OnboardingSurveyScreen.step2.teamOrIndividualChoiceGroup.Team",
  step2_teamOrIndividualChoiceGroup_Individual:
    "OnboardingSurveyScreen.step2.teamOrIndividualChoiceGroup.Individual",
  step3_team_companySizeChoiceGroup_label:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.label",
  ["step3_team_companySizeChoiceGroup_1-49"]:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.1-49",
  ["step3_team_companySizeChoiceGroup_50-99"]:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.50-99",
  ["step3_team_companySizeChoiceGroup_100-499"]:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.100-499",
  ["step3_team_companySizeChoiceGroup_500-1999"]:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.500-1999",
  ["step3_team_companySizeChoiceGroup_2000+"]:
    "OnboardingSurveyScreen.step3-team.companySizeChoiceGroup.2000+",
  step4_reasonChoiceGroup_Auth_title:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Auth.title",
  step4_reasonChoiceGroup_Auth_subtitle:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Auth.subtitle",
  step4_reasonChoiceGroup_SSO_title:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.SSO.title",
  step4_reasonChoiceGroup_SSO_subtitle:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.SSO.subtitle",
  step4_reasonChoiceGroup_Security_title:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Security.title",
  step4_reasonChoiceGroup_Security_subtitle:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Security.subtitle",
  step4_reasonChoiceGroup_Portal_title:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Portal.title",
  step4_reasonChoiceGroup_Portal_subtitle:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Portal.subtitle",
  step4_reasonChoiceGroup_Other_title:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Other.title",
  step4_reasonChoiceGroup_Other_subtitle:
    "OnboardingSurveyScreen.step4.reasonChoiceGroup.Other.subtitle",
};
const localStorageKey = "authgear-onboarding-survey";

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
      className={cn(checked ? "ring-2" : "ring-1", styles.CompoundChoiceButton)}
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
        const key = (prefix +
          "_" +
          choice) as keyof typeof buttonTranslationKeys;
        return (
          <ChoiceButton
            key={key}
            text={renderToString(buttonTranslationKeys[key])}
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
        const key = (prefix +
          "_" +
          choice) as keyof typeof buttonTranslationKeys;
        return (
          <ChoiceButton
            key={key}
            text={renderToString(buttonTranslationKeys[key])}
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
  const labelKey = (prefix + "_label") as keyof typeof buttonTranslationKeys;
  return (
    <div>
      <Label>{renderToString(buttonTranslationKeys[labelKey])}</Label>
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
        const titleKey = (prefix +
          "_" +
          choice +
          "_title") as keyof typeof buttonTranslationKeys;
        const subtitleKey = (prefix +
          "_" +
          choice +
          "_subtitle") as keyof typeof buttonTranslationKeys;
        return (
          <CompoundChoiceButton
            key={titleKey}
            text={renderToString(buttonTranslationKeys[titleKey])}
            secondaryText={renderToString(buttonTranslationKeys[subtitleKey])}
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

interface LocallyStoredData {
  roleChoices?: string;
  toriChoices?: string;
  companyName?: string;
  companySize?: string;
  companyPhone?: PhoneTextFieldValues;
  individualWebsite?: string;
  individualPhone?: PhoneTextFieldValues;
  reasonChoices?: string[];
  otherReason?: string;
}

function getFromLocalStorage(
  prop: keyof LocallyStoredData
): string | string[] | PhoneTextFieldValues | undefined {
  const locallyStoredData = localStorage.getItem(localStorageKey);
  let localJson: LocallyStoredData = {};
  if (locallyStoredData === null) return undefined;
  localJson = JSON.parse(locallyStoredData);
  return localJson[prop];
}

function setLocalStorage(prop: keyof LocallyStoredData, value: any): void {
  const locallyStoredData = localStorage.getItem(localStorageKey);
  let localJson: LocallyStoredData = {};
  if (locallyStoredData) localJson = JSON.parse(locallyStoredData);
  localJson[prop] = value;
  localStorage.setItem(localStorageKey, JSON.stringify(localJson));
}

function goToFirstUnfilled(
  currentStep: number,
  navigate: NavigateFunction
): void {
  if (currentStep > 1 && getFromLocalStorage("roleChoices") === undefined) {
    navigate("./../1");
  }
  if (currentStep > 2 && getFromLocalStorage("toriChoices") === undefined) {
    navigate("./../2");
  }
  if (
    currentStep > 3 &&
    getFromLocalStorage("toriChoices") === "Team" &&
    (getFromLocalStorage("companyName") === undefined ||
      getFromLocalStorage("companySize") === undefined)
  ) {
    navigate("./../3");
  }
}

interface StepProps {}

function Step1(_props: StepProps) {
  const prefix = "step1";
  const roleChoiceGroup = "roleChoiceGroup";
  const roleChoices = ["Dev", "IT", "PM", "PD", "Market", "Owner", "Other"];
  const roleChoicesFromLocalStorage = getFromLocalStorage("roleChoices");
  const defaultRoleChoicesState: string[] = (
    roleChoicesFromLocalStorage === undefined
      ? []
      : [roleChoicesFromLocalStorage]
  ) as string[];
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
      setLocalStorage("roleChoices", roleChoicesState[0]);
      navigate("./../2");
    },
    [navigate, roleChoicesState]
  );
  return (
    <SurveyLayout
      contentClassName={styles.step1Content}
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
        prefix={[prefix, roleChoiceGroup].join("_")}
        availableChoices={roleChoices}
        selectedChoices={roleChoicesState}
        setChoice={setRoleChoicesState}
      />
    </SurveyLayout>
  );
}

function Step2(_props: StepProps) {
  const prefix = "step2";
  const toriChoiceGroup = "teamOrIndividualChoiceGroup";
  const toriChoices = ["Team", "Individual"];
  const toriChoicesFromLocalStorage = getFromLocalStorage("toriChoices");
  const defaultToriChoices: string[] = (
    toriChoicesFromLocalStorage === undefined
      ? []
      : [toriChoicesFromLocalStorage]
  ) as string[];
  const [toriChoicesState, setToriChoicesState] = useState(defaultToriChoices);
  const empty = useMemo(() => {
    return toriChoicesState.length === 0;
  }, [toriChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  useEffect(() => goToFirstUnfilled(2, navigate), [navigate]);
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      setLocalStorage("toriChoices", toriChoicesState[0]);
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
      contentClassName={styles.step2Content}
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
        prefix={[prefix, toriChoiceGroup].join("_")}
        availableChoices={toriChoices}
        selectedChoices={toriChoicesState}
        setChoice={setToriChoicesState}
      />
    </SurveyLayout>
  );
}

function Step3Team(_props: StepProps) {
  const prefix = "step3_team";
  const companyNameFromLocalStorage = getFromLocalStorage("companyName");
  const [companyName, setCompanyName] = useState(
    companyNameFromLocalStorage === undefined
      ? ""
      : (companyNameFromLocalStorage as string)
  );
  const companyPhoneFromLocalStorage = getFromLocalStorage("companyPhone");
  const defaultPhone: PhoneTextFieldValues = (
    companyPhoneFromLocalStorage === undefined
      ? { rawInputValue: "" }
      : companyPhoneFromLocalStorage
  ) as PhoneTextFieldValues;
  const [companyPhone, setCompanyPhone] = useState(defaultPhone);
  const companySizeChoiceGroup = "companySizeChoiceGroup";
  const companySizeChoices = ["1-49", "50-99", "100-499", "500-1999", "2000+"];
  const defaultCompanySizeFromLocalStorage = getFromLocalStorage("companySize");
  const defaultCompanySize: string[] = (
    defaultCompanySizeFromLocalStorage === undefined
      ? []
      : [defaultCompanySizeFromLocalStorage]
  ) as string[];
  const [companySizeChoicesState, setCompanySizeChoicesState] =
    useState(defaultCompanySize);
  const companySizeEmpty = useMemo(() => {
    return companySizeChoicesState.length === 0;
  }, [companySizeChoicesState]);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  useEffect(() => {
    goToFirstUnfilled(3, navigate);
    if (getFromLocalStorage("toriChoices") === "Individual")
      navigate("./../3-individual");
  }, [navigate]);
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      setLocalStorage("companyName", companyName);
      setLocalStorage("companySize", companySizeChoicesState[0]);
      if (companyPhone.rawInputValue !== "")
        setLocalStorage("companyPhone", companyPhone);
      navigate("./../4");
    },
    [navigate, companyName, companySizeChoicesState, companyPhone]
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
      contentClassName={styles.step3Content}
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
            prefix={[prefix, companySizeChoiceGroup].join("_")}
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
  const individualWebsiteFromLocalStorage =
    getFromLocalStorage("individualWebsite");
  const [individualWebsite, setIndividualWebsite] = useState(
    individualWebsiteFromLocalStorage === undefined
      ? ""
      : (individualWebsiteFromLocalStorage as string)
  );
  const individualPhoneFromLocalStorage = getFromLocalStorage("companyPhone");
  const defaultPhone: PhoneTextFieldValues = (
    individualPhoneFromLocalStorage === undefined
      ? { rawInputValue: "" }
      : individualPhoneFromLocalStorage
  ) as PhoneTextFieldValues;
  const [individualPhone, setIndividualPhone] = useState(defaultPhone);
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  useEffect(() => {
    goToFirstUnfilled(3, navigate);
    if (getFromLocalStorage("toriChoices") === "Team") navigate("./../3-team");
  }, [navigate]);
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (individualWebsite !== "")
        setLocalStorage("individualWebsite", individualWebsite);
      if (individualPhone.rawInputValue !== "")
        setLocalStorage("individualPhone", individualPhone);
      navigate("./../4");
    },
    [navigate, individualWebsite, individualPhone]
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
      contentClassName={styles.step3Content}
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
  const prefix = "step4";
  const reasonChoiceGroup = "reasonChoiceGroup";
  const reasonChoices = ["Auth", "SSO", "Security", "Portal", "Other"];
  const reasonChoicesFromLocalStorage = getFromLocalStorage("reasonChoices");
  const defaultReasonChoices: string[] = (
    reasonChoicesFromLocalStorage === undefined
      ? []
      : reasonChoicesFromLocalStorage
  ) as string[];
  const [reasonChoicesState, setReasonChoicesState] =
    useState(defaultReasonChoices);
  const empty = useMemo(() => {
    return reasonChoicesState.length === 0;
  }, [reasonChoicesState]);
  const { renderToString } = useContext(Context);
  const otherReasonFromLocalStorage = getFromLocalStorage("otherReason");
  const [otherReason, setOtherReason] = useState(
    otherReasonFromLocalStorage === undefined
      ? ""
      : (otherReasonFromLocalStorage as string)
  );
  const navigate = useNavigate();
  useEffect(() => goToFirstUnfilled(4, navigate), [navigate]);
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      setLocalStorage("reasonChoices", reasonChoicesState);
      if (otherReason !== "") setLocalStorage("otherReason", otherReason);
      const companyName = getFromLocalStorage("companyName");
      localStorage.removeItem(localStorageKey);
      if (companyName !== undefined)
        navigate("./../../projects/create", {
          state: { company_name: companyName },
        });
      else navigate("./../../projects/create");
    },
    [navigate, reasonChoicesState, otherReason]
  );
  const onClickBack = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate(
        getFromLocalStorage("toriChoices") === "Team"
          ? "./../3-team"
          : "./../3-individual"
      );
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
      contentClassName={styles.step4Content}
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
      <div className={styles.step4}>
        <MultiChoiceButtonGroup
          prefix={[prefix, reasonChoiceGroup].join("_")}
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
        <Route path="/4" element={<Step4 />} />
        <Route path="*" element={<Navigate to="1" replace={true} />} />
      </Routes>
    );
  };

export default OnboardingSurveyScreen;
