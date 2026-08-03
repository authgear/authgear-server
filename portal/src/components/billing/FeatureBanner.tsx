import React from "react";
import {
  CheckCircledIcon,
  CheckIcon,
  LayersIcon,
  RocketIcon,
} from "@radix-ui/react-icons";
import { Text as RadixText } from "@radix-ui/themes";
import styles from "./FeatureBanner.module.css";
import { FormattedMessage } from "../../intl";

interface FeatureBannerProps {}

export function FeatureBanner({}: FeatureBannerProps): React.ReactElement {
  return (
    <div className={styles.bannerContainer}>
      <div className={styles.contentSection}>
        <div className={styles.header}>
          <RadixText size="8" as="p" className={styles.title}>
            <FormattedMessage id="FeatureBanner.title" />
          </RadixText>
          <RadixText size="4" as="p" className={styles.subtitle}>
            <FormattedMessage id="FeatureBanner.subtitle" />
          </RadixText>
        </div>
        <FeatureList />
      </div>
      <div className={styles.highlightedSection}>
        <HighlightedFeatureList />
      </div>
    </div>
  );
}

const featureMessageIDs = [
  "FeatureBanner.features.customizeSignInPage",
  "FeatureBanner.features.unlimitedSocialLogin",
  "FeatureBanner.features.unlimitedMFA",
  "FeatureBanner.features.rbac",
  "FeatureBanner.features.customDomain",
  "FeatureBanner.features.loginWithSMSWhatsappOTP",
  "FeatureBanner.features.unlimitedHooks",
  "FeatureBanner.features.twoEnvironments",
  "FeatureBanner.features.botProtection",
  "FeatureBanner.features.iso27001AndSoc2Compliance",
];

function FeatureList() {
  return (
    <ul className={styles.featureList}>
      {featureMessageIDs.map((id) => (
        <li key={id} className={styles.featureItem}>
          <CheckIcon className={styles.featureIcon} width="1rem" height="1rem" />
          <RadixText size="2" weight="medium" className={styles.featureText}>
            <FormattedMessage id={id} />
          </RadixText>
        </li>
      ))}
    </ul>
  );
}

type HighlightedFeatureIcon = typeof CheckCircledIcon;

const highlightedFeatures: {
  messageID: string;
  icon: HighlightedFeatureIcon;
}[] = [
  {
    messageID: "FeatureBanner.highlightedFeatures.fullAccessToAllFeatures",
    icon: CheckCircledIcon,
  },
  {
    messageID: "FeatureBanner.highlightedFeatures.startBuildingForFree",
    icon: RocketIcon,
  },
  {
    messageID: "FeatureBanner.highlightedFeatures.flexibleUsageBasedAddOns",
    icon: LayersIcon,
  },
];

function HighlightedFeatureList() {
  return (
    <ul className={styles.highlightedFeatureList}>
      {highlightedFeatures.map((feature) => {
        const Icon = feature.icon;
        return (
          <li key={feature.messageID} className={styles.highlightedFeatureItem}>
            <div className={styles.highlightedFeatureIconBox}>
              <Icon
                className={styles.highlightedFeatureIcon}
                width="1.875rem"
                height="1.875rem"
              />
            </div>
            <RadixText size="3" weight="medium" className={styles.highlightedFeatureText}>
              <FormattedMessage id={feature.messageID} />
            </RadixText>
          </li>
        );
      })}
    </ul>
  );
}
