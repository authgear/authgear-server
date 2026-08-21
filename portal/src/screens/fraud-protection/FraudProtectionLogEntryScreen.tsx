import React, { useContext, useMemo } from "react";
import { Navigate, useParams } from "react-router-dom";
import { useQuery } from "@apollo/client";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { formatDatetime } from "../../util/formatDatetime";
import { FraudProtectionDecision } from "../../graphql/adminapi/globalTypes.generated";
import {
  FraudProtectionLogEntryQueryDocument,
  FraudProtectionLogEntryQueryQuery,
  FraudProtectionLogEntryQueryQueryVariables,
} from "../../graphql/adminapi/query/fraudProtectionLogEntryQuery.generated";
import styles from "./FraudProtectionLogEntryScreen.module.css";

function getResultMessageID(
  decision: FraudProtectionDecision,
  triggeredWarnings: readonly unknown[]
): string {
  if (decision === FraudProtectionDecision.Blocked) {
    return "FraudProtectionConfigurationScreen.logs.result.blocked";
  }
  if (triggeredWarnings.length > 0) {
    return "FraudProtectionConfigurationScreen.logs.result.flagged";
  }
  return "FraudProtectionConfigurationScreen.logs.result.allowed";
}

const Field: React.VFC<{
  label: React.ReactNode;
  children: React.ReactNode;
  monospace?: boolean;
}> = function Field({ label, children, monospace }) {
  return (
    <div className={styles.field}>
      <Text as="p" size="1" className={styles.fieldLabel}>
        {label}
      </Text>
      <Text
        as="p"
        size="2"
        className={cn(styles.fieldValue, monospace && styles.fieldValueMono)}
      >
        {children}
      </Text>
    </div>
  );
};

const FraudProtectionLogEntryScreen: React.VFC =
  function FraudProtectionLogEntryScreen() {
    const { logID, appID } = useParams() as { logID: string; appID: string };
    const { renderToString, locale } = useContext(Context);

    const { data, loading, error, refetch } = useQuery<
      FraudProtectionLogEntryQueryQuery,
      FraudProtectionLogEntryQueryQueryVariables
    >(FraudProtectionLogEntryQueryDocument, {
      variables: { logID },
    });

    const node =
      data?.node?.__typename === "FraudProtectionDecisionRecord"
        ? data.node
        : null;

    const createdAt =
      node != null ? formatDatetime(locale, node.createdAt) ?? "—" : "—";

    const action =
      node != null
        ? renderToString(
            "FraudProtectionConfigurationScreen.logs.action.smsotp"
          )
        : "—";

    const triggeredWarnings: readonly string[] = node?.triggeredWarnings ?? [];
    const decision: FraudProtectionDecision | null = node?.decision ?? null;

    const verdict =
      decision == null
        ? "—"
        : renderToString(getResultMessageID(decision, triggeredWarnings));

    const verdictClassName = (() => {
      if (decision === FraudProtectionDecision.Blocked) {
        return styles.badgeBlocked;
      }
      if (triggeredWarnings.length > 0) {
        return styles.badgeFlagged;
      }
      return styles.badgeAllowed;
    })();

    const ipAddress = node?.ipAddress || "—";
    const geoLocationCode = node?.geoLocationCode || "—";
    const userAgent = node?.userAgent || "—";

    const phoneNumber =
      node?.actionDetail.__typename ===
      "FraudProtectionDecisionSendSMSActionDetail"
        ? node.actionDetail.recipient
        : "—";
    const phoneCountryCode =
      node?.actionDetail.__typename ===
      "FraudProtectionDecisionSendSMSActionDetail"
        ? node.actionDetail.phoneNumberCountryCode ?? "—"
        : "—";

    const rawEventLog = useMemo(() => {
      if (node?.data == null) return "{}";
      return JSON.stringify(node.data, null, 2);
    }, [node]);

    if (!loading && error == null && node == null) {
      return (
        <Navigate
          to={`/project/${appID}/attack-protection/fraud-protection#logs`}
          replace={true}
        />
      );
    }

    return (
      <APIResourceScreenLayout
        layout="list"
        breadcrumbItems={[
          {
            to: "~/attack-protection/fraud-protection#logs",
            label: (
              <FormattedMessage id="FraudProtectionLogEntryScreen.breadcrumb.root" />
            ),
          },
          {
            to: "",
            label: (
              <FormattedMessage id="FraudProtectionLogEntryScreen.title" />
            ),
          },
        ]}
      >
        {error != null ? (
          <ShowError
            error={error}
            onRetry={() => {
              void refetch();
            }}
          />
        ) : loading ? (
          <ShowLoading />
        ) : (
          <div className={styles.content}>
            <section className={styles.card}>
              <div className={styles.summaryGrid}>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.timestamp" />
                  }
                >
                  {createdAt}
                </Field>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.action" />
                  }
                >
                  {action}
                </Field>
                <div className={styles.field}>
                  <Text as="p" size="1" className={styles.fieldLabel}>
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.result" />
                  </Text>
                  <span className={cn(styles.badge, verdictClassName)}>
                    {verdict}
                  </span>
                </div>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.column.ip" />
                  }
                  monospace={true}
                >
                  {ipAddress}
                </Field>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.geoLocation" />
                  }
                >
                  {geoLocationCode}
                </Field>
              </div>
            </section>

            <div className={styles.detailsGrid}>
              <section className={styles.card}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.cardTitle}
                >
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.deviceInfo" />
                </Text>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.userAgent" />
                  }
                  monospace={true}
                >
                  {userAgent}
                </Field>
              </section>

              <section className={styles.card}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.cardTitle}
                >
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.targetInfo" />
                </Text>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phone" />
                  }
                  monospace={true}
                >
                  {phoneNumber}
                </Field>
                <Field
                  label={
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phoneCountryCode" />
                  }
                >
                  {phoneCountryCode}
                </Field>
              </section>

              <section className={styles.card}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.cardTitle}
                >
                  <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.riskAssessment" />
                </Text>
                <div className={styles.field}>
                  <Text as="p" size="1" className={styles.fieldLabel}>
                    <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.reasonCodes" />
                  </Text>
                  {triggeredWarnings.length > 0 ? (
                    <div className={styles.reasonCodes}>
                      {triggeredWarnings.map((code) => (
                        <span key={code} className={styles.reasonCodeTag}>
                          {code}
                        </span>
                      ))}
                    </div>
                  ) : (
                    <Text as="p" size="2" className={styles.fieldValue}>
                      <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.none" />
                    </Text>
                  )}
                </div>
              </section>

              <section className={cn(styles.card, styles.cardFull)}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.cardTitle}
                >
                  <FormattedMessage id="FraudProtectionLogEntryScreen.rawEventLog" />
                </Text>
                <pre className={styles.rawLogPre}>{rawEventLog}</pre>
              </section>
            </div>
          </div>
        )}
      </APIResourceScreenLayout>
    );
  };

export default FraudProtectionLogEntryScreen;
