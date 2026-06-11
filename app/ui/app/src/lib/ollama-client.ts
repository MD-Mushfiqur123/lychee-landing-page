import { Lychee as LycheeClient } from "lychee/browser";
import { LYCHEE_HOST } from "./config";

let _lycheeClient: LycheeClient | null = null;

export const lycheeClient = new Proxy({} as LycheeClient, {
  get(_target, prop) {
    if (!_lycheeClient) {
      _lycheeClient = new LycheeClient({
        host: LYCHEE_HOST,
      });
    }
    const value = _lycheeClient[prop as keyof LycheeClient];
    return typeof value === "function" ? value.bind(_lycheeClient) : value;
  },
});
