// Since we define `.postcssrc.json`, we cannot use `.module.` to opt-in CSS module on a per-file basis.
// .module.scss is just a naming convention.
// All CSS files are parsed with CSS module, even the filename does not contain `.module.`.
declare module "*.module.scss" {
  const classes: { [key: string]: string };
  export default classes;
}
