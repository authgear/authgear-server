import React, { useCallback, useContext, useMemo } from "react";
import { Dropdown, PrimaryButton } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useDropdown } from "../../hook/useInput";

import { IdentityLists } from "./UserDetailsConnectedIdentities";

import styles from "./PrimaryIdentitiesSelectionForm.module.scss";
import deepEqual from "deep-equal";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";

interface PrimaryIdentitiesSelectionFormProps {
  className?: string;
  identityLists: IdentityLists;
}

const PrimaryIdentitiesSelectionForm: React.FC<PrimaryIdentitiesSelectionFormProps> = function PrimaryIdentitiesSelectionForm(
  props: PrimaryIdentitiesSelectionFormProps
) {
  const { className, identityLists } = props;
  const { renderToString } = useContext(Context);
  const {
    selectedKey: selectedPrimaryEmail,
    options: primaryEmailOptions,
    onChange: onPrimaryEmailOptionsChange,
  } = useDropdown(identityLists.email.map((identity) => identity.value));

  const {
    selectedKey: selectedPrimaryPhone,
    options: primaryPhoneOptions,
    onChange: onPrimaryPhoneOptionsChange,
  } = useDropdown(identityLists.phone.map((identity) => identity.value));

  const {
    selectedKey: selectedPrimaryUsername,
    options: primaryUsernameOptions,
    onChange: onPrimaryUsernameOptionsChange,
  } = useDropdown(identityLists.username.map((identity) => identity.value));

  const primaryIdentitiesState = useMemo(
    () => ({
      email: selectedPrimaryEmail,
      phone: selectedPrimaryPhone,
      username: selectedPrimaryUsername,
    }),
    [selectedPrimaryEmail, selectedPrimaryPhone, selectedPrimaryUsername]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(
      {
        email: undefined,
        phone: undefined,
        username: undefined,
      },
      primaryIdentitiesState
    );
  }, [primaryIdentitiesState]);

  const onSaveClicked = useCallback(() => {
    // TODO: to be implemented
  }, []);

  return (
    <div className={className}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <section className={styles.primaryIdentities}>
        <Dropdown
          className={styles.primaryEmail}
          disabled={identityLists.email.length === 0}
          options={primaryEmailOptions}
          onChange={onPrimaryEmailOptionsChange}
          label={renderToString(
            "UserDetails.connected-identities.primary-email"
          )}
        />
        <Dropdown
          className={styles.primaryPhone}
          disabled={identityLists.phone.length === 0}
          options={primaryPhoneOptions}
          onChange={onPrimaryPhoneOptionsChange}
          label={renderToString(
            "UserDetails.connected-identities.primary-phone"
          )}
        />
        <Dropdown
          className={styles.primaryUsername}
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
