import React, { useContext, useMemo } from "react";
import { Bar } from "react-chartjs-2";
import { ChartOptions } from "chart.js";
import { DateTime } from "luxon";
import { Context } from "../../intl";
import { FraudProtectionOverviewQueryQuery } from "../../graphql/adminapi/query/fraudProtectionOverviewQuery.generated";
import styles from "./OverviewRequestsChart.module.css";

type TimeBucket =
  FraudProtectionOverviewQueryQuery["fraudProtectionOverview"]["sendSMS"]["timeBuckets"][number];

interface SlottedBucket {
  label: string;
  total: number;
  blocked: number;
  flagged: number;
}

interface OverviewRequestsChartProps {
  timeBuckets: TimeBucket[];
  timeRange: "24h" | "7d";
  rangeFrom: string;
  rangeTo: string;
}

// Build hourly slots for 24h view.
function buildHourlySlots(
  timeBuckets: TimeBucket[],
  rangeFrom: string,
  rangeTo: string
): SlottedBucket[] {
  const byHour = new Map<string, TimeBucket>();
  for (const b of timeBuckets) {
    const key = DateTime.fromISO(b.hour).toUTC().startOf("hour").toISO();
    if (key != null) byHour.set(key, b);
  }

  const from = DateTime.fromISO(rangeFrom).toUTC().startOf("hour");
  const to = DateTime.fromISO(rangeTo).toUTC().startOf("hour");

  const slots: SlottedBucket[] = [];
  let cursor = from;
  while (cursor <= to) {
    const key = cursor.toISO();
    const bucket = key != null ? byHour.get(key) : undefined;
    slots.push({
      label: cursor.toLocal().toFormat("h a"),
      total: bucket?.total ?? 0,
      blocked: bucket?.blocked ?? 0,
      flagged: bucket?.flagged ?? 0,
    });
    cursor = cursor.plus({ hours: 1 });
  }
  return slots;
}

// Build daily slots for 7d view by aggregating hourly buckets into days.
function buildDailySlots(
  timeBuckets: TimeBucket[],
  rangeFrom: string,
  rangeTo: string
): SlottedBucket[] {
  // Aggregate by local day key "yyyy-MM-dd"
  const byDay = new Map<string, SlottedBucket>();
  for (const b of timeBuckets) {
    const dayKey = DateTime.fromISO(b.hour).toLocal().toFormat("yyyy-MM-dd");
    const existing = byDay.get(dayKey);
    if (existing != null) {
      existing.total += b.total;
      existing.blocked += b.blocked;
      existing.flagged += b.flagged;
    } else {
      const label = DateTime.fromISO(b.hour).toLocal().toFormat("LLL d");
      byDay.set(dayKey, { label, total: b.total, blocked: b.blocked, flagged: b.flagged });
    }
  }

  // Generate every day in range to fill gaps
  const from = DateTime.fromISO(rangeFrom).toLocal().startOf("day");
  const to = DateTime.fromISO(rangeTo).toLocal().startOf("day");

  const slots: SlottedBucket[] = [];
  let cursor = from;
  while (cursor <= to) {
    const dayKey = cursor.toFormat("yyyy-MM-dd");
    const existing = byDay.get(dayKey);
    slots.push(
      existing ?? { label: cursor.toFormat("LLL d"), total: 0, blocked: 0, flagged: 0 }
    );
    cursor = cursor.plus({ days: 1 });
  }
  return slots;
}

function buildSlots(
  timeBuckets: TimeBucket[],
  rangeFrom: string,
  rangeTo: string,
  timeRange: "24h" | "7d"
): SlottedBucket[] {
  if (timeRange === "7d") {
    return buildDailySlots(timeBuckets, rangeFrom, rangeTo);
  }
  return buildHourlySlots(timeBuckets, rangeFrom, rangeTo);
}

