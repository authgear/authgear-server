import { useCallback, useMemo } from "react";
import { IStyleFunctionOrObject } from "@fluentui/react";
import {
  concatStyleSetsWithProps,
  concatStyleSets,
  IConcatenatedStyleSet,
  IStyleSetBase,
} from "@fluentui/merge-styles";

export function useMergedStyles<TStylesProps, TStyleSet extends IStyleSetBase>(
  ...styless: (IStyleFunctionOrObject<TStylesProps, TStyleSet> | undefined)[]
): (
  props: TStylesProps
) => ReturnType<typeof concatStyleSetsWithProps<TStylesProps, TStyleSet>> {
  return useCallback(
    (props) => {
      return concatStyleSetsWithProps(props, ...styless);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [...styless]
  );
}

export function useMergedStylesPlain<TStyleSet extends IStyleSetBase>(
  ...styless: (TStyleSet | undefined)[]
): IConcatenatedStyleSet<TStyleSet> {
  return useMemo(
    () => concatStyleSets(...styless),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [...styless]
  ) as IConcatenatedStyleSet<TStyleSet>;
}
