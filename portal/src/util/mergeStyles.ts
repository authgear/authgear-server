import { useCallback, useMemo } from "react";
import { IStyleFunctionOrObject, IStyleSet } from "@fluentui/react";
import {
  concatStyleSetsWithProps,
  concatStyleSets,
  IConcatenatedStyleSet,
} from "@fluentui/merge-styles";

export function useMergedStyles<TStylesProps, IStyleSet>(
  ...styless: (IStyleFunctionOrObject<TStylesProps, IStyleSet> | undefined)[]
): IStyleFunctionOrObject<TStylesProps, IStyleSet> {
  return useCallback(
    (props) => {
      return concatStyleSetsWithProps(props, ...styless);
    },
    // eslint-disable-next-line
    [...styless]
  );
}

export function useMergedStylesPlain(
  // eslint-disable-next-line @typescript-eslint/no-redundant-type-constituents
  ...styless: (IStyleSet | undefined)[]
): IConcatenatedStyleSet<IStyleSet> {
  return useMemo(
    () => concatStyleSets(...styless),
    // eslint-disable-next-line
    [...styless]
  ) as IConcatenatedStyleSet<IStyleSet>;
}
