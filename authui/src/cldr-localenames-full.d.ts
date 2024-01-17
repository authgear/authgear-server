declare module "cldr-localenames-full/main/*/territories.json" {
  import defaultTerritories from "cldr-localenames-full/main/en/territories.json";

  const territoriesMap: Record<string, typeof defaultTerritories>;
  export default territoriesMap;
}
