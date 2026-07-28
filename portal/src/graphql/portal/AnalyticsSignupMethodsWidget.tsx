import React, { useContext, useMemo } from "react";
import { Context, FormattedMessage } from "../../intl";
import { Pie } from "react-chartjs-2";
import { Spinner, Text } from "@radix-ui/themes";

import { DataPoint } from "./globalTypes.generated";
import { AnalyticChartsQueryQuery } from "./query/analyticChartsQuery.generated";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./AnalyticsSignupMethodsWidget.module.css";

const NoDataPlaceholderColor = "#EAEAEA";
const ColorMap: Record<string, string> = {
  email: "#FF7629",
  phone: "#92674F",
  username: "#FFB900",
  google: "#EA4335",
  facebook: "#4267B2",
  github: "#171515",
  linkedin: "#0077B5",
  azureadv2: "#007FFF",
  azureadb2c: "#2DE7FD",
  adfs: "#49A8EC",
  apple: "#A2AAAD",
  wechat: "#45B049",
  anonymous: "#957AFF",
};
const UnknownMethodColor = "#EAEAEA";

function getColorCodeByMethod(method: string): string {
  if (method in ColorMap) {
    return ColorMap[method];
  }
  return UnknownMethodColor;
}
interface AnalyticsSignupMethodsChartProps {
  dataset: DataPoint[];
}

const AnalyticsSignupMethodsChart: React.VFC<AnalyticsSignupMethodsChartProps> =
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
              backgroundColor: [NoDataPlaceholderColor],
              borderColor: [NoDataPlaceholderColor],
              borderWidth: 1,
            },
          ],
        };
      }

      return {
        labels: dataset.map((pt) => {
          const label = renderToString(
            `AnalyticsSignupMethodsWidget.chart.${pt.label}.label`
          );
          return label ? label : pt.label;
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
        {noDataAvailable ? (
          <div className={styles.noDataAvailableLabel}>
            <Text as="p" size="2" className={styles.emptyLabel}>
              <FormattedMessage
                id={`AnalyticsSignupMethodsWidget.no-data-available.label`}
              />
            </Text>
          </div>
        ) : null}
      </div>
    );
  };

const AnalyticsSignupMethodsWidgetContent: React.VFC<AnalyticsSignupMethodsWidgetProps> =
  function AnalyticsSignupMethodsWidgetContent(props) {
    const { loading, signupByMethodsChart } = props;
    const { renderToString } = useContext(Context);

    const dataset = useMemo(
      () =>
        // remove null and zero data items
        (signupByMethodsChart?.dataset.filter((pt) => pt && pt.data !== 0) ??
          []) as DataPoint[],
      [signupByMethodsChart]
    );

    const methodLabels = useMemo(
      () =>
        dataset.map((pt) => {
          const label = renderToString(
            `AnalyticsSignupMethodsWidget.chart.${pt.label}.label`
          );
          return label ? label : pt.label;
        }),
      [dataset, renderToString]
    );

    if (loading) {
      return (
        <div className={styles.loadingWrapper}>
          <Spinner size="3" />
        </div>
      );
    }

    return (
      <div>
        <AnalyticsSignupMethodsChart dataset={dataset} />
        <div className={styles.legend}>
          {dataset.map((pt, i) => {
            return (
              <div
                key={`legend-${pt.label}`}
                className={styles.legendItem}
                style={{ borderColor: getColorCodeByMethod(pt.label) }}
              >
                <Text as="p" size="1" className={styles.legendLabel}>
                  {methodLabels[i]}
                </Text>
                <Text as="p" size="2" weight="medium" className={styles.bold}>
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
  signupByMethodsChart: AnalyticChartsQueryQuery["signupByMethodsChart"] | null;
}

const AnalyticsSignupMethodsWidget: React.VFC<AnalyticsSignupMethodsWidgetProps> =
  function AnalyticsSignupMethodsWidget(props) {
    return (
      <SettingsSectionCard
        className={props.className}
        layout="stacked"
        title={<FormattedMessage id="AnalyticsSignupMethodsWidget.title" />}
        contentClassName={styles.content}
      >
        <AnalyticsSignupMethodsWidgetContent {...props} />
      </SettingsSectionCard>
    );
  };
export default AnalyticsSignupMethodsWidget;
