import React, { useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Pie } from "react-chartjs-2";
import { Text } from "@fluentui/react";

import {
  AnalyticChartsQuery_signupByMethodsChart,
  AnalyticChartsQuery_signupByMethodsChart_dataset,
} from "./query/__generated__/AnalyticChartsQuery";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ShowLoading from "../../ShowLoading";
import styles from "./AnalyticsSignupMethodsWidget.module.scss";

const noDataPlaceholderColor = "#EAEAEA";
const colorMap: Record<string, string> = {
  email: "#FF7629",
  phone: "#92674F",
  username: "#FFB900",
  google: "#EA4335",
  facebook: "#4267B2",
  linkedin: "#4267B2",
  azureadv2: "#007FFF",
  adfs: "#49A8EC",
  apple: "#A2AAAD",
  wechat: "#45B049",
  anonymous: "#957AFF",
};

function getColorCodeByMethod(method: string): string {
  if (method in colorMap) {
    return colorMap[method];
  }
  return "";
}
interface AnalyticsSignupMethodsChartProps {
  dataset: AnalyticChartsQuery_signupByMethodsChart_dataset[];
}

const AnalyticsSignupMethodsChart: React.FC<AnalyticsSignupMethodsChartProps> =
  function AnalyticsSignupMethodsChart(props) {
    const { renderToString } = useContext(Context);
    const { dataset } = props;

    const noDataAvailable = useMemo(() => dataset.length === 0, [dataset]);

    const options = useMemo(() => {
      return {
        maintainAspectRatio: false,
        responsive: true,
        plugins: {
          tooltip: {
            enabled: !noDataAvailable,
          },
        },
      };
    }, [noDataAvailable]);

    const colorList = useMemo(() => {
      return dataset.map((pt) => getColorCodeByMethod(pt.label));
    }, [dataset]);

    const data = useMemo(() => {
      if (noDataAvailable) {
        // show the grey circle when there is no data
        return {
          datasets: [
            {
              data: [1],
              backgroundColor: [noDataPlaceholderColor],
              borderColor: [noDataPlaceholderColor],
              borderWidth: 1,
            },
          ],
        };
      }

      return {
        labels: dataset.map((pt) => {
          return renderToString(
            `AnalyticsSignupMethodsWidget.chart.${pt.label}.label`
          );
        }),
        datasets: [
          {
            data: dataset.map((pt) => pt.data),
            backgroundColor: colorList,
            borderColor: colorList,
            borderWidth: 1,
          },
        ],
      };
    }, [dataset, colorList, renderToString, noDataAvailable]);

    return (
      <div className={styles.chartContainer}>
        <Pie data={data} options={options} />
        {noDataAvailable && (
          <div className={styles.noDataAvailableLabel}>
            <Text variant="medium">
              <FormattedMessage
                id={`AnalyticsSignupMethodsWidget.no-data-available.label`}
              />
            </Text>
          </div>
        )}
      </div>
    );
  };

const AnalyticsSignupMethodsWidgetContent: React.FC<AnalyticsSignupMethodsWidgetProps> =
  function AnalyticsSignupMethodsWidgetContent(props) {
    const { loading, signupByMethodsChart } = props;

    const dataset = useMemo(
      () =>
        // remove null and zero data items
        (signupByMethodsChart?.dataset.filter((pt) => pt && pt.data !== 0) ??
          []) as AnalyticChartsQuery_signupByMethodsChart_dataset[],
      [signupByMethodsChart]
    );

    if (loading) {
      return (
        <div className={styles.loadingWrapper}>
          <ShowLoading />
        </div>
      );
    }

    return (
      <div>
        <AnalyticsSignupMethodsChart dataset={dataset} />
        <div className={styles.legend}>
          {dataset.map((pt) => {
            return (
              <div
                key={`legend-${pt.label}`}
                className={styles.legendItem}
                style={{ borderColor: getColorCodeByMethod(pt.label) }}
              >
                <Text variant="smallPlus">
                  <FormattedMessage
                    id={`AnalyticsSignupMethodsWidget.chart.${pt.label}.label`}
                  />
                </Text>
                <Text variant="medium" className={styles.bold}>
                  {pt.data}
                </Text>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

interface AnalyticsSignupMethodsWidgetProps {
  className?: string;
  loading: boolean;
  signupByMethodsChart: AnalyticChartsQuery_signupByMethodsChart | null;
}

const AnalyticsSignupMethodsWidget: React.FC<AnalyticsSignupMethodsWidgetProps> =
  function AnalyticsSignupMethodsWidget(props) {
    return (
      <Widget className={props.className}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsSignupMethodsWidget.title" />
        </WidgetTitle>
        <AnalyticsSignupMethodsWidgetContent {...props} />
      </Widget>
    );
  };
export default AnalyticsSignupMethodsWidget;
