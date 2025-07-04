export default async function (e: any): Promise<any> {
  return {
    is_allowed: true,
    rate_limits: {
      "authentication.account_enumeration": {
        weight: 100,
      },
    },
  };
}
