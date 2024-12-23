import React, {
  Suspense,
  SuspenseProps,
  lazy,
  useCallback,
  useMemo,
  useState,
  useEffect,
} from "react";
import { ErrorBoundary, FallbackRender } from "@sentry/react";

export interface ErrorBoundSuspenseProps {
  errorBoundaryFallback: FallbackRender;
  suspenseFallback: SuspenseProps["fallback"];
  factory: () => Promise<{ default: React.ComponentType<any> }>;
  children: (Component: React.ComponentType<any>) => React.ReactNode;
}

export type ErrorBoundaryFallbackProps = Parameters<FallbackRender>[0];

const extractImportArgRegex = /import\(["']([-_./a-zA-Z0-9]+)["']\)/;

function extractURLToModuleFromSourceCode(
  factory: () => Promise<{ default: React.ComponentType<any> }>
): URL | null {
  // The sourceCode contains a relative URL.
  // It is relative to the original module that loads it.
  // For example, the original module that loads it may be
  // https://portal/shared-assets/index-deadbeef.js
  // The relative URL is "./SomeScreen-deadbeef.js"
  //
  // If we call import directly on the relative URL,
  // it is relative to window.location, that is https://portal/project/PROJECT_ID/configuration/authentication
  // Thus load https://portal/project/PROJECT_ID/configuration/authentcation/SomeScreen-deadbeef.js
  // That will give a index.html because we are a SPA.
  //
  // To get the absolute URL to the module,
  // we have to know the URL to the original module.
  // This is impossible because we cannot know it from here.
  //
  // Fortunately, we can make an assumption here.
  // Since all assets are put in the assets directory,
  // if we know the URL to the current module (that is, this file),
  // then we can assume that the original module is next to this file.

  try {
    // We are doing some arcane art here.
    // To know the URL to load the module,
    // we call toString on factory.
    // That is, we peek the compiled source code of the function!
    const sourceCode = factory.toString();
    const match = extractImportArgRegex.exec(sourceCode);
    if (match != null) {
      const relativePath = match[1];
      const urlToThisFile = import.meta.url;
      // new URL("./SomeScreen.js", "https://localhost/shared-assets/index.js") =>
      // https://localhost/shared-assets/SomeScreen.js
      const url = new URL(relativePath, urlToThisFile);
      return url;
    }
  } catch (_e: unknown) {
    // ignore
  }

  return null;
}

function makeFactory(urlToModule: URL) {
  // Browsers cache modules, including failed attempt.
  // In order to force them to retry, we need to convince them that the module is different.
  // The simplest way is that attach a query to the URL.
  const patchedURL = new URL(urlToModule);
  patchedURL.searchParams.set("t", new Date().getTime().toString());
  // Tell vite that this particular import() call is intentionally dynamic.
  return async () => import(/* @vite-ignore */ patchedURL.toString());
}

// ErrorBoundSuspense is known to have this bug.
// https://linear.app/authgear/issue/DEV-2387/wrong-path-when-no-application-in-portal
// Do not use it.

// ErrorBoundSuspense borrows the idea from https://github.com/cj0x39e/retrying-dynamic-import/blob/main/packages/retrying-dynamic-import/src/index.ts
// There is one difference though.
// ErrorBoundSuspense currently can only reload failed JavaScript module.
// It does not reload failed CSS.
// The main reason is failed CSS does not trigger React ErrorBoundary, and
// we do not know a dependent CSS file failed to load.
//
// To stimulate failed module load locally, you can use the following diff.
// It deliberately fails the first attempt (the attempt without ?t query).
//
// diff --git i/pkg/util/httputil/file_server.go w/pkg/util/httputil/file_server.go
// index 3cb6fec0c..1b1528429 100644
// --- i/pkg/util/httputil/file_server.go
// +++ w/pkg/util/httputil/file_server.go
// @@ -3,6 +3,7 @@ package httputil
//  import (
//         "bytes"
//         "errors"
// +       "fmt"
//         htmltemplate "html/template"
//         "io"
//         "io/fs"
// @@ -150,6 +151,9 @@ func (s *FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//         // If the request fetches a non-name-hashed file,
//         // we fallback to index.html for not found.
//         if strings.HasPrefix(r.URL.Path, "/"+s.AssetsDir) {
// +               if strings.Contains(r.URL.Path, "Screen-") && strings.HasSuffix(r.URL.Path, ".js") && !r.URL.Query().Has("t") {
// +                       panic(fmt.Errorf("fail deliberately"))
// +               }
//                 s.serveAsset(w, r)
//         } else {
//                 s.serveOther(w, r)
function ErrorBoundSuspense(
  props: ErrorBoundSuspenseProps
): React.ReactElement<any, any> | null {
  const { errorBoundaryFallback, suspenseFallback, factory, children } = props;

  const [LazyComponent, setLazyComponent] = useState(() => lazy(factory));

  // factory may change when route changes.
  // We have to re-create the LazyComponent when this happens.
  useEffect(() => {
    setLazyComponent(lazy(factory));

    // In some screens, for example, AuditLogScreen,
    // these screen will call setSearchParams in render.
    // Setting searchParams will cause ErrorBoundSuspense to re-render.
    // Since factory is a function, putting it in the deps array will always result in re-running the effect.
    // Thus a new LazyComponent is created, causing the screens to re-render again.
    // This will result in an infinite render loop.
    // To break this loop, we compare the identity of factory by its source code, rather than by its pointer.
    //
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [factory.toString()]);

  const urlToModule = useMemo(() => {
    return extractURLToModuleFromSourceCode(factory);
  }, [factory]);

  const onReset = useCallback(() => {
    if (urlToModule != null) {
      const newFactory = makeFactory(urlToModule);
      setLazyComponent(lazy(newFactory));
    }
  }, [urlToModule, setLazyComponent]);

  return (
    <ErrorBoundary fallback={errorBoundaryFallback} onReset={onReset}>
      <Suspense fallback={suspenseFallback}>{children(LazyComponent)}</Suspense>
    </ErrorBoundary>
  );
}

export default ErrorBoundSuspense;
