import { useCallback, useMemo } from "react";
import { IStyleFunctionOrObject, IStyleSet } from "@fluentui/react";
import {
  concatStyleSetsWithProps,
  concatStyleSets,
  IConcatenatedStyleSet,
} from "@fluentui/merge-styles";

export function useMergedStyles<
  TStylesProps,
  TStyleSet extends IStyleSet<TStyleSet>
>(
  ...styless: (IStyleFunctionOrObject<TStylesProps, TStyleSet> | undefined)[]
): IStyleFunctionOrObject<TStylesProps, TStyleSet> {
  return useCallback(
    (props) => {
      return concatStyleSetsWithProps(props, ...styless);
    },
    // eslint-disable-next-line
    [...styless]
  );
}

export function useMergedStylesPlain<TStyleSet extends IStyleSet<TStyleSet>>(
  ...styless: (TStyleSet | undefined)[]
): IConcatenatedStyleSet<TStyleSet> {
  return useMemo(
    () => concatStyleSets(...styless),
    // eslint-disable-next-line
    [...styless]
  ) as IConcatenatedStyleSet<TStyleSet>;
}
