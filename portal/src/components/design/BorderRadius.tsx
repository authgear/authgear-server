import React, { useCallback, useContext, useEffect, useState } from "react";
import { SegmentedControl } from "@radix-ui/themes";
import { Context as MFContext } from "../../intl";
import cn from "classnames";

import {
  AllBorderRadiusStyleTypes,
  BorderRadiusStyle,
  BorderRadiusStyleType,
  DEFAULT_BORDER_RADIUS,
} from "../../model/themeAuthFlowV2";
import { TextField } from "../v2/TextField/TextField";
import { ErrorParseRule } from "../../error/parse";

import styles from "./BorderRadius.module.css";
import toggleStyles from "./toggle-group.module.css";

interface BorderRadiusIconProps {
  type: BorderRadiusStyleType;
  className?: string;
}

const BorderRadiusIcon: React.VFC<BorderRadiusIconProps> =
  function BorderRadiusIcon(props) {
    const { type, className } = props;
    const svgProps = {
      className: cn(styles.borderRadiusIcon, className),
      width: 16,
      height: 16,
      viewBox: "0 0 16 16",
      fill: "none",
      xmlns: "http://www.w3.org/2000/svg",
      "aria-hidden": true,
    } as const;

    switch (type) {
      case "none":
        return (
          <svg {...svgProps}>
            <rect
              width="16"
              height="16"
              fill="currentColor"
              fillOpacity="0.06"
            />
            <path d="M13 3H3V13" stroke="currentColor" />
          </svg>
        );
      case "rounded":
        return (
          <svg {...svgProps}>
            <path
              d="M3 7C3 4.79086 4.79086 3 7 3H13V13H3V7Z"
              fill="currentColor"
              fillOpacity="0.08"
            />
            <path
              d="M13 3H7C4.79086 3 3 4.79086 3 7V13"
              stroke="currentColor"
            />
          </svg>
        );
      case "rounded-full":
        return (
          <svg {...svgProps}>
            <path
              d="M3 13C3 7.47715 7.47715 3 13 3V3V13H3V13Z"
              fill="currentColor"
              fillOpacity="0.06"
            />
            <path d="M13 3C7.47715 3 3 7.47715 3 13" stroke="currentColor" />
          </svg>
        );
      default:
        return null;
    }
  };

interface BorderRadiusProps {
  value: BorderRadiusStyle;
  onChange: (value: BorderRadiusStyle) => void;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
  className?: string;
}

const BorderRadius: React.VFC<BorderRadiusProps> = function BorderRadius(
  props
) {
  const {
    value,
    onChange,
    parentJSONPointer,
    fieldName,
    errorRules,
    className,
  } = props;
  const { renderToString } = useContext(MFContext);

  const [radiusValue, setRadiusValue] = useState(() => {
    if (value.type !== "rounded") {
      return DEFAULT_BORDER_RADIUS;
    }
    return value.radius;
  });

  const valueRadiusValue = value.type === "rounded" ? value.radius : undefined;
  useEffect(() => {
    if (valueRadiusValue == null) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setRadiusValue(DEFAULT_BORDER_RADIUS);
    } else {
      setRadiusValue(valueRadiusValue);
    }
  }, [valueRadiusValue]);

  const onValueChange = useCallback(
    (nextType: string) => {
      const type = nextType as BorderRadiusStyleType;
      if (type === "rounded") {
        onChange({
          type,
          radius: radiusValue !== "" ? radiusValue : DEFAULT_BORDER_RADIUS,
        });
      } else {
        onChange({ type });
      }
    },
    [radiusValue, onChange]
  );

  const onBorderRadiusChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const nextValue = e.target.value;
      setRadiusValue(nextValue);
      if (nextValue !== "") {
        onChange({
          type: "rounded",
          radius: nextValue,
        });
      }
    },
    [onChange]
  );

  const onBorderRadiusBlur = useCallback(
    (ev: React.FocusEvent<HTMLInputElement>) => {
      if (ev.target.value === "") {
        setRadiusValue(DEFAULT_BORDER_RADIUS);
        onChange({
          type: "rounded",
          radius: DEFAULT_BORDER_RADIUS,
        });
        return;
      }
      setRadiusValue(ev.target.value);
      onChange({
        type: "rounded",
        radius: ev.target.value,
      });
    },
    [onChange]
  );

  return (
    <div className={className}>
      <SegmentedControl.Root
        className={toggleStyles.toggleGroup}
        value={value.type}
        onValueChange={onValueChange}
        size="1"
      >
        {AllBorderRadiusStyleTypes.map((type) => (
          <SegmentedControl.Item key={type} value={type}>
            <BorderRadiusIcon type={type} />
          </SegmentedControl.Item>
        ))}
      </SegmentedControl.Root>
      {value.type === "rounded" ? (
        <div className={styles.radiusValueField}>
          <TextField
            size="2"
            parentJSONPointer={parentJSONPointer}
            fieldName={fieldName}
            errorRules={errorRules}
            label={renderToString(
              "DesignScreen.configuration.borderRadius.label"
            )}
            value={radiusValue}
            onChange={onBorderRadiusChange}
            onBlur={onBorderRadiusBlur}
          />
        </div>
      ) : null}
    </div>
  );
};

export default BorderRadius;
