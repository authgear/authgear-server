import { Group } from "../graphql/adminapi/globalTypes.generated";

export interface SearchableGroup extends Pick<Group, "id" | "key" | "name"> {}

export function searchGroups<G extends SearchableGroup>(
  groups: G[],
  searchKeyword: string
): G[] {
  if (searchKeyword === "") {
    return groups;
  }
  const keywords = searchKeyword
    .toLowerCase()
    .split(" ")
    .flatMap((keyword) => {
      const trimmedKeyword = keyword.trim();
      if (trimmedKeyword) {
        return [trimmedKeyword];
      }
      return [];
    });
  return groups.filter((group) => {
    const groupID = group.id.toLowerCase();
    const groupKey = group.key.toLowerCase();
    const groupName = group.name?.toLowerCase();
    for (const keyword of keywords) {
      if (groupID === keyword) {
        return true;
      }
      if (groupKey.includes(keyword)) {
        return true;
      }
      if (groupName?.includes(keyword)) {
        return true;
      }
    }
    return false;
  });
}
