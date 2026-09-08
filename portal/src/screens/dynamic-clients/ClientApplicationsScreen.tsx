import React, { useCallback } from "react";
import cn from "classnames";
import { Heading } from "@radix-ui/themes";
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
import styles from "./Applications.module.css";

const ClientApplicationsScreen: React.VFC =
  function ClientApplicationsScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);

    // The plan's client limits gate the Add button, so a failed feature-config
    // load is a load failure for the page, not a quota of undefined.
    const loadError = form.loadError ?? featureConfig.loadError;
    const onRetry = useCallback(() => {
      if (form.loadError) {
        form.reload();
      }
      if (featureConfig.loadError) {
        featureConfig.refetch().finally(() => {});
      }
    }, [form, featureConfig]);

    if (form.isLoading || featureConfig.isLoading) {
      return <ShowLoading />;
    }
    if (loadError) {
      return <ShowError error={loadError} onRetry={onRetry} />;
    }

    // FormContainer renders a whole page shell, so it stays outside
    // ScreenContent -- the other way round drops a page layout into a grid
    // cell and the content collapses.
    return (
      <FormContainer form={form}>
        <ScreenContent layout="list">
          <div className={cn(styles.widget, styles.pageHeader)}>
            <Heading
              as="h1"
              size="5"
              weight="bold"
              className={styles.pageTitle}
            >
              <FormattedMessage id="ClientApplicationsScreen.title" />
            </Heading>
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

export default ClientApplicationsScreen;
