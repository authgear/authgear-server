import React, { useContext, useMemo, useState } from "react";
import { CommandBarButton, ICommandBarItemProps } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useAnalyticChartsQuery } from "./query/analyticChartsQuery";
import { Periodical } from "./__generated__/globalTypes";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import AnalyticsActivityWidget from "./AnalyticsActivityWidget";
import ShowError from "../../ShowError";
import CommandBarContainer from "../../CommandBarContainer";
import styles from "./AnalyticsScreen.module.scss";

const CommandBarLabelValue = (label: string, value: string) => {
  return (props: ICommandBarItemProps) => {
    const { commandBarButtonProps } = props;
    return (
      <CommandBarButton
        className={styles.commandBarButtonLabelValue}
        {...commandBarButtonProps}
      >
        <span className={styles.label}>{label}</span>
        <span className={styles.value}>{value}</span>
      </CommandBarButton>
    );
  };
};

const OnRenderCommandBarToLabel = () => {
  return (
    <div className={styles.commandBarButtonTo}>
      <FormattedMessage id="AnalyticsScreen.to.label"></FormattedMessage>
    </div>
  );
};

const AnalyticsScreen: React.FC = function AnalyticsScreen() {
  // FIXME: support user input
  const to = new Date();
  const from = new Date();
  from.setFullYear(from.getFullYear() - 1);
  const rangeTo = to.toISOString().split("T")[0];
  const rangeFrom = from.toISOString().split("T")[0];

  const [periodical, setPeriodical] = useState<Periodical>(Periodical.MONTHLY);

  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const { loading, error, refetch, activeUserChart, totalUserCountChart } =
    useAnalyticChartsQuery(appID, periodical, rangeFrom, rangeTo);

  const primaryItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "startDate",
        commandBarButtonAs: CommandBarLabelValue(
          renderToString("AnalyticsScreen.start-date.label"),
          rangeFrom
        ),
        commandBarButtonProps: {
          iconProps: { iconName: "Calendar" },
          onClick: () => {},
        },
      },
      {
        key: "to",
        onRender: OnRenderCommandBarToLabel,
      },
      {
        key: "endDate",
        commandBarButtonAs: CommandBarLabelValue(
          renderToString("AnalyticsScreen.end-date.label"),
          rangeTo
        ),
        commandBarButtonProps: {
          iconProps: { iconName: "Calendar" },
          onClick: () => {},
        },
      },
    ];
  }, [renderToString, rangeFrom, rangeTo]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <>
      <CommandBarContainer isLoading={loading} primaryItems={primaryItems}>
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
      </CommandBarContainer>
    </>
  );
};

export default AnalyticsScreen;
