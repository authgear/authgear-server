import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { usePortalClient } from "../apollo";
import { Periodical } from "../globalTypes.generated";
import {
  AnalyticChartsQueryQuery,
  AnalyticChartsQueryQueryVariables,
  AnalyticChartsQueryDocument,
} from "./analyticChartsQuery.generated";

export interface AnalyticChartsQueryResult
  extends Pick<
    QueryResult<AnalyticChartsQueryQuery, AnalyticChartsQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  activeUserChart: AnalyticChartsQueryQuery["activeUserChart"] | null;
  totalUserCountChart: AnalyticChartsQueryQuery["totalUserCountChart"] | null;
  signupConversionRate: AnalyticChartsQueryQuery["signupConversionRate"] | null;
  signupByMethodsChart: AnalyticChartsQueryQuery["signupByMethodsChart"] | null;
}
export const useAnalyticChartsQuery = (
  appID: string,
  periodical: Periodical,
  rangeFrom: string,
  rangeTo: string
): AnalyticChartsQueryResult => {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<AnalyticChartsQueryQuery>(
    AnalyticChartsQueryDocument,
    {
      client,
      variables: {
        appID: appID,
        periodical: periodical,
        rangeFrom: rangeFrom,
        rangeTo: rangeTo,
      },
    }
  );

  const queryData = useMemo(() => {
    return {
      activeUserChart: data?.activeUserChart ?? null,
      totalUserCountChart: data?.totalUserCountChart ?? null,
      signupConversionRate: data?.signupConversionRate ?? null,
      signupByMethodsChart: data?.signupByMethodsChart ?? null,
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
