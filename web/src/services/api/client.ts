import {
  MyRemyxes,
  Playlist,
  Remyx,
  RemyxCreateResponse,
  RemyxPlaylistDeleteResponse,
} from "./models";

import { APIError } from "./errors";

const ENDPOINT = (import.meta.env.VITE_SERVER_ADDRESS ?? "") + "/api";

export const loginUrl = (redirect?: string) =>
  `${ENDPOINT}/oauth/login${!!redirect ? "?redirect=" + redirect : ""}`;

export class Client {
  constructor() {}

  playlists(limit: number = 20, offset: number = 0): Promise<Playlist[]> {
    return this.req("GET", `playlists?limit=${limit}&offset=${offset}`);
  }

  remyx(id: string): Promise<Remyx> {
    return this.req("GET", `remyxes/${id}`);
  }

  remyxes(): Promise<MyRemyxes> {
    return this.req("GET", "remyxes");
  }

  createRemyx(
    playlist_id: string,
    head: number,
    name?: string
  ): Promise<RemyxCreateResponse> {
    return this.req("POST", "remyxes/create", { playlist_id, name, head });
  }

  updateRemyx(id: string, name?: string, head?: number): Promise<Remyx> {
    return this.req("POST", `remyxes/${id}`, { name, head });
  }

  deleteRemyx(id: string): Promise<{}> {
    return this.req("DELETE", `remyxes/${id}`);
  }

  deleteRemyxPlaylist(
    id: string,
    playlistId: string
  ): Promise<RemyxPlaylistDeleteResponse> {
    return this.req("DELETE", `remyxes/${id}/${playlistId}`);
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
      let err: Error | undefined = undefined;
      try {
        err = await res.json();
      } catch {}
      throw new APIError(res, res.status, err?.message ?? res.statusText);
    }

    if (res.status != 204) {
      return await res.json();
    }

    return {} as T;
  }
}
