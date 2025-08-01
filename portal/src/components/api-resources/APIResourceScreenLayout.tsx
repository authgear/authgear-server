import React from "react";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "../../ErrorMessageBar";
import ScreenContent from "../../ScreenContent";
import ScreenContentHeader from "../../ScreenContentHeader";

export interface APIResourceLayoutProps {
  children?: React.ReactNode;
  breadcrumbItems: BreadcrumbItem[];
  headerDescription?: React.ReactNode;
  headerSuffix?: React.ReactNode;
}

const APIResourceScreenLayout: React.VFC<APIResourceLayoutProps> =
  function APIResourceScreenLayout({
    children,
    breadcrumbItems,
    headerDescription,
    headerSuffix,
  }) {
    return (
      <ErrorMessageBarContextProvider>
        <div className="flex-1 flex flex-col">
          <ErrorMessageBar />
          <ScreenContent className="flex-1" layout="list">
            <ScreenContentHeader
              title={<NavBreadcrumb items={breadcrumbItems} />}
              description={headerDescription}
              suffix={headerSuffix}
            />
            {children}
          </ScreenContent>
        </div>
      </ErrorMessageBarContextProvider>
    );
  };

export default APIResourceScreenLayout;
