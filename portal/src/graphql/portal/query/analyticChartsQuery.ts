import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { client } from "../apollo";
import { Periodical } from "../__generated__/globalTypes";
import {
  AnalyticChartsQuery,
  AnalyticChartsQueryVariables,
  AnalyticChartsQuery_activeUserChart,
  AnalyticChartsQuery_totalUserCountChart,
} from "./__generated__/AnalyticChartsQuery";

export const analyticChartsQuery = gql`
  query AnalyticChartsQuery(
    $appID: ID!
    $periodical: Periodical!
    $rangeFrom: Date!
    $rangeTo: Date!
  ) {
    activeUserChart(
      appID: $appID
      periodical: $periodical
      rangeFrom: $rangeFrom
      rangeTo: $rangeTo
    ) {
      dataset {
        label
        data
      }
    }
    totalUserCountChart(
      appID: $appID
      rangeFrom: $rangeFrom
      rangeTo: $rangeTo
    ) {
      dataset {
        label
        data
      }
    }
  }
`;

export interface AnalyticChartsQueryResult
  extends Pick<
    QueryResult<AnalyticChartsQuery, AnalyticChartsQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  activeUserChart: AnalyticChartsQuery_activeUserChart | null;
  totalUserCountChart: AnalyticChartsQuery_totalUserCountChart | null;
}
export const useAnalyticChartsQuery = (
  appID: string,
  periodical: Periodical,
  rangeFrom: string,
  rangeTo: string
): AnalyticChartsQueryResult => {
  const { data, loading, error, refetch } = useQuery<AnalyticChartsQuery>(
    analyticChartsQuery,
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
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
