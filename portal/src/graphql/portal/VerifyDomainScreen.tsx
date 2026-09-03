import React, { useCallback, useContext, useMemo, useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import { Dialog, Flex, Heading, Text } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ErrorRenderer from "../../ErrorRenderer";
import { Domain } from "./globalTypes.generated";
import { useDomainsQuery } from "./query/domainsQuery";
import { useVerifyDomainMutation } from "./mutations/verifyDomainMutation";
import styles from "./VerifyDomainScreen.module.css";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import Link from "../../Link";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";

interface VerifyDomainProps {
  domain: Domain;
  nonCustomVerifiedDomain: Domain;
}

type DNSRecordType = "CNAME" | "TXT";

interface DNSRecordListItem {
  recordType: DNSRecordType;
  host: string;
  value: string;
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

interface DNSRecordRowProps {
  item: DNSRecordListItem;
}

const DNSRecordRow: React.VFC<DNSRecordRowProps> = function DNSRecordRow(
  props
) {
  const { item } = props;

  return (
    <div className={styles.recordsTableRow}>
      <div className={styles.recordsTableCellType}>
        <Text size="2">{item.recordType}</Text>
      </div>
      <div className={styles.recordsTableCellHost}>
        <Text size="2" className={styles.cellText}>
          {item.host}
        </Text>
      </div>
      <div className={styles.recordsTableCellValue}>
        <div className={styles.valueCellInner}>
          <Text size="2" className={styles.cellText}>
            {item.value}
          </Text>
          <CopyIconButton textToCopy={item.value} />
        </div>
      </div>
    </div>
  );
};

const VerifyDomain: React.VFC<VerifyDomainProps> = function VerifyDomain(
  props: VerifyDomainProps
) {
  const { domain, nonCustomVerifiedDomain } = props;
  const navigate = useNavigate();
  const { appID } = useParams() as { appID: string };

  const {
    verifyDomain,
    loading: verifyingDomain,
    error: verifyDomainError,
  } = useVerifyDomainMutation(appID);

  const backURL = `/project/${appID}/branding/custom-domains`;

  const dnsRecordListItems = useMemo(() => {
    return makeDNSRecordListItems(domain, nonCustomVerifiedDomain);
  }, [domain, nonCustomVerifiedDomain]);

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

  const [errorDialogOpen, setErrorDialogOpen] = useState(false);

  const [prevVerifyDomainError, setPrevVerifyDomainError] =
    useState<unknown>(null);
  if (verifyDomainError !== prevVerifyDomainError) {
    setPrevVerifyDomainError(verifyDomainError);
    if (verifyDomainError != null) {
      setErrorDialogOpen(true);
    }
  }

  const verifyErrors = useMemo(() => {
    const apiErrors = parseRawError(verifyDomainError);
    const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
    return topErrors;
  }, [verifyDomainError, errorRules]);

  return (
    <ScreenLayoutScrollView>
      <ScreenContent layout="list">
        <div className={styles.widget}>
          <Link to={backURL} className={styles.backLink}>
            <ChevronLeftIcon className={styles.backLinkIcon} />
            <span>
              <FormattedMessage id="CustomDomainListScreen.title" />
            </span>
          </Link>
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="VerifyDomainScreen.title" />
          </Heading>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage
              id="VerifyDomainScreen.desc-main"
              values={{
                domain: domain.domain,
                // eslint-disable-next-line react/no-unstable-nested-components
                b: (chunks: React.ReactNode) => <b>{chunks}</b>,
              }}
            />
          </Text>
        </div>

        <div className={styles.widget}>
          <div className={styles.recordsTableWrapper}>
            <div className={styles.recordsTable}>
              <div className={styles.recordsTableHeader}>
                <div className={styles.recordsTableHeaderCellType}>
                  <FormattedMessage id="VerifyDomainScreen.list.header.record-type" />
                </div>
                <div className={styles.recordsTableHeaderCellHost}>
                  <FormattedMessage id="VerifyDomainScreen.list.header.host" />
                </div>
                <div className={styles.recordsTableHeaderCellValue}>
                  <FormattedMessage id="VerifyDomainScreen.list.header.value" />
                </div>
              </div>
              {dnsRecordListItems.map((item) => (
                <DNSRecordRow
                  key={`${item.recordType}:${item.host}`}
                  item={item}
                />
              ))}
            </div>
          </div>

          <div className={styles.verifyButton}>
            <PrimaryButton
              size="2"
              loading={verifyingDomain}
              onClick={onVerifyClick}
              text={<FormattedMessage id="verify" />}
            />
          </div>
        </div>

        <Dialog.Root open={errorDialogOpen} onOpenChange={setErrorDialogOpen}>
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="error" />
            </Dialog.Title>
            <Dialog.Description size="2">
              <ErrorRenderer errors={verifyErrors} />
            </Dialog.Description>
            <Flex gap="3" mt="4" justify="end">
              <PrimaryButton
                size="2"
                onClick={() => {
                  setErrorDialogOpen(false);
                }}
                text={<FormattedMessage id="ok" />}
              />
            </Flex>
          </Dialog.Content>
        </Dialog.Root>
      </ScreenContent>
    </ScreenLayoutScrollView>
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
    return (
      <ShowError
        error={error}
        onRetry={() => {
          refetch().finally(() => {});
        }}
      />
    );
  }

  if (domain == null) {
    // The domain does not exist in this project (e.g. after switching
    // projects); fall back to the custom domain list.
    return (
      <Navigate
        to={`/project/${appID}/branding/custom-domains`}
        replace={true}
      />
    );
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
