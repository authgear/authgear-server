import React from "react";
import { Icon, Text } from "@fluentui/react";
import styles from "./FeatureBanner.module.css";
import { FormattedMessage } from "../../intl";

interface FeatureBannerProps {}

export function FeatureBanner({}: FeatureBannerProps): React.ReactElement {
  return (
    <div className={styles.bannerContainer}>
      <div className="space-y-4 flex-1-0-auto">
        <div className="space-y-2">
          <Text variant="xxLarge" block={true}>
            <FormattedMessage id="FeatureBanner.title" />
          </Text>
          <Text variant="large" className="text-text-secondary" block={true}>
            <FormattedMessage id="FeatureBanner.subtitle" />
          </Text>
        </div>
        <FeatureList />
      </div>
      <div className="flex-1 min-w-100">
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
        <li key={id} className="flex items-center">
          <Icon iconName={"CheckMark"} className="text-sm text-theme-primary" />
          <Text className="font-semibold ml-2" variant="medium">
            <FormattedMessage id={id} />
          </Text>
        </li>
      ))}
    </ul>
  );
}

const highlightedFeatures = [
  {
    messageID: "FeatureBanner.highlightedFeatures.fullAccessToAllFeatures",
    iconName: "VerifiedBrand",
  },
  {
    messageID: "FeatureBanner.highlightedFeatures.startBuildingForFree",
    iconName: "FavoriteList",
  },
  {
    messageID: "FeatureBanner.highlightedFeatures.flexibleUsageBasedAddOns",
    iconName: "ExploreContent",
  },
];

function HighlightedFeatureList() {
  return (
    <ul className={styles.highlightedFeatureList}>
      {highlightedFeatures.map((feature) => (
        <li
          key={feature.messageID}
          className="flex items-center rounded-xl bg-brand-100 px-4 py-3"
        >
          <div className="bg-brand-50 w-14 h-14 flex items-center justify-center rounded-lg">
            <Icon
              iconName={feature.iconName}
              className="text-3xl text-theme-primary"
            />
          </div>
          <Text className="font-semibold ml-4" variant="large">
            <FormattedMessage id={feature.messageID} />
          </Text>
        </li>
      ))}
    </ul>
  );
}
