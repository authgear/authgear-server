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

export interface AnalyticChartsQuery_signupConversionRate {
  __typename: "SignupConversionRate";
  totalSignup: number;
  totalSignupUniquePageView: number;
}

export interface AnalyticChartsQuery_signupByMethodsChart_dataset {
  __typename: "DataPoint";
  label: string;
  data: number;
}

export interface AnalyticChartsQuery_signupByMethodsChart {
  __typename: "Chart";
  dataset: (AnalyticChartsQuery_signupByMethodsChart_dataset | null)[];
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
  /**
   * Signup conversion rate dashboard data
   */
  signupConversionRate: AnalyticChartsQuery_signupConversionRate | null;
  /**
   * Signup by methods dataset
   */
  signupByMethodsChart: AnalyticChartsQuery_signupByMethodsChart | null;
}

export interface AnalyticChartsQueryVariables {
  appID: string;
  periodical: Periodical;
  rangeFrom: GQL_Date;
  rangeTo: GQL_Date;
}
