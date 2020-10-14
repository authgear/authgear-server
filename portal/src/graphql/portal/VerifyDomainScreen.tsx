import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";

import styles from "./VerifyDomainScreen.module.scss";

const VerifyDomainScreen: React.FC = function VerifyDomainScreen() {
  const navBreadcrumbItems = useMemo(() => {
    return [
      {
        to: "../..",
        label: <FormattedMessage id="DNSConfigurationScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="VerifyDomainScreen.title" /> },
    ];
  }, []);

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
    </main>
  );
};

export default VerifyDomainScreen;
