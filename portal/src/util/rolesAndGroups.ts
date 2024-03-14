export function processKeyword(keyword: string): string[] {
  return keyword
    .toLowerCase()
    .split(" ")
    .flatMap((keyword) => {
      const trimmedKeyword = keyword.trim();
      if (trimmedKeyword) {
        return [trimmedKeyword];
      }
      return [];
    });
}
