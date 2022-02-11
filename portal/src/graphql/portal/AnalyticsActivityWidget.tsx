import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { IPivotItemProps, Pivot, PivotItem, Text } from "@fluentui/react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  PointElement,
  LineElement,
} from "chart.js";
import { Bar, Line } from "react-chartjs-2";
import {
  AnalyticChartsQuery_activeUserChart,
  AnalyticChartsQuery_totalUserCountChart,
} from "./query/__generated__/AnalyticChartsQuery";
import { Periodical } from "./__generated__/globalTypes";
import { isoWeekLabel, monthLabel } from "../../util/date";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ShowLoading from "../../ShowLoading";
import styles from "./AnalyticsActivityWidget.module.scss";

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip);
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip
);

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
    return chartData ? (
      <Bar className={styles.chart} options={options} data={data} />
    ) : (
      <></>
    );
  };

interface AnalyticsActivityWidgetTotalUserChartProps {
  chartData: AnalyticChartsQuery_totalUserCountChart | null;
}

const AnalyticsActivityWidgetTotalUserChart: React.FC<AnalyticsActivityWidgetTotalUserChartProps> =
  function AnalyticsActivityWidgetTotalUserChart(props) {
    const { renderToString } = useContext(Context);
    const { chartData } = props;
    const options = {
      responsive: true,
      scales: {
        y: {
          title: {
            display: true,
            text: renderToString("AnalyticsActivityWidget.total-user.label"),
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
      return {
        labels: chartData?.dataset.map((pt) => (pt ? pt.label : "")),
        datasets: [
          {
            label: renderToString("AnalyticsActivityWidget.total-user.label"),
            data: chartData?.dataset.map((pt) => pt?.data),
            borderColor: "#176DF3",
            backgroundColor: "#176DF3",
          },
        ],
      };
    }, [chartData, renderToString]);
    return chartData ? (
      <Line className={styles.chart} options={options} data={data} />
    ) : (
      <></>
    );
  };

const AnalyticsActivityCharts: React.FC<AnalyticsActivityWidgetProps> =
  function AnalyticsActivityCharts(props) {
    const totalNumberOfUser = useMemo(() => {
      const dataset = props.totalUserCountChartData?.dataset;
      if (dataset == null || dataset.length === 0) {
        return "-";
      }
      return dataset[dataset.length - 1]?.data;
    }, [props.totalUserCountChartData]);

    if (props.loading) {
      return (
        <div className={styles.loadingWrapper}>
          <ShowLoading />
        </div>
      );
    }

    return (
      <>
        <AnalyticsActivityWidgetActiveUserChart
          chartData={props.activeUserChartData}
          periodical={props.periodical}
        />
        <div className={styles.totalUserLabel}>
          <Text variant="medium" block={true}>
            <FormattedMessage id="AnalyticsActivityWidget.total-user.label" />
          </Text>
          <Text variant="xLarge" block={true}>
            {totalNumberOfUser}
          </Text>
        </div>
        <AnalyticsActivityWidgetTotalUserChart
          chartData={props.totalUserCountChartData}
        />
      </>
    );
  };

interface AnalyticsActivityWidgetProps {
  loading: boolean;
  periodical: Periodical;
  onPeriodicalChange: (periodical: Periodical) => void;
  activeUserChartData: AnalyticChartsQuery_activeUserChart | null;
  totalUserCountChartData: AnalyticChartsQuery_totalUserCountChart | null;
}

const AnalyticsActivityWidget: React.FC<AnalyticsActivityWidgetProps> =
  function AnalyticsActivityWidget(props) {
    const { renderToString } = useContext(Context);
    const { periodical, onPeriodicalChange } = props;
    const onPeriodicalClick = useCallback(
      (item?: { props: IPivotItemProps }) => {
        const itemKey = item?.props.itemKey;
        if (itemKey) {
          if (itemKey !== periodical) {
            if (itemKey in Periodical) {
              onPeriodicalChange(itemKey as Periodical);
            }
          }
        }
      },
      [periodical, onPeriodicalChange]
    );
    return (
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsActivityWidget.title" />
        </WidgetTitle>
        <Pivot
          className={styles.pivot}
          onLinkClick={onPeriodicalClick}
          selectedKey={periodical}
        >
          <PivotItem
            headerText={renderToString("AnalyticsActivityWidget.monthly.label")}
            itemKey={Periodical.MONTHLY}
          >
            <AnalyticsActivityCharts {...props} />
          </PivotItem>
          <PivotItem
            headerText={renderToString("AnalyticsActivityWidget.weekly.label")}
            itemKey={Periodical.WEEKLY}
          >
            <AnalyticsActivityCharts {...props} />
          </PivotItem>
        </Pivot>
      </Widget>
    );
  };

export default AnalyticsActivityWidget;
