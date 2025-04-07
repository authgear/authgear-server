import React from "react";
import cn from "classnames";
import { CheckIcon } from "@radix-ui/react-icons";
import styles from "./Stepper.module.css";

export interface Step {
  text: React.ReactNode;
  checked: boolean;
}

export interface StepperProps {
  darkMode?: boolean;
  steps: Step[];
}

export function Stepper({ darkMode, steps }: StepperProps): React.ReactElement {
  return (
    <div className={cn(styles.stepper, darkMode ? "dark" : null)}>
      {steps.map((step, idx) => {
        const active = idx === 0 ? true : steps[idx - 1].checked;
        if (idx === steps.length - 1) {
          return (
            <StepIcon
              key={idx}
              checked={step.checked}
              active={active}
              text={step.text}
            />
          );
        }
        return (
          <React.Fragment key={idx}>
            <StepIcon checked={step.checked} active={active} text={step.text} />
            <StepTrail progress={step.checked ? 1 : active ? 0.5 : 0} />
          </React.Fragment>
        );
      })}
    </div>
  );
}

export interface StepIconProps {
  darkMode?: boolean;
  checked?: boolean;
  active?: boolean;
  text?: React.ReactNode;
}

export function StepIcon({
  darkMode,
  checked,
  active,
  text,
}: StepIconProps): React.ReactElement {
  return (
    <div
      className={cn(
        styles.stepIcon,
        active ? styles["stepIcon--active"] : null,
        darkMode ? "dark" : null
      )}
    >
      {checked ? <CheckIcon className={styles.stepIcon__checkIcon} /> : text}
    </div>
  );
}

export interface StepTrailProps {
  darkMode?: boolean;
  progress: 0 | 0.5 | 1;
}

export function StepTrail({
  darkMode,
  progress,
}: StepTrailProps): React.ReactElement {
  return (
    <div className={cn(styles.stepTrail, darkMode ? "dark" : null)}>
      <div
        className={styles.stepTrailFilling}
        style={{ right: `${Math.round((1 - progress) * 100)}%` }}
      />
    </div>
  );
}
