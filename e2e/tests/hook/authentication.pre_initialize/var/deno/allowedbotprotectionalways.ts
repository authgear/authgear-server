export default async function (e: any): Promise<any> {
  return {
    is_allowed: true,
    bot_protection: {
      mode: "always",
    },
  };
}
