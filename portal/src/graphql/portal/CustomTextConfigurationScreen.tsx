import React, { useMemo } from "react";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { ResourceSpecifier } from "../../util/resource";
import {
  DEFAULT_TEMPLATE_LOCALE,
  RESOURCE_TRANSLATION_JSON,
} from "../../resources";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";

const CustomTextConfigurationScreen: React.VFC =
  function CustomTextConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const config = useAppAndSecretConfigQuery(appID);

    const initialSupportedLanguages = useMemo(() => {
      return (
        config.effectiveAppConfig?.localization?.supported_languages ?? [
          config.effectiveAppConfig?.localization?.fallback_language ??
            DEFAULT_TEMPLATE_LOCALE,
        ]
      );
    }, [config.effectiveAppConfig?.localization]);

    const specifiers = useMemo<ResourceSpecifier[]>(() => {
      const specifiers = [];

      const supportedLanguages = [...initialSupportedLanguages];
      if (!supportedLanguages.includes(DEFAULT_TEMPLATE_LOCALE)) {
        supportedLanguages.push(DEFAULT_TEMPLATE_LOCALE);
      }

      for (const locale of supportedLanguages) {
        specifiers.push({
          def: RESOURCE_TRANSLATION_JSON,
          locale,
          extension: null,
        });
      }
      return specifiers;
    }, [initialSupportedLanguages]);

    const resourceForm = useResourceForm(appID, specifiers);

    return <FormContainer form={resourceForm} canSave={true}></FormContainer>;
  };

export default CustomTextConfigurationScreen;
