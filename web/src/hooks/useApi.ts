import { APIError } from "../services/api/errors";
import { Client } from "../services/api/client";
import { NotificationContext } from "../services/notifications/notifications";
import { useContext } from "solid-js";
import { useNavigate } from "@solidjs/router";

const API_CLIENT = new Client();

export const useApi = () => {
  const nav = useNavigate();
  const notifications = useContext(NotificationContext);

  async function fetch<T>(req: (c: Client) => Promise<T>): Promise<T> {
    try {
      return await req(API_CLIENT);
    } catch (e) {
      if (e instanceof APIError) {
        switch (e.code) {
          case 401:
            nav("/login");
            break;
          case 500:
          case 501:
          case 502:
            notifications?.show(
              "error",
              `Oops, something went wrong (${e.message}).`,
              8000
            );
            break;
          default:
            throw e;
        }
        return {} as T;
      }
      throw e;
    }
  }

  return fetch;
};
