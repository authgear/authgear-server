import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "../../intl";
import { SegmentedControl, Spinner, Text } from "@radix-ui/themes";
import { TooltipItem } from "chart.js";
import { Bar, Line } from "react-chartjs-2";
import { AnalyticChartsQueryQuery } from "./query/analyticChartsQuery.generated";
import { Periodical } from "./globalTypes.generated";
import { isoWeekLabels, monthLabel } from "../../util/date";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./AnalyticsActivityWidget.module.css";

interface AnalyticsActivityWidgetActiveUserChartProps {
  chartData: AnalyticChartsQueryQuery["activeUserChart"] | null;
  periodical: Periodical;
}

const AnalyticsActivityWidgetActiveUserChart: React.VFC<AnalyticsActivityWidgetActiveUserChartProps> =
  function AnalyticsActivityWidgetActiveUserChart(props) {
    const { renderToString } = useContext(Context);
    const { chartData, periodical } = props;
    const options = {
      maintainAspectRatio: false,
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
      plugins: {
        tooltip: {
          callbacks: {
            title: function (tooltipItem: TooltipItem<"bar">[]) {
              const item = tooltipItem[0];
              const dataLabels = item.chart.data.labels;
              if (dataLabels) {
                const labels = dataLabels[item.dataIndex];
                // join multiple line labels to one line in the tooltip title
                if (Array.isArray(labels)) {
                  return labels.join(" ");
                }
              }
              return tooltipItem[0].label;
            },
          },
        },
      },
    };
    const data = useMemo(() => {
      let labelFn = (iosDate: any) => iosDate;
      switch (periodical) {
        case Periodical.Monthly:
          labelFn = monthLabel;
          break;
        case Periodical.Weekly:
          labelFn = isoWeekLabels;
          break;
      }

      return {
        labels: chartData?.dataset.map((pt) => (pt ? labelFn(pt.label) : "")),
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
      <div className={styles.chartContainer}>
        <Bar options={options} data={data} />
      </div>
    ) : (
      <></>
    );
  };

interface AnalyticsActivityWidgetTotalUserChartProps {
  chartData: AnalyticChartsQueryQuery["totalUserCountChart"] | null;
}

const AnalyticsActivityWidgetTotalUserChart: React.VFC<AnalyticsActivityWidgetTotalUserChartProps> =
  function AnalyticsActivityWidgetTotalUserChart(props) {
    const { renderToString } = useContext(Context);
    const { chartData } = props;
    const options = {
      maintainAspectRatio: false,
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
      <div className={styles.chartContainer}>
        <Line options={options} data={data} />
      </div>
    ) : (
      <></>
    );
  };

const AnalyticsActivityCharts: React.VFC<AnalyticsActivityWidgetProps> =
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
          <Spinner size="3" />
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
          <Text as="p" size="2" className={styles.metricLabel}>
            <FormattedMessage id="AnalyticsActivityWidget.total-user.label" />
          </Text>
          <Text as="p" size="6" weight="medium" className={styles.metricValue}>
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
  className?: string;
  loading: boolean;
  periodical: Periodical;
  onPeriodicalChange: (periodical: Periodical) => void;
  activeUserChartData: AnalyticChartsQueryQuery["activeUserChart"] | null;
  totalUserCountChartData:
    | AnalyticChartsQueryQuery["totalUserCountChart"]
    | null;
}

const AnalyticsActivityWidget: React.VFC<AnalyticsActivityWidgetProps> =
  function AnalyticsActivityWidget(props) {
    const { renderToString } = useContext(Context);
    const { periodical, onPeriodicalChange } = props;
    const onPeriodicalValueChange = useCallback(
      (value: string) => {
        if (value !== periodical) {
          if (Object.values(Periodical).includes(value as Periodical)) {
            onPeriodicalChange(value as Periodical);
          }
        }
      },
      [periodical, onPeriodicalChange]
    );
    return (
      <SettingsSectionCard
        className={props.className}
        layout="stacked"
        title={<FormattedMessage id="AnalyticsActivityWidget.title" />}
        contentClassName={styles.content}
      >
        <div className={styles.periodicalControl}>
          <SegmentedControl.Root
            size="2"
            value={periodical}
            onValueChange={onPeriodicalValueChange}
          >
            <SegmentedControl.Item value={Periodical.Monthly}>
              {renderToString("AnalyticsActivityWidget.monthly.label")}
            </SegmentedControl.Item>
            <SegmentedControl.Item value={Periodical.Weekly}>
              {renderToString("AnalyticsActivityWidget.weekly.label")}
            </SegmentedControl.Item>
          </SegmentedControl.Root>
        </div>
        <AnalyticsActivityCharts {...props} />
      </SettingsSectionCard>
    );
  };

export default AnalyticsActivityWidget;
