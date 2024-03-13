export interface Searchable {
  id: string 
  key: string
  name?: string | null
}

export function searchRolesAndGroups<T extends Searchable>(
  data: T[],
  searchKeyword: string
): T[] {
  if(searchKeyword === ""){
    return data
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
  return data.filter((item) => {
    const id = item.id.toLowerCase();
    const key = item.key.toLowerCase();
    const name = item.name?.toLowerCase();
    for (const keyword of keywords) {
      if (id === keyword) {
        return true;
      }
      if (key.includes(keyword)) {
        return true;
      }
      if (name?.includes(keyword)) {
        return true;
      }
    }
    return false;
  });
}
