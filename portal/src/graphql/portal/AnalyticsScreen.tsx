import React, { useState } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useAnalyticChartsQuery } from "./query/analyticChartsQuery";
import { Periodical } from "./__generated__/globalTypes";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import AnalyticsActivityWidget from "./AnalyticsActivityWidget";
import ShowError from "../../ShowError";
import styles from "./AnalyticsScreen.module.scss";

const AnalyticsScreen: React.FC = function AnalyticsScreen() {
  // FIXME: support user input
  const to = new Date();
  const from = new Date();
  from.setFullYear(from.getFullYear() - 1);
  const rangeTo = to.toISOString().split("T")[0];
  const rangeFrom = from.toISOString().split("T")[0];

  const [periodical, setPeriodical] = useState<Periodical>(Periodical.MONTHLY);

  const { appID } = useParams();
  const { loading, error, refetch, activeUserChart, totalUserCountChart } =
    useAnalyticChartsQuery(appID, periodical, rangeFrom, rangeTo);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <ScreenContent>
      <ScreenTitle>
        <FormattedMessage id="AnalyticsScreen.title" />
      </ScreenTitle>
      <AnalyticsActivityWidget
        loading={loading}
        periodical={periodical}
        onPeriodicalChange={setPeriodical}
        activeUserChartData={activeUserChart}
        totalUserCountChartData={totalUserCountChart}
      />
    </ScreenContent>
  );
};

export default AnalyticsScreen;
