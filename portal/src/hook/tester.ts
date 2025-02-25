import { useCallback, useState } from "react";
import { useGenerateTesterTokenMutation } from "../graphql/portal/mutations/generateTesterTokenMutation";

export function useTester(
  appID: string,
  publicOrigin: string
): { triggerTester: () => Promise<void>; isLoading: boolean } {
  const [isLoading, setIsLoading] = useState(false);
  const { generateTesterToken } = useGenerateTesterTokenMutation(appID);
  const triggerTester = useCallback(async () => {
    setIsLoading(true);
    try {
      const token = await generateTesterToken(window.location.href);
      const destination = new URL(publicOrigin);
      destination.pathname = "/tester";
      destination.search = new URLSearchParams({ token }).toString();
      window.location.assign(destination);
    } finally {
      setIsLoading(false);
    }
  }, [generateTesterToken, publicOrigin]);

  return {
    triggerTester,
    isLoading,
  };
}
