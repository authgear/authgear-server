import React, { useCallback, useContext, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DetailsList,
  IColumn,
  IconButton,
  IDetailsListProps,
  SelectionMode,
  Text,
} from "@fluentui/react";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import NavBreadcrumb from "../../NavBreadcrumb";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";
import { Domain } from "./globalTypes.generated";
import { useDomainsQuery } from "./query/domainsQuery";
import { useVerifyDomainMutation } from "./mutations/verifyDomainMutation";
import { useCopyFeedback } from "../../hook/useCopyFeedback";

import styles from "./VerifyDomainScreen.module.css";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import ScreenContent from "../../ScreenContent";
import Widget from "../../Widget";

interface VerifyDomainProps {
  domain: Domain;
  nonCustomVerifiedDomain: Domain;
}

// Supported DNS record types
type DNSRecordType = "CNAME" | "TXT";

interface DNSRecordListItem {
  recordType: DNSRecordType;
  host: string;
  value: string;
}

interface DNSRecordListValueCellProps {
  value: string;
}

function makeDNSRecordListColumns(
  renderToString: (messageID: string) => string
): IColumn[] {
  return [
    {
      key: "recordType",
      fieldName: "recordType",
      name: renderToString("VerifyDomainScreen.list.header.record-type"),
      minWidth: 120,
      maxWidth: 120,
      className: styles.dnsRecordListColumn,
    },
    {
      key: "host",
      fieldName: "host",
      name: renderToString("VerifyDomainScreen.list.header.host"),
      minWidth: 100,
      maxWidth: 400,
      className: styles.dnsRecordListColumn,
    },
    {
      key: "value",
      fieldName: "value",
      name: renderToString("VerifyDomainScreen.list.header.value"),
      minWidth: 400,
      className: styles.dnsRecordListColumn,
    },
  ];
}

function makeDNSRecordListItems(
  domain: Domain,
  nonCustomVerifiedDomain: Domain
): DNSRecordListItem[] {
  return [
    {
      recordType: "CNAME",
      host: domain.domain,
      value: nonCustomVerifiedDomain.domain,
    },
    {
      recordType: "TXT",
      host: domain.apexDomain,
      value: domain.verificationDNSRecord,
    },
  ];
}

const DNSRecordListValueCell: React.VFC<DNSRecordListValueCellProps> =
  function DNSRecordListValueCell(props: DNSRecordListValueCellProps) {
    const { value } = props;

    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: value,
    });

    return (
      <>
        <div className={styles.valueCell}>
          <span className={styles.valueCellText}>{value}</span>
          <IconButton {...copyButtonProps} className={styles.copyIconButton} />
        </div>
        <Feedback />
      </>
    );
  };

