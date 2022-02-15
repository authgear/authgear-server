import React, { useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Text } from "@fluentui/react";
import { Pie } from "react-chartjs-2";
import { TooltipItem } from "chart.js";
import ChartDataLabels, {
  Context as ChartDataLabelsContext,
} from "chartjs-plugin-datalabels";
import { AnalyticChartsQuery_signupConversionRate } from "./query/__generated__/AnalyticChartsQuery";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ShowLoading from "../../ShowLoading";
import styles from "./AnalyticsSignupConversionWidget.module.scss";

interface AnalyticsSignupConversionChartProps {
  signupConversionRate: AnalyticChartsQuery_signupConversionRate | null;
}

const signupViewCountColor = "#176DF3";
const notSignupViewCountColor = "#EAEAEA";
const labelColor = "#FFFFFF";
const labelBackgroundColor = "#B0B0B0";

const AnalyticsSignupConversionChart: React.FC<AnalyticsSignupConversionChartProps> =
  function AnalyticsSignupConversionChart(props) {
    const { renderToString } = useContext(Context);
    const { totalSignup = 0, totalSignupUniquePageView = 0 } =
      props.signupConversionRate ?? {};

    const signedUpPercentage = useMemo(() => {
      if (totalSignupUniquePageView <= 0) {
        return 0;
      }
      let p = (totalSignup / totalSignupUniquePageView) * 100;
      p = Math.round(p * 100) / 100;
      return p;
    }, [totalSignup, totalSignupUniquePageView]);

    const notSignedUpViewPercentage = useMemo(() => {
      if (signedUpPercentage >= 100) {
        // in normal case, signedUpPercentage should not be larger than 100
        return 0;
      }
      return 100 - signedUpPercentage;
    }, [signedUpPercentage]);

    const noDataAvailable = useMemo(
      () => signedUpPercentage === 0 && totalSignupUniquePageView === 0,
      [signedUpPercentage, totalSignupUniquePageView]
    );

    const options = {
      maintainAspectRatio: false,
      responsive: true,
      plugins: {
        datalabels: {
          display: true,
          formatter: (val: number, ctx: ChartDataLabelsContext) => {
            if (ctx.dataIndex === 0 && val > 0) {
              return ` ${val}% `;
            }
            return "";
          },
          color: labelColor,
          backgroundColor: labelBackgroundColor,
        },
        tooltip: {
          filter: function (tooltipItem: TooltipItem<"pie">) {
            // only show the tooltips for signed up percentage
            return tooltipItem.dataIndex === 0;
          },
          callbacks: {
            label: function (tooltipItem: TooltipItem<"pie">) {
              if (tooltipItem.dataIndex === 0) {
                return renderToString(
                  "AnalyticsSignupConversionWidget.chart.signup-percentage.label",
                  {
                    percentage: signedUpPercentage,
                  }
                );
              }
              return "";
            },
          },
        },
      },
    };

    const data = useMemo(() => {
      return {
        datasets: [
          {
            data: [signedUpPercentage, notSignedUpViewPercentage],
            backgroundColor: [signupViewCountColor, notSignupViewCountColor],
            borderColor: [signupViewCountColor, notSignupViewCountColor],
            borderWidth: 1,
          },
        ],
      };
    }, [signedUpPercentage, notSignedUpViewPercentage]);

    return (
      <div className={styles.chartContainer}>
        <Pie data={data} options={options} plugins={[ChartDataLabels]} />
        {noDataAvailable && (
          <div className={styles.noDataAvailableLabel}>
            <Text variant="medium">
              <FormattedMessage
                id={`AnalyticsSignupConversionWidget.no-data-available.label`}
              />
            </Text>
          </div>
        )}
      </div>
    );
  };

const AnalyticsSignupConversionWidgetContent: React.FC<AnalyticsSignupConversionWidgetProps> =
  function AnalyticsSignupConversionWidgetContent(props) {
    const { loading, signupConversionRate } = props;

    if (loading) {
      return (
        <div className={styles.loadingWrapper}>
          <ShowLoading />
        </div>
      );
    }

    return (
      <>
        <div className={styles.summaryList}>
          <div className={styles.summaryItem}>
            <Text variant="medium">
              <FormattedMessage id="AnalyticsSignupConversionWidget.unique-signup-view.label" />
            </Text>
            <Text variant="xLarge">
              {signupConversionRate?.totalSignupUniquePageView ?? "-"}
            </Text>
          </div>
          <div className={styles.summaryItem}>
            <Text variant="medium">
              <FormattedMessage id="AnalyticsSignupConversionWidget.signup.label" />
            </Text>
            <Text variant="xLarge">
              {signupConversionRate?.totalSignup ?? "-"}
            </Text>
          </div>
        </div>
        <AnalyticsSignupConversionChart
          signupConversionRate={signupConversionRate}
        />
      </>
    );
  };

interface AnalyticsSignupConversionWidgetProps {
  className?: string;
  loading: boolean;
  signupConversionRate: AnalyticChartsQuery_signupConversionRate | null;
}

const AnalyticsSignupConversionWidget: React.FC<AnalyticsSignupConversionWidgetProps> =
  function AnalyticsSignupConversionWidget(props) {
    return (
      <Widget className={props.className}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsSignupConversionWidget.title" />
        </WidgetTitle>
        <AnalyticsSignupConversionWidgetContent {...props} />
      </Widget>
    );
  };
export default AnalyticsSignupConversionWidget;
