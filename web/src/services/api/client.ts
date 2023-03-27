import { Playlist, Remyx, RemyxCreateResponse } from "./models";

import { APIError } from "./errors";

const ENDPOINT =
  (import.meta.env.VITE_SERVER_ADDRESS ?? "http://localhost:8081") + "/api";

export const loginUrl = () => `${ENDPOINT}/oauth/login`;

export class Client {
  constructor() {}

  playlists(limit: number = 20, offset: number = 0): Promise<Playlist[]> {
    return this.req("GET", `playlists?limit=${limit}&offset=${offset}`);
  }

  remyx(id: string): Promise<Remyx> {
    return this.req("GET", `remyxes/${id}`);
  }

  remyxes(): Promise<Remyx[]> {
    return this.req("GET", "remyxes");
  }

  createRemyx(playlist_id: string, head: number): Promise<RemyxCreateResponse> {
    return this.req("POST", "remyxes/create", { playlist_id, head });
  }

  connectRemyx(id: string, playlist_id: string): Promise<RemyxCreateResponse> {
    return this.req("POST", `remyxes/connect/${id}`, { playlist_id });
  }

  private async req<T>(
    method: string,
    path: string,
    payload?: any
  ): Promise<T> {
    const res = await fetch(`${ENDPOINT}/${path}`, {
      body: !!payload ? JSON.stringify(payload) : undefined,
      credentials: "include",
      method,
    });

    if (res.status >= 400) {
      throw new APIError(res, res.status, res.statusText);
    }

    if (res.status != 203) {
      return await res.json();
    }

    return {} as T;
  }
}
