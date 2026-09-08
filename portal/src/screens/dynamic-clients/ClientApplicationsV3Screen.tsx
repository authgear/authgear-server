/**
 * SIDE-BY-SIDE EXPERIMENT (variant 3) — delete this file, its CSS, its route
 * and its nav entry once one layout wins.
 *
 * Variant 3 splits the surface by audience rather than by mechanism: an
 * Applications group whose children are the static client list (here) and a
 * page for the self-onboarding ones (AIAgentsV3Screen). Compare with the
 * shipped design (ApplicationsConfigurationScreen + a nav page per
 * mechanism) and variant 2 (ApplicationsVariantScreen).
 *
 * The list itself is the same component the shipped screen renders, so only
 * the arrangement differs.
 */
import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { useAppFeatureConfigQuery } from "../../graphql/portal/query/appFeatureConfigQuery";
import {
  ApplicationsListSection,
  constructConfig,
  constructFormState,
} from "../../graphql/portal/ApplicationsListSection";
import styles from "./ApplicationsV3.module.css";

const ClientApplicationsV3Screen: React.VFC =
  function ClientApplicationsV3Screen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.isLoading || featureConfig.isLoading) {
      return <ShowLoading />;
    }
    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    // FormContainer renders a whole page shell, so it stays outside
    // ScreenContent -- the other way round drops a page layout into a grid
    // cell and the content collapses.
    return (
      <FormContainer form={form}>
        <ScreenContent layout="list">
          <div className={cn(styles.widget, styles.pageHeader)}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="ClientApplicationsV3Screen.title" />
            </Text>
          </div>
          <ApplicationsListSection
            form={form}
            planName={featureConfig.planName}
            oauthClientsSoftMaximum={
              featureConfig.effectiveFeatureConfig?.oauth?.client?.soft_maximum
            }
            oauthClientsHardMaximum={
              featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum
            }
          />
        </ScreenContent>
      </FormContainer>
    );
  };

export default ClientApplicationsV3Screen;
