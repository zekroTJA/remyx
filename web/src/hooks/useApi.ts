import { APIError } from "../services/api/errors";
import { Client } from "../services/api/client";
import { useNavigate } from "@solidjs/router";

const API_CLIENT = new Client();

export const useApi = () => {
  const nav = useNavigate();

  async function fetch<T>(req: (c: Client) => Promise<T>): Promise<T> {
    try {
      return await req(API_CLIENT);
    } catch (e) {
      if (e instanceof APIError && e.code === 401) {
        nav("/login");
        return {} as T;
      }
      throw e;
    }
  }

  return fetch;
};
