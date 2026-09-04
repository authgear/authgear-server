import React, { useContext } from "react";
import cn from "classnames";
import { Heading, Text } from "@radix-ui/themes";
import ScreenContent from "../../ScreenContent";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { FormattedMessage, Context } from "../../intl";
import { TextField } from "../../components/v2/TextField/TextField";
import { formatDateOnly } from "../../util/formatDateOnly";
import styles from "./LicenseScreen.module.css";

function LicenseScreen(): React.ReactElement {
  const { locale } = useContext(Context);
  const {
    authgearOnceLicenseKey,
    authgearOnceLicenseeEmail,
    authgearOnceLicenseExpireAt,
  } = useSystemConfig();
  const expireAt: Date | null =
    authgearOnceLicenseExpireAt !== ""
      ? new Date(authgearOnceLicenseExpireAt)
      : null;

  return (
    <ScreenContent>
      <div className={cn(styles.widget, styles.pageHeader)}>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="LicenseScreen.title" />
        </Heading>
        <Text as="p" size="2" color="gray" className={styles.pageDescription}>
          <FormattedMessage id="LicenseScreen.description" />
        </Text>
      </div>
      <div className={styles.widget}>
        <TextField
          size="2"
          labelSize="2"
          label={<FormattedMessage id="LicenseScreen.license-key" />}
          readOnly={true}
          value={authgearOnceLicenseKey}
        />
      </div>
      <div className={styles.widget}>
        <TextField
          size="2"
          labelSize="2"
          label={<FormattedMessage id="LicenseScreen.email" />}
          readOnly={true}
          value={authgearOnceLicenseeEmail}
        />
      </div>
      <div className={styles.widget}>
        <Text as="p" size="2" weight="medium">
          <FormattedMessage id="LicenseScreen.lifetime-usage" />
        </Text>
        <Text as="p" size="2">
          {expireAt != null ? (
            <FormattedMessage
              id="LicenseScreen.license-expiry"
              values={{
                expireAt: formatDateOnly(locale, expireAt) ?? "",
              }}
            />
          ) : null}
        </Text>
      </div>
    </ScreenContent>
  );
}

export default LicenseScreen;
