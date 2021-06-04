import React, { useState, useMemo, useCallback, useContext } from "react";
import { ICommandBarItemProps, IDropdownOption } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarDropdown, {
  CommandBarDropdownProps,
} from "../../CommandBarDropdown";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  AuditLogListQuery,
  AuditLogListQueryVariables,
} from "./__generated__/AuditLogListQuery";
import { AuditLogActivityType } from "./__generated__/globalTypes";

import styles from "./AuditLogScreen.module.scss";

const pageSize = 10;

const QUERY = gql`
  query AuditLogListQuery(
    $pageSize: Int!
    $cursor: String
    $activityTypes: [AuditLogActivityType!]
  ) {
    auditLogs(first: $pageSize, after: $cursor, activityTypes: $activityTypes) {
      edges {
        node {
          id
          createdAt
          activityType
          user {
            id
          }
        }
      }
      totalCount
    }
  }
`;

function CommandBarDropdownWrapper(props: ICommandBarItemProps) {
  const { dropdownProps } = props;
  return <CommandBarDropdown {...dropdownProps} />;
}

const AuditLogScreen: React.FC = function AuditLogScreen() {
  const [offset, setOffset] = useState(0);
  const [selectedKey, setSelectedKey] = useState("ALL");

  const { renderToString } = useContext(Context);

  const activityTypeOptions = useMemo(() => {
    const options = [
      {
        key: "ALL",
        text: renderToString("AuditLogActivityType.ALL"),
      },
    ];
    for (const key of Object.keys(AuditLogActivityType)) {
      options.push({
        key: key,
        text: renderToString("AuditLogActivityType." + key),
      });
    }
    return options;
  }, [renderToString]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (selectedKey === "ALL") {
      return null;
    }
    return [selectedKey] as AuditLogActivityType[];
  }, [selectedKey]);

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="AuditLogScreen.title" /> }];
  }, []);

  const cursor = useMemo(() => {
    if (offset === 0) {
      return null;
    }
    return encodeOffsetToCursor(offset - 1);
  }, [offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const { data, error, loading, refetch } = useQuery<
    AuditLogListQuery,
    AuditLogListQueryVariables
  >(QUERY, {
    variables: {
      pageSize,
      cursor,
      activityTypes,
    },
    fetchPolicy: "network-only",
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

  const onChangeSelectedKey = useCallback(
    (_e: React.FormEvent<HTMLDivElement>, item?: IDropdownOption) => {
      if (item != null && typeof item.key === "string") {
        setOffset(0);
        setSelectedKey(item.key);
      }
    },
    []
  );

  const dropdownProps: CommandBarDropdownProps = useMemo(() => {
    return {
      selectedKey,
      placeholder: "",
      label: "",
      options: activityTypeOptions,
      iconProps: {
        iconName: "PC1",
      },
      onChange: onChangeSelectedKey,
    };
  }, [selectedKey, onChangeSelectedKey, activityTypeOptions]);

  const commandBarFarItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "activityTypes",
        commandBarButtonAs: CommandBarDropdownWrapper,
        dropdownProps,
      },
    ];
  }, [dropdownProps]);

  return (
    <CommandBarContainer
      isLoading={loading}
      className={styles.root}
      messageBar={messageBar}
      farItems={commandBarFarItems}
    >
      <main className={styles.content}>
        <NavBreadcrumb items={items} />
        <AuditLogList
          className={styles.list}
          loading={loading}
          auditLogs={data?.auditLogs ?? null}
          offset={offset}
          pageSize={pageSize}
          totalCount={data?.auditLogs?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      </main>
    </CommandBarContainer>
  );
};

export default AuditLogScreen;
