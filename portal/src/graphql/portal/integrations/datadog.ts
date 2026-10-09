import { produce } from "immer";
import {
  PortalAPIAppConfig,
  TelemetryAuditLogStreamConfig,
} from "../../../types";
import { clearEmptyObject } from "../../../util/misc";

export const DATADOG_DEFAULT_SITE = "datadoghq.com";
export const DATADOG_SITE_OTHER = "__other__";
const DATADOG_STREAM_NAME = "datadog";

export interface DatadogSiteOption {
  site: string;
  label: string;
}

export const DATADOG_SITE_OPTIONS: DatadogSiteOption[] = [
  { site: DATADOG_DEFAULT_SITE, label: "US1" },
  { site: "us3.datadoghq.com", label: "US3" },
  { site: "us5.datadoghq.com", label: "US5" },
  { site: "datadoghq.eu", label: "EU1" },
  { site: "ap1.datadoghq.com", label: "AP1" },
  { site: "ap2.datadoghq.com", label: "AP2" },
  { site: "uk1.datadoghq.com", label: "UK1" },
  { site: "ddog-gov.com", label: "US1-FED" },
];

function streamsOf(
  config: PortalAPIAppConfig
): TelemetryAuditLogStreamConfig[] {
  return config.telemetry?.audit_logs?.streams ?? [];
}

export function findManagedDatadogStream(
  config: PortalAPIAppConfig
): TelemetryAuditLogStreamConfig | null {
  return streamsOf(config).find((s) => s.type === "datadog") ?? null;
}

export function allocateDatadogStreamName(config: PortalAPIAppConfig): string {
  const taken = new Set(streamsOf(config).map((s) => s.name));
  if (!taken.has(DATADOG_STREAM_NAME)) {
    return DATADOG_STREAM_NAME;
  }
  for (let n = 2; ; n++) {
    const name = `${DATADOG_STREAM_NAME}-${n}`;
    if (!taken.has(name)) {
      return name;
    }
  }
}

export function siteToFormValue(site: string | undefined): {
  option: string;
  other: string;
} {
  const value = site ?? DATADOG_DEFAULT_SITE;
  return DATADOG_SITE_OPTIONS.some((o) => o.site === value)
    ? { option: value, other: "" }
    : { option: DATADOG_SITE_OTHER, other: value };
}

export function formValueToSite(option: string, other: string): string {
  return option === DATADOG_SITE_OTHER ? other.trim() : option;
}

const DATADOG_HOST_PATTERN = /^[a-z0-9-]+(\.[a-z0-9-]+)+$/;

// Datadog serves a bare two-label site at app.<site> and a regional site at
// <site> itself. Orgs live on one site, so the link must follow the site.
export function datadogAppOrigin(site: string): string {
  let host = site.trim().toLowerCase();
  if (!DATADOG_HOST_PATTERN.test(host)) {
    host = DATADOG_DEFAULT_SITE;
  }
  return host.split(".").length === 2
    ? `https://app.${host}`
    : `https://${host}`;
}

export function applyDatadogConnect(
  config: PortalAPIAppConfig,
  site: string
): PortalAPIAppConfig {
  const name = allocateDatadogStreamName(config);
  return produce(config, (draft) => {
    draft.telemetry ??= {};
    draft.telemetry.audit_logs ??= {};
    draft.telemetry.audit_logs.streams ??= [];
    draft.telemetry.audit_logs.streams.push({
      name,
      type: "datadog",
      transport: "http",
      datadog: { site },
    });
  });
}

export function applyDatadogEdit(
  config: PortalAPIAppConfig,
  site: string
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    const stream = draft.telemetry?.audit_logs?.streams?.find(
      (s) => s.type === "datadog"
    );
    // site and http.endpoint are mutually exclusive on the server.
    if (stream == null || stream.http?.endpoint != null) {
      return;
    }
    stream.datadog ??= {};
    stream.datadog.site = site;
  });
}

export function applyDatadogDelete(
  config: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    const auditLogs = draft.telemetry?.audit_logs;
    const streams = auditLogs?.streams;
    if (auditLogs == null || streams == null) {
      return;
    }
    const idx = streams.findIndex((s) => s.type === "datadog");
    if (idx >= 0) {
      streams.splice(idx, 1);
    }
    if (streams.length === 0) {
      delete auditLogs.streams;
    }
    // Drops audit_logs and telemetry once they are empty.
    clearEmptyObject(draft);
  });
}

export function remainingStreamNames(config: PortalAPIAppConfig): string[] {
  return streamsOf(config).map((s) => s.name);
}
