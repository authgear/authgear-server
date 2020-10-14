import React, { useCallback, useContext, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DefaultButton,
  DetailsList,
  IColumn,
  IconButton,
  IDetailsListProps,
  SelectionMode,
  Stack,
  Text,
} from "@fluentui/react";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import NavBreadcrumb from "../../NavBreadcrumb";
import ButtonWithLoading from "../../ButtonWithLoading";
import { Domain, useDomainsQuery } from "./query/domainsQuery";
import { copyToClipboard } from "../../util/clipboard";

import styles from "./VerifyDomainScreen.module.scss";

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
      minWidth: 200,
      maxWidth: 400,
      className: styles.dnsRecordListColumn,
    },
    {
      key: "value",
      fieldName: "value",
      name: renderToString("VerifyDomainScreen.list.header.value"),
      minWidth: 300,
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

const DNSRecordListValueCell: React.FC<DNSRecordListValueCellProps> = function DNSRecordListValueCell(
  props: DNSRecordListValueCellProps
) {
  const { value } = props;

  const onCopyClick = useCallback(() => {
    copyToClipboard(value);
  }, [value]);

  return (
    <div className={styles.valueCell}>
      <span className={styles.valueCellText}>{value}</span>
      <IconButton
        className={styles.copyIconButton}
        onClick={onCopyClick}
        iconProps={{ iconName: "Copy" }}
      />
    </div>
  );
};

const VerifyDomain: React.FC<VerifyDomainProps> = function VerifyDomain(
  props: VerifyDomainProps
) {
  const { domain, nonCustomVerifiedDomain } = props;
  const navigate = useNavigate();

  const { renderToString } = useContext(Context);

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
            {item && item[column?.fieldName as keyof DNSRecordListItem]}
          </span>
        );
    }
  }, []);

  const onVerifyClick = useCallback(() => {
    // TODO: to be implemented
  }, []);

  const onCancelClick = useCallback(() => {
    navigate("../..");
  }, [navigate]);

  return (
    <section className={styles.content}>
      <Text className={styles.desc}>
        <span>
          <FormattedMessage id="VerifyDomainScreen.desc-main" />
        </span>
        <span className={styles.descDomain}>{domain.domain}</span>
      </Text>
      <DetailsList
        columns={dnsRecordListColumns}
        items={dnsRecordListItems}
        selectionMode={SelectionMode.none}
        onRenderItemColumn={renderDNSRecordListColumn}
      />
      <Stack
        className={styles.controlButtons}
        horizontal={true}
        tokens={{ childrenGap: 10 }}
      >
        <ButtonWithLoading
          labelId="verify"
          loading={false}
          onClick={onVerifyClick}
        />
        <DefaultButton onClick={onCancelClick}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </Stack>
    </section>
  );
};

const VerifyDomainScreen: React.FC = function VerifyDomainScreen() {
  const { appID, domainID } = useParams();
  const { domains, loading, error, refetch } = useDomainsQuery(appID);
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems = useMemo(() => {
    return [
      {
        to: "../..",
        label: <FormattedMessage id="DNSConfigurationScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="VerifyDomainScreen.title" /> },
    ];
  }, []);

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
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <VerifyDomain
        domain={domain}
        nonCustomVerifiedDomain={nonCustomVerifiedDomain}
      />
    </main>
  );
};

export default VerifyDomainScreen;
