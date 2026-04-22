import React, { useMemo } from "react";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  IChoiceGroupProps,
  IChoiceGroupStyles,
  Text,
} from "@fluentui/react";
import styles from "./ChoiceGroupWithDescriptions.module.css";

export interface ChoiceGroupWithDescriptionOption extends IChoiceGroupOption {
  description?: React.ReactNode;
}

export interface ChoiceGroupWithDescriptionsProps
  extends Omit<IChoiceGroupProps, "options"> {
  options: ChoiceGroupWithDescriptionOption[];
}

export default function ChoiceGroupWithDescriptions(
  props: ChoiceGroupWithDescriptionsProps
): React.ReactElement {
  const { options, styles: choiceGroupStylesProp, ...restProps } = props;

  const choiceGroupStyles: Partial<IChoiceGroupStyles> = useMemo(
    () => ({
      flexContainer: {
        selectors: {
          ".ms-ChoiceField": {
            display: "block",
          },
        },
      },
    }),
    []
  );

  const normalizedOptions = useMemo<IChoiceGroupOption[]>(() => {
    return options.map((option) => {
      if (option.description == null) {
        return option;
      }
      return {
        ...option,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderField: (
          fieldProps?: IChoiceGroupOption & IChoiceGroupOptionProps,
          render?: (
            props?: IChoiceGroupOption & IChoiceGroupOptionProps
          ) => JSX.Element | null
        ) => {
          return (
            <>
              {render?.(fieldProps)}
              <Text variant="small" block={true} className={styles.description}>
                {option.description}
              </Text>
            </>
          );
        },
      };
    });
  }, [options]);

  return (
    <ChoiceGroup
      {...restProps}
      options={normalizedOptions}
      styles={choiceGroupStylesProp ?? choiceGroupStyles}
    />
  );
}
