/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { Periodical } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL query operation: AnalyticChartsQuery
// ====================================================

export interface AnalyticChartsQuery_activeUserChart_dataset {
  __typename: "DataPoint";
  label: string;
  data: number;
}

export interface AnalyticChartsQuery_activeUserChart {
  __typename: "Chart";
  dataset: (AnalyticChartsQuery_activeUserChart_dataset | null)[];
}

export interface AnalyticChartsQuery_totalUserCountChart_dataset {
  __typename: "DataPoint";
  label: string;
  data: number;
}

export interface AnalyticChartsQuery_totalUserCountChart {
  __typename: "Chart";
  dataset: (AnalyticChartsQuery_totalUserCountChart_dataset | null)[];
}

export interface AnalyticChartsQuery {
  /**
   * Active users chart dataset
   */
  activeUserChart: AnalyticChartsQuery_activeUserChart | null;
  /**
   * Total users count chart dataset
   */
  totalUserCountChart: AnalyticChartsQuery_totalUserCountChart | null;
}

export interface AnalyticChartsQueryVariables {
  appID: string;
  periodical: Periodical;
  rangeFrom: GQL_Date;
  rangeTo: GQL_Date;
}