const VerifyDomain: React.VFC<VerifyDomainProps> = function VerifyDomain(
  props: VerifyDomainProps
) {
  const { domain, nonCustomVerifiedDomain } = props;
  const navigate = useNavigate();
  const { appID } = useParams() as { appID: string };

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems = useMemo(() => {
    return [
      {
        to: `/project/${appID}/branding/custom-domains`,
        label: <FormattedMessage id="CustomDomainListScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="VerifyDomainScreen.title" /> },
    ];
  }, [appID]);

  const {
    verifyDomain,
    loading: verifyingDomain,
    error: verifyDomainError,
  } = useVerifyDomainMutation(appID);

  const dnsRecordListColumns = useMemo(() => {
    return makeDNSRecordListColumns(renderToString);
  }, [renderToString]);

  const dnsRecordListItems = useMemo(() => {
    return makeDNSRecordListItems(domain, nonCustomVerifiedDomain);
  }, [domain, nonCustomVerifiedDomain]);

  const renderDNSRecordListColumn = useCallback<
    Required<IDetailsListProps>["onRenderItemColumn"]
  >((item?: DNSRecordListItem, _index?: number, column?: IColumn) => {
    switch (column?.key) {
      case "value":
        return <DNSRecordListValueCell value={item?.value ?? ""} />;
      default:
        return (
          <span>
            {item ? item[column?.fieldName as keyof DNSRecordListItem] : null}
          </span>
        );
    }
  }, []);

  const onVerifyClick = useCallback(() => {
    verifyDomain(domain.id)
      .then((success) => {
        if (success) {
          navigate("./../..?verify=success");
        }
      })
      .catch(() => {});
  }, [verifyDomain, domain, navigate]);

  const errorRules: ErrorParseRule[] = useMemo(() => {
    return [
      makeReasonErrorParseRule(
        "DuplicatedDomain",
        "VerifyDomainScreen.error.duplicated-error"
      ),
      makeReasonErrorParseRule(
        "DomainVerified",
        "VerifyDomainScreen.error.verified-error"
      ),
      makeReasonErrorParseRule(
        "DomainNotFound",
        "VerifyDomainScreen.error.not-found-error"
      ),
      makeReasonErrorParseRule(
        "DomainNotCustom",
        "VerifyDomainScreen.error.not-custom-error"
      ),
      makeReasonErrorParseRule(
        "DomainVerificationFailed",
        "VerifyDomainScreen.error.verification-error"
      ),
    ];
  }, []);

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <Widget className={styles.widget}>
        <Text className={styles.description} block={true}>
          <FormattedMessage
            id="VerifyDomainScreen.desc-main"
            values={{
              domain: domain.domain,
            }}
          />
        </Text>
        <DetailsList
          columns={dnsRecordListColumns}
          items={dnsRecordListItems}
          selectionMode={SelectionMode.none}
          onRenderItemColumn={renderDNSRecordListColumn}
        />
        <ButtonWithLoading
          className={styles.verifyButton}
          labelId="verify"
          loading={verifyingDomain}
          onClick={onVerifyClick}
        />
        <ErrorDialog error={verifyDomainError} rules={errorRules} />
      </Widget>
    </ScreenContent>
  );
};

const VerifyDomainScreen: React.VFC = function VerifyDomainScreen() {
  const { appID, domainID } = useParams() as {
    appID: string;
    domainID: string;
  };
  const { domains, loading, error, refetch } = useDomainsQuery(appID);
  const { renderToString } = useContext(Context);

  const domain = useMemo(() => {
    return (domains ?? []).find((domain) => domain.id === domainID);
  }, [domains, domainID]);

  const nonCustomVerifiedDomain: Domain | null = useMemo(() => {
    const nonCustomVerifiedDomainList = (domains ?? [])
      .filter((domain) => {
        return !domain.isCustom && domain.isVerified;
      })
      .map((domain) => ({
        ...domain,
        createdTimestamp: new Date(domain.createdAt).getTime(),
      }));
    const sortedList = nonCustomVerifiedDomainList.sort((domain1, domain2) => {
      return domain1.createdTimestamp - domain2.createdTimestamp;
    });
    return sortedList.length > 0 ? sortedList[0] : null;
  }, [domains]);

  const domainNotFoundError = useMemo(() => {
    const errorMessage = renderToString(
      "VerifyDomainScreen.error.domain-not-found"
    );
    return new Error(errorMessage);
  }, [renderToString]);

  const nonCustomVerifiedDomainNotFoundError = useMemo(() => {
    const errorMessage = renderToString(
      "VerifyDomainScreen.error.non-custom-verified-domain-not-found"
    );
    return new Error(errorMessage);
  }, [renderToString]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (domain == null) {
    return <ShowError error={domainNotFoundError} />;
  }

  if (nonCustomVerifiedDomain == null) {
    return <ShowError error={nonCustomVerifiedDomainNotFoundError} />;
  }

  return (
    <VerifyDomain
      domain={domain}
      nonCustomVerifiedDomain={nonCustomVerifiedDomain}
    />
  );
};

export default VerifyDomainScreen;
