import React, { useCallback, useContext, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { Dropdown, PrimaryButton } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { IdentityLists } from "./UserDetailsConnectedIdentities";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { useDropdown } from "../../hook/useInput";

import styles from "./PrimaryIdentitiesSelectionForm.module.scss";

interface PrimaryIdentitiesSelectionFormProps {
  className?: string;
  identityLists: IdentityLists;
}

interface PrimaryIdentitiesSelectionFormState {
  email?: string;
  phone?: string;
  username?: string;
}

const PrimaryIdentitiesSelectionForm: React.FC<PrimaryIdentitiesSelectionFormProps> = function PrimaryIdentitiesSelectionForm(
  props: PrimaryIdentitiesSelectionFormProps
) {
  const { className, identityLists } = props;
  const { renderToString } = useContext(Context);

  const initialFormData = useMemo<PrimaryIdentitiesSelectionFormState>(() => {
    return {
      email: undefined,
      phone: undefined,
      username: undefined,
    };
  }, []);

  const [formData, setFormData] = useState(initialFormData);
  const { email, phone, username } = formData;

  const {
    options: primaryEmailOptions,
    onChange: onPrimaryEmailOptionsChange,
  } = useDropdown(
    identityLists.email.map((identity) => identity.claimValue),
    (option) => {
      setFormData((prev) => ({
        ...prev,
        email: option,
      }));
    },
    email
  );

  const {
    options: primaryPhoneOptions,
    onChange: onPrimaryPhoneOptionsChange,
  } = useDropdown(
    identityLists.phone.map((identity) => identity.claimValue),
    (option) => {
      setFormData((prev) => ({
        ...prev,
        phone: option,
      }));
    },
    phone
  );

  const {
    options: primaryUsernameOptions,
    onChange: onPrimaryUsernameOptionsChange,
  } = useDropdown(
    identityLists.username.map((identity) => identity.claimValue),
    (option) => {
      setFormData((prev) => ({
        ...prev,
        username: option,
      }));
    },
    username
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(formData, initialFormData);
  }, [formData, initialFormData]);

  const onSaveClicked = useCallback(() => {
    // TODO: to be implemented
  }, []);

  return (
    <div className={className}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <section className={styles.primaryIdentities}>
        <Dropdown
          className={styles.primaryEmail}
          selectedKey={email ?? null}
          disabled={identityLists.email.length === 0}
          options={primaryEmailOptions}
          onChange={onPrimaryEmailOptionsChange}
          label={renderToString(
            "UserDetails.connected-identities.primary-email"
          )}
        />
        <Dropdown
          className={styles.primaryPhone}
          selectedKey={phone ?? null}
          disabled={identityLists.phone.length === 0}
          options={primaryPhoneOptions}
          onChange={onPrimaryPhoneOptionsChange}
          label={renderToString(
            "UserDetails.connected-identities.primary-phone"
          )}
        />
        <Dropdown
          className={styles.primaryUsername}
          selectedKey={username ?? null}
          disabled={identityLists.username.length === 0}
          options={primaryUsernameOptions}
          onChange={onPrimaryUsernameOptionsChange}
          label={renderToString(
            "UserDetails.connected-identities.primary-username"
          )}
        />
      </section>
      <PrimaryButton disabled={!isFormModified} onClick={onSaveClicked}>
        <FormattedMessage id="save" />
      </PrimaryButton>
    </div>
  );
};

export default PrimaryIdentitiesSelectionForm;
