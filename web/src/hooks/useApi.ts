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
            nav("/login", {
              state: {
                redirect:
                  window.location.pathname != "/"
                    ? window.location.pathname.substring(1)
                    : undefined,
              },
            });
            break;
          case 429:
            console.log("TEST RATE LIMIT");
            notifications?.show(
              "error",
              "Oh snap, you have been rate limited! Please try again later."
            );
            throw e;
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
