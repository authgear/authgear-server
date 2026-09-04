import React, { useContext, useMemo } from "react";
import { Context, FormattedMessage } from "../../intl";
import { Spinner, Text } from "@radix-ui/themes";
import { Pie } from "react-chartjs-2";
import { TooltipItem } from "chart.js";
import ChartDataLabels, {
  Context as ChartDataLabelsContext,
} from "chartjs-plugin-datalabels";
import { AnalyticChartsQueryQuery } from "./query/analyticChartsQuery.generated";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./AnalyticsSignupConversionWidget.module.css";

interface AnalyticsSignupConversionChartProps {
  signupConversionRate: AnalyticChartsQueryQuery["signupConversionRate"] | null;
}

const SignedUpPercentageDataIndex = 0; // index of SignedUpPercentage data in the dataset
const SignedUpPercentageColor = "#176DF3";
const NotSignedUpPercentageColor = "#EAEAEA";
const LabelColor = "#FFFFFF";
const LabelBorderColor = "#FFFFFF";
const LabelBackgroundColor = "#176DF3";

const AnalyticsSignupConversionChart: React.VFC<AnalyticsSignupConversionChartProps> =
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

    function shouldShowLabelForData(ctx: ChartDataLabelsContext): boolean {
      const val = ctx.dataset.data[ctx.dataIndex] as number;
      if (ctx.dataIndex === SignedUpPercentageDataIndex && val > 0) {
        return true;
      }
      return false;
    }

    const options = {
      maintainAspectRatio: false,
      responsive: true,
      plugins: {
        datalabels: {
          display: (ctx: ChartDataLabelsContext) => shouldShowLabelForData(ctx),
          formatter: (val: number, ctx: ChartDataLabelsContext) => {
            if (shouldShowLabelForData(ctx)) {
              return ` ${val}% `;
            }
            return "";
          },
          color: LabelColor,
          backgroundColor: LabelBackgroundColor,
          borderColor: LabelBorderColor,
          borderWidth: 2,
          borderRadius: 2,
        },
        tooltip: {
          filter: function (tooltipItem: TooltipItem<"pie">) {
            // only show the tooltips for signed up percentage
            return tooltipItem.dataIndex === SignedUpPercentageDataIndex;
          },
          callbacks: {
            label: function (tooltipItem: TooltipItem<"pie">) {
              if (tooltipItem.dataIndex === SignedUpPercentageDataIndex) {
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
            backgroundColor: [
              SignedUpPercentageColor,
              NotSignedUpPercentageColor,
            ],
            borderColor: [SignedUpPercentageColor, NotSignedUpPercentageColor],
            borderWidth: 1,
          },
        ],
      };
    }, [signedUpPercentage, notSignedUpViewPercentage]);

    return (
      <div className={styles.chartContainer}>
        <Pie data={data} options={options} plugins={[ChartDataLabels]} />
        {noDataAvailable ? (
          <div className={styles.noDataAvailableLabel}>
            <Text as="p" size="2" className={styles.emptyLabel}>
              <FormattedMessage
                id={`AnalyticsSignupConversionWidget.no-data-available.label`}
              />
            </Text>
          </div>
        ) : null}
      </div>
    );
  };

const AnalyticsSignupConversionWidgetContent: React.VFC<AnalyticsSignupConversionWidgetProps> =
  function AnalyticsSignupConversionWidgetContent(props) {
    const { loading, signupConversionRate } = props;

    if (loading) {
      return (
        <div className={styles.loadingWrapper}>
          <Spinner size="3" />
        </div>
      );
    }

    return (
      <>
        <div className={styles.summaryList}>
          <div className={styles.summaryItem}>
            <Text as="p" size="2" className={styles.metricLabel}>
              <FormattedMessage id="AnalyticsSignupConversionWidget.unique-signup-view.label" />
            </Text>
            <Text
              as="p"
              size="6"
              weight="medium"
              className={styles.metricValue}
            >
              {signupConversionRate?.totalSignupUniquePageView ?? "-"}
            </Text>
          </div>
          <div className={styles.summaryItem}>
            <Text as="p" size="2" className={styles.metricLabel}>
              <FormattedMessage id="AnalyticsSignupConversionWidget.signup.label" />
            </Text>
            <Text
              as="p"
              size="6"
              weight="medium"
              className={styles.metricValue}
            >
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
  signupConversionRate: AnalyticChartsQueryQuery["signupConversionRate"] | null;
}

const AnalyticsSignupConversionWidget: React.VFC<AnalyticsSignupConversionWidgetProps> =
  function AnalyticsSignupConversionWidget(props) {
    return (
      <SettingsSectionCard
        className={props.className}
        layout="stacked"
        title={<FormattedMessage id="AnalyticsSignupConversionWidget.title" />}
        contentClassName={styles.content}
      >
        <AnalyticsSignupConversionWidgetContent {...props} />
      </SettingsSectionCard>
    );
  };
export default AnalyticsSignupConversionWidget;
