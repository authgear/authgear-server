import React, { useContext } from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { Label, Text, useTheme } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import TextField from "../../TextField";
import { formatDateOnly } from "../../util/formatDateOnly";
import styles from "./LicenseScreen.module.css";

function LicenseScreen(): React.ReactElement {
  const theme = useTheme();
  const { locale, renderToString } = useContext(Context);
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
      <ScreenTitle className={styles.widget}>
        <FormattedMessage id="LicenseScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="LicenseScreen.description" />
      </ScreenDescription>
      <TextField
        className={styles.widget}
        label={renderToString("LicenseScreen.license-key")}
        readOnly={true}
        value={authgearOnceLicenseKey}
        styles={{
          field: {
            backgroundColor: theme.semanticColors.disabledBackground,
          },
        }}
      />
      <TextField
        className={styles.widget}
        label={renderToString("LicenseScreen.email")}
        readOnly={true}
        value={authgearOnceLicenseeEmail}
        styles={{
          field: {
            backgroundColor: theme.semanticColors.disabledBackground,
          },
        }}
      />
      <div className={styles.widget}>
        <Label>
          <FormattedMessage id="LicenseScreen.lifetime-usage" />
        </Label>
        <Text as="p">
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
