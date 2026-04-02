import React, { useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { useQuery } from "@apollo/client";
import { Context, FormattedMessage } from "../../intl";
import CommandBarContainer from "../../CommandBarContainer";
import NavBreadcrumb from "../../NavBreadcrumb";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import { formatDatetime } from "../../util/formatDatetime";
import { FraudProtectionDecision } from "../../graphql/adminapi/globalTypes.generated";
import {
  FraudProtectionLogEntryQueryDocument,
  FraudProtectionLogEntryQueryQuery,
  FraudProtectionLogEntryQueryQueryVariables,
} from "../../graphql/adminapi/query/fraudProtectionLogEntryQuery.generated";
import styles from "./FraudProtectionLogEntryScreen.module.css";

const FraudProtectionLogEntryScreen: React.VFC =
  function FraudProtectionLogEntryScreen() {
    const { logID, appID } = useParams() as { logID: string; appID: string };
    const { renderToString, locale } = useContext(Context);

    const navBreadcrumbItems = useMemo(
      () => [
        {
          to: `/project/${appID}/attack-protection/fraud-protection#logs`,
          label: (
            <FormattedMessage id="FraudProtectionLogEntryScreen.breadcrumb.root" />
          ),
        },
        {
          to: ".",
          label: <FormattedMessage id="FraudProtectionLogEntryScreen.title" />,
        },
      ],
      [appID]
    );

    const { data, loading, error, refetch } = useQuery<
      FraudProtectionLogEntryQueryQuery,
      FraudProtectionLogEntryQueryQueryVariables
    >(FraudProtectionLogEntryQueryDocument, {
      variables: { logID },
    });

    const messageBar = useMemo(() => {
      if (error != null) {
        return <ShowError error={error} onRetry={refetch} />;
      }
      return null;
    }, [error, refetch]);

    const node =
      data?.node?.__typename === "FraudProtectionDecisionRecord"
        ? data.node
        : null;

    const createdAt =
      node != null ? formatDatetime(locale, node.createdAt) ?? "—" : "—";
    const action = (() => {
      if (node == null) {
        return "—";
      }
      switch (node.action) {
        case "send_sms":
          return renderToString(
            "FraudProtectionConfigurationScreen.logs.action.smsotp"
          );
        default:
          return node.action;
      }
    })();
    const verdict = (() => {
      if (node == null) {
        return "—";
      }
      switch (node.decision) {
        case FraudProtectionDecision.Blocked:
          return renderToString(
            "FraudProtectionConfigurationScreen.logs.verdict.blocked"
          );
        case FraudProtectionDecision.Allowed:
          return renderToString(
            "FraudProtectionConfigurationScreen.logs.verdict.allowed"
          );
        default:
          return node.decision;
      }
    })();
    const verdictClassName = (() => {
      switch (node?.decision) {
        case FraudProtectionDecision.Blocked:
          return styles.summaryBadgeBlocked;
        case FraudProtectionDecision.Allowed:
          return styles.summaryBadgeAllowed;
        default:
          return styles.summaryBadgeAllowed;
      }
    })();

    const phoneNumber = (() => {
      switch (node?.actionDetail.__typename) {
        case "FraudProtectionDecisionSendSMSActionDetail":
          return node.actionDetail.recipient;
        default:
          return "—";
      }
    })();
    const phoneCountryCode = (() => {
      switch (node?.actionDetail.__typename) {
        case "FraudProtectionDecisionSendSMSActionDetail":
          return node.actionDetail.phoneNumberCountryCode ?? "—";
        default:
          return "—";
      }
    })();
    const rawEventLog = useMemo(() => {
      if (node?.data == null) {
        return "{}";
      }
      return JSON.stringify(node.data, null, 2);
    }, [node?.data]);

    return (
      <CommandBarContainer
        isLoading={loading}
        messageBar={messageBar}
        hideCommandBar={true}
      >
        <ScreenContent layout="list">
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <section className={styles.summaryCard}>
            <div className={styles.summaryRow}>
              <div className={styles.summaryItem}>
                <span className={styles.summaryLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.timestamp" />
                </span>
                <span className={styles.summaryValue}>{createdAt}</span>
              </div>
              <div className={styles.summaryItem}>
                <span className={styles.summaryLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.action" />
                </span>
                <span className={styles.summaryValue}>{action}</span>
              </div>
              <div className={styles.summaryItem}>
                <span className={styles.summaryLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.verdict" />
                </span>
                <span className={`${styles.summaryBadge} ${verdictClassName}`}>
                  {verdict}
                </span>
              </div>
              <div className={styles.summaryItem}>
                <span className={styles.summaryLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.ip" />
                </span>
                <span className={styles.summaryValue}>
                  {node?.ipAddress || "—"}
                </span>
              </div>
              <div className={styles.summaryItem}>
                <span className={styles.summaryLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.geoLocation" />
                </span>
                <span className={styles.summaryValue}>
                  {node?.geoLocationCode || "—"}
                </span>
              </div>
            </div>
          </section>

          <div className={styles.detailsGrid}>
            <section className={styles.section}>
              <span className={styles.sectionTitle}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.deviceInfo" />
              </span>
              <div className={styles.detailRow}>
                <span className={styles.detailLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.userAgent" />
                </span>
                <span className={styles.detailValueMonospace}>
                  {node?.userAgent || "—"}
                </span>
              </div>
            </section>

            <section className={styles.section}>
              <span className={styles.sectionTitle}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.targetInfo" />
              </span>
              <div className={styles.detailRow}>
                <span className={styles.detailLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phone" />
                </span>
                <span className={styles.detailValueMonospace}>
                  {phoneNumber}
                </span>
              </div>
              <div className={styles.detailRow}>
                <span className={styles.detailLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phoneCountryCode" />
                </span>
                <span className={styles.detailValue}>{phoneCountryCode}</span>
              </div>
            </section>

            <section className={styles.section}>
              <span className={styles.sectionTitle}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.riskAssessment" />
              </span>
              <div className={styles.detailRow}>
                <span className={styles.detailLabel}>
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.reasonCodes" />
                </span>
                {node != null && node.triggeredWarnings.length > 0 ? (
                  <div className={styles.reasonCodes}>
                    {node.triggeredWarnings.map((code) => (
                      <span key={code} className={styles.reasonCodeTag}>
                        {code}
                      </span>
                    ))}
                  </div>
                ) : (
                  <span className={styles.detailValue}>
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.none" />
                  </span>
                )}
              </div>
            </section>

            <section className={`${styles.section} ${styles.sectionFull}`}>
              <div className={styles.rawLogHeader}>
                <FormattedMessage id="FraudProtectionLogEntryScreen.rawEventLog" />
              </div>
              <div className={styles.rawLogContent}>
                <pre className={styles.rawLogPre}>{rawEventLog}</pre>
              </div>
            </section>
          </div>
        </ScreenContent>
      </CommandBarContainer>
    );
  };

export default FraudProtectionLogEntryScreen;
