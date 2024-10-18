import React, { useCallback, useContext, useMemo } from "react";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import { SAMLIdpSigningCertificate } from "../../types";
import { FormState } from "../../hook/useSAMLCertificateForm";
import WidgetTitle from "../../WidgetTitle";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import {
  DetailsList,
  IColumn,
  SelectionMode,
  Text,
  ChoiceGroup,
  IChoiceGroupOption,
} from "@fluentui/react";
import LinkButton from "../../LinkButton";
import { downloadStringAsFile } from "../../util/download";

interface EditSAMLCertificateFormProps {
  form: AppSecretConfigFormModel<FormState>;
  certificates: SAMLIdpSigningCertificate[];
}

export function EditSAMLCertificateForm({
  form,
  certificates,
}: EditSAMLCertificateFormProps): React.ReactElement {
  const { onSubmit } = useFormContainerBaseContext();
  const { renderToString } = useContext(MessageContext);

  const onClickDownloadCert = useMemo(() => {
    const callbacks: Record<string, () => void> = {};
    for (const cert of certificates) {
      callbacks[cert.keyID] = () => {
        downloadStringAsFile({
          content: cert.certificatePEM,
          mimeType: "application/x-pem-file",
          filename: `${cert.certificateFingerprint}.pem`,
        });
      };
    }
    return callbacks;
  }, [certificates]);

  const choiceGroupOptionsByKeyID = useMemo(() => {
    const optionsByKeyID: Record<string, IChoiceGroupOption[]> = {};
    for (const cert of certificates) {
      optionsByKeyID[cert.keyID] = [
        {
          key: cert.keyID,
          text: renderToString(
            "EditSAMLCertificateForm.certificates.column.status.active"
          ),
        },
      ];
    }
    return optionsByKeyID;
  }, [certificates, renderToString]);

  const onChangeActiveKey = useCallback(
    (_: unknown, option?: IChoiceGroupOption) => {
      if (!option) {
        return;
      }
      form.setState((prevState) => ({ ...prevState, activeKeyID: option.key }));
    },
    [form]
  );

  const columns: IColumn[] = useMemo(() => {
    const renderFingerprint = (
      item?: SAMLIdpSigningCertificate,
      _index?: number,
      _column?: IColumn
    ) => {
      if (!item) {
        return null;
      }
      return (
        <div className="grid grid-cols-1 gap-y-2">
          <Text className="text-neutral-secondary" block={true}>
            {item.certificateFingerprint}
          </Text>
          <LinkButton
            styles={{ root: { fontSize: 14 } }}
            onClick={onClickDownloadCert[item.keyID]}
          >
            <FormattedMessage id="EditSAMLCertificateForm.certificates.download" />
          </LinkButton>
        </div>
      );
    };
    const renderStatus = (
      item?: SAMLIdpSigningCertificate,
      _index?: number,
      _column?: IColumn
    ) => {
      if (!item) {
        return null;
      }
      return (
        <ChoiceGroup
          options={choiceGroupOptionsByKeyID[item.keyID]}
          selectedKey={form.state.activeKeyID}
          onChange={onChangeActiveKey}
        />
      );
    };
    return [
      {
        key: "certificateFingerprint",
        fieldName: "certificateFingerprint",
        name: renderToString(
          "EditSAMLCertificateForm.certificates.column.fingerprint"
        ),
        minWidth: 150,
        onRender: renderFingerprint,
      },
      {
        key: "status",
        name: renderToString(
          "EditSAMLCertificateForm.certificates.column.status"
        ),
        minWidth: 150,
        onRender: renderStatus,
      },
    ];
  }, [
    renderToString,
    onClickDownloadCert,
    choiceGroupOptionsByKeyID,
    form.state.activeKeyID,
    onChangeActiveKey,
  ]);

  return (
    <form onSubmit={onSubmit}>
      <WidgetTitle>
        <FormattedMessage id="EditSAMLCertificateForm.certificates.title" />
      </WidgetTitle>
      <DetailsList
        items={certificates}
        columns={columns}
        selectionMode={SelectionMode.none}
      />
    </form>
  );
}
