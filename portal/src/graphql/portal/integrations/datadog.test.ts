import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../types";
import {
  allocateDatadogStreamName,
  applyDatadogConnect,
  applyDatadogDelete,
  applyDatadogEdit,
  findManagedDatadogStream,
  formValueToSite,
  remainingStreamNames,
  siteToFormValue,
  DATADOG_SITE_OTHER,
} from "./datadog";

const syslog = {
  name: "datadog",
  type: "syslog" as const,
  transport: "tcp" as const,
  tcp: { address: "collector:5140" },
  syslog: { format: "rfc5424", framing: "newline" },
};

describe("datadog helpers", () => {
  it("connect appends a stream and keeps others", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: { audit_logs: { streams: [syslog] } },
    };
    const out = applyDatadogConnect(cfg, "datadoghq.eu");
    expect(out.telemetry?.audit_logs?.streams).toEqual([
      syslog,
      {
        name: "datadog-2",
        type: "datadog",
        transport: "http",
        datadog: { site: "datadoghq.eu" },
      },
    ]);
    expect(cfg.telemetry?.audit_logs?.streams).toHaveLength(1); // input not mutated
  });

  it("allocates datadog when free", () => {
    expect(allocateDatadogStreamName({ id: "app" })).toBe("datadog");
  });

  it("edit preserves service/source/tags", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: {
        audit_logs: {
          streams: [
            {
              name: "dd",
              type: "datadog",
              transport: "http",
              datadog: {
                site: "datadoghq.com",
                service: "s",
                tags: { env: "prod" },
              },
            },
          ],
        },
      },
    };
    const out = applyDatadogEdit(cfg, "us5.datadoghq.com");
    expect(findManagedDatadogStream(out)?.datadog).toEqual({
      site: "us5.datadoghq.com",
      service: "s",
      tags: { env: "prod" },
    });
  });

  it("edit does not set site when http.endpoint is used", () => {
    const stream = {
      name: "dd",
      type: "datadog" as const,
      transport: "http" as const,
      http: { endpoint: "https://opw.internal/api/v2/logs" },
    };
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: { audit_logs: { streams: [stream] } },
    };
    expect(
      findManagedDatadogStream(applyDatadogEdit(cfg, "datadoghq.eu"))
    ).toEqual(stream);
  });

  it("delete of the only stream removes telemetry entirely", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: {
        audit_logs: {
          streams: [{ name: "datadog", type: "datadog", transport: "http" }],
        },
      },
    };
    const out = applyDatadogDelete(cfg);
    expect(out.telemetry).toBeUndefined();
    expect(remainingStreamNames(out)).toEqual([]);
  });

  it("delete keeps other streams", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: {
        audit_logs: {
          streams: [
            syslog,
            { name: "datadog-2", type: "datadog", transport: "http" },
          ],
        },
      },
    };
    expect(remainingStreamNames(applyDatadogDelete(cfg))).toEqual(["datadog"]);
  });

  it("unknown site round-trips through Other", () => {
    expect(siteToFormValue("us2.ddog-gov.com")).toEqual({
      option: DATADOG_SITE_OTHER,
      other: "us2.ddog-gov.com",
    });
    expect(siteToFormValue(undefined)).toEqual({
      option: "datadoghq.com",
      other: "",
    });
    expect(formValueToSite(DATADOG_SITE_OTHER, "  us2.ddog-gov.com ")).toBe(
      "us2.ddog-gov.com"
    );
    expect(formValueToSite("datadoghq.eu", "ignored")).toBe("datadoghq.eu");
  });
});
