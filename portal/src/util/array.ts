export function deduplicate<Type>(arr: Type[]): Type[] {
  return arr.filter((item, index) => arr.indexOf(item) === index);
}
