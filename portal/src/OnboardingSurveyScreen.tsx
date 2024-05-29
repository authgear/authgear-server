import React, { ComponentType, useContext, useState, useMemo } from "react";
import { Routes, Navigate, Route } from "react-router-dom";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import SurveyLayout from "./OnboardingSurveyLayout";
import styles from "./OnboardingSurveyScreen.module.css";

interface ButtonProps {
  buttonKey: string;
  pressed: boolean;
  onClick: () => void;
}

function ChoiceButton(props: ButtonProps) {
  const { buttonKey, pressed, onClick } = props;
  const { renderToString } = useContext(Context);
  return (
    <DefaultButton
      text={renderToString(buttonKey)}
      toggle={true}
      checked={pressed}
      className={pressed ? styles.pressedChoiceButton : styles.choiceButton}
      onClick={onClick}
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
type Statify<K extends Readonly<string[]>> = K extends [
  infer F extends string,
  ...infer R extends string[]
]
  ? Statify<R> & { [key in F]: boolean }
  : object;

interface ChoiceButtonGroupProps<Choices extends Readonly<string[]>> {
  prefix: string;
  state: Statify<Choices>;
  setChoice: (choice: Statify<Choices>) => void;
  processChoice: (
    oldChoice: Statify<Choices>,
    newChoice: Statify<Choices>,
    selectedChoice: keyof Statify<Choices>
  ) => Statify<Choices>;
  Button: ComponentType<ButtonProps>;
}

function ChoiceButtonGroup<Choices extends Readonly<string[]>>(
  props: ChoiceButtonGroupProps<Choices>
) {
  const { prefix, state, setChoice, processChoice, Button } = props;
  const buttons = useMemo(
    () =>
      (Object.keys(state) as (keyof Statify<Choices>)[]).map((choice) => {
        return (
          <Button
            buttonKey={[prefix, choice as string].join(".")}
            key={prefix + (choice as string)}
            pressed={state[choice] as boolean}
            onClick={() => {
              const newState = state;
              newState[choice] = !state[
                choice
              ] as Statify<Choices>[keyof Statify<Choices>];
              setChoice(
                structuredClone(processChoice(state, newState, choice))
              );
            }}
          />
        );
      }),
    [state, Button, prefix, processChoice, setChoice]
  );
  return <div className={styles.singleChoiceButtonGroup}>{buttons}</div>;
}

interface StepProps {}

function processSingleChoice<K extends Readonly<string[]>>(
  oldChoice: Statify<K>,
  newChoice: Statify<K>,
  choice: keyof Statify<K>
): Statify<K> {
  const keys = Object.keys(oldChoice) as (keyof Statify<K>)[];
  let newCount = 0;
  for (const key of keys) newCount += (newChoice[key] as boolean) ? 1 : 0;
  if (newCount === 2)
    for (const key of keys)
      newChoice[key] = (key === choice) as Statify<K>[keyof Statify<K>];
  return newChoice;
}

function Step1(_props: StepProps) {
  const prefix = "OnboardingSurveyScreen.step1";
  const roleChoiceGroup = "roleChoiceGroup";
  type RoleChoices = ["Dev", "IT", "PM", "PD", "Market", "Owner", "Other"];
  const defaultRoleChoices: Statify<RoleChoices> = {
    Dev: false,
    IT: false,
    PM: false,
    PD: false,
    Market: false,
    Owner: false,
    Other: false,
  };
  const [roleChoicesState, setRoleChoicesState] = useState(defaultRoleChoices);
  const empty = useMemo(() => {
    let newCount = 0;
    for (const key of Object.keys(
      roleChoicesState
    ) as (keyof typeof roleChoicesState)[])
      newCount += roleChoicesState[key] ? 1 : 0;
    return newCount === 0;
  }, [roleChoicesState]);
  const { renderToString } = useContext(Context);
  const onClickNext = () => {};
  const onClickPrev = () => {};
  return (
    <SurveyLayout
      title={renderToString(prefix + ".title")}
      subtitle={renderToString(prefix + ".subtitle")}
      backButtonDisabled={true}
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
          onClick={onClickPrev}
          text={<FormattedMessage id="prev" />}
          className={styles.prevButton}
        />
      }
    >
      <ChoiceButtonGroup<RoleChoices>
        prefix={[prefix, roleChoiceGroup].join(".")}
        state={roleChoicesState}
        setChoice={setRoleChoicesState}
        processChoice={processSingleChoice<RoleChoices>}
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
        <Route path="*" element={<Navigate to="1" replace={true} />} />
      </Routes>
    );
  };

export default OnboardingSurveyScreen;
