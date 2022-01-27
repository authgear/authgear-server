import React, { useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
} from "chart.js";
import { Bar } from "react-chartjs-2";
import {
  AnalyticChartsQuery_activeUserChart,
  AnalyticChartsQuery_totalUserCountChart,
} from "./query/__generated__/AnalyticChartsQuery";
import { Periodical } from "./__generated__/globalTypes";
import { isoWeekLabel, monthLabel } from "../../util/date";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import styles from "./AnalyticsActivityWidget.module.scss";

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip);

interface AnalyticsActivityWidgetActiveUserChartProps {
  chartData: AnalyticChartsQuery_activeUserChart | null;
  periodical: Periodical;
}

const AnalyticsActivityWidgetActiveUserChart: React.FC<AnalyticsActivityWidgetActiveUserChartProps> =
  function AnalyticsActivityWidgetActiveUserChart(props) {
    const { renderToString } = useContext(Context);
    const { chartData, periodical } = props;
    const options = {
      responsive: true,
      scales: {
        y: {
          title: {
            display: true,
            text: renderToString("AnalyticsActivityWidget.active-user.label"),
          },
          min: 0,
        },
        x: {
          ticks: {
            maxTicksLimit: 12,
          },
        },
      },
    };
    const data = useMemo(() => {
      let labalFn = (iosDate: string) => iosDate;
      switch (periodical) {
        case Periodical.MONTHLY:
          labalFn = monthLabel;
          break;
        case Periodical.WEEKLY:
          labalFn = isoWeekLabel;
          break;
      }

      return {
        labels: chartData?.dataset.map((pt) => (pt ? labalFn(pt.label) : "")),
        datasets: [
          {
            label: renderToString("AnalyticsActivityWidget.active-user.label"),
            data: chartData?.dataset.map((pt) => pt?.data),
            backgroundColor: "#176DF3",
          },
        ],
      };
    }, [chartData, periodical, renderToString]);
    return chartData ? <Bar options={options} data={data} /> : <></>;
  };
interface AnalyticsActivityWidgetProps {
  periodical: Periodical;
  activeUserChartData: AnalyticChartsQuery_activeUserChart | null;
  totalUserCountChartData: AnalyticChartsQuery_totalUserCountChart | null;
}

const AnalyticsActivityWidget: React.FC<AnalyticsActivityWidgetProps> =
  function AnalyticsActivityWidget(props) {
    return (
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsActivityWidget.title" />
        </WidgetTitle>
        <>
          <AnalyticsActivityWidgetActiveUserChart
            chartData={props.activeUserChartData}
            periodical={props.periodical}
          />
        </>
      </Widget>
    );
  };

export default AnalyticsActivityWidget;