const OverviewRequestsChart: React.VFC<OverviewRequestsChartProps> =
  function OverviewRequestsChart({ timeBuckets, timeRange, rangeFrom, rangeTo }) {
    const { renderToString } = useContext(Context);

    const chartLabels = useMemo(
      () => ({
        title: renderToString(
          "FraudProtectionConfigurationScreen.overview.chart.title"
        ),
        blocked: renderToString(
          "FraudProtectionConfigurationScreen.overview.chart.blocked"
        ),
        flagged: renderToString(
          "FraudProtectionConfigurationScreen.overview.chart.flagged"
        ),
        totalRequests: renderToString(
          "FraudProtectionConfigurationScreen.overview.chart.totalRequests"
        ),
      }),
      [renderToString]
    );

    const slots = useMemo(
      () => buildSlots(timeBuckets, rangeFrom, rangeTo, timeRange),
      [timeBuckets, rangeFrom, rangeTo, timeRange]
    );

    const data = useMemo(
      () => ({
        labels: slots.map((s) => s.label),
        datasets: [
          {
            label: chartLabels.blocked,
            data: slots.map((s) => s.blocked),
            backgroundColor: "#fca5a5",
            borderWidth: 0,
            stack: "stack",
          },
          {
            label: chartLabels.flagged,
            data: slots.map((s) => s.flagged),
            backgroundColor: "#fde68a",
            borderWidth: 0,
            stack: "stack",
          },
          {
            label: chartLabels.totalRequests,
            data: slots.map((s) => Math.max(0, s.total - s.blocked - s.flagged)),
            backgroundColor: "#e5e5e5",
            borderWidth: 0,
            stack: "stack",
          },
        ],
      }),
      [chartLabels.blocked, chartLabels.flagged, chartLabels.totalRequests, slots]
    );

    const options = useMemo<ChartOptions<"bar">>(
      () => ({
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: "bottom" as const,
            labels: {
              usePointStyle: true,
              pointStyle: "rect" as const,
              boxWidth: 12,
              boxHeight: 12,
              padding: 24,
              color: "#605e5c",
              font: { size: 12, family: "'Segoe UI', system-ui, -apple-system, sans-serif" },
              generateLabels: (chart) => {
                const datasets = chart.data.datasets;
                return datasets.map((ds, i) => ({
                  text: ds.label ?? "",
                  fillStyle: ds.backgroundColor as string,
                  strokeStyle: "transparent",
                  lineWidth: 0,
                  datasetIndex: i,
                  hidden: false,
                  fontColor: "#605e5c",
                }));
              },
            },
          },
          tooltip: {
            mode: "index" as const,
            backgroundColor: "#ffffff",
            borderColor: "#edebe9",
            borderWidth: 1,
            titleColor: "#323130",
            bodyColor: "#605e5c",
            titleFont: { size: 12, weight: 600, family: "'Segoe UI', system-ui, -apple-system, sans-serif" },
            bodyFont: { size: 12, family: "'Segoe UI', system-ui, -apple-system, sans-serif" },
            padding: 10,
            cornerRadius: 2,
            boxWidth: 10,
            boxHeight: 10,
            usePointStyle: true,
            callbacks: {
              label: (ctx) => {
                if (ctx.datasetIndex === 2) {
                  const blocked = ctx.chart.data.datasets[0].data[ctx.dataIndex] as number;
                  const flagged = ctx.chart.data.datasets[1].data[ctx.dataIndex] as number;
                  const allowed = ctx.raw as number;
                  return `  ${renderToString(
                    "FraudProtectionConfigurationScreen.overview.chart.tooltip.totalRequests",
                    { count: blocked + flagged + allowed }
                  )}`;
                }
                const label = ctx.dataset.label ?? "";
                return `  ${renderToString(
                  "FraudProtectionConfigurationScreen.overview.chart.tooltip.item",
                  { label, count: ctx.raw }
                )}`;
              },
            },
          },
        },
        scales: {
          x: {
            stacked: true,
            grid: { display: false },
            ticks: {
              maxRotation: 0,
              autoSkip: false,
              font: { size: 11, family: "'Segoe UI', system-ui, -apple-system, sans-serif" },
              color: "#8a8886",
            },
          },
          y: {
            stacked: true,
            min: 0,
            grid: { color: "#f3f2f1" },
            ticks: {
              precision: 0,
              font: { size: 11, family: "'Segoe UI', system-ui, -apple-system, sans-serif" },
              color: "#8a8886",
            },
            border: { display: false },
          },
        },
      }),
      [renderToString, timeRange]
    );

    return (
      <div className={styles.container}>
        <div className={styles.title}>{chartLabels.title}</div>
        <div className={styles.chartWrap}>
          <Bar options={options} data={data} />
        </div>
      </div>
    );
  };

export default OverviewRequestsChart;
