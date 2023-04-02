export type Playlist = {
  uid: string;
  name: string;
  description: string;
  url: string;
  image_url: string;
  n_tracks: number;
  owner_name?: string;
};

export type RemyxCreateResponse = {
  uid: string;
  expires: string;
};

export type Entity = {
  uid: string;
  created_at: string;
};

export type Remyx = Entity & {
  creator_uid: string;
  head: number;
  name?: string;
  playlist_count?: number;
  expires?: string;
  playlists?: Playlist[];
  mine?: boolean;
};

export type MyRemyxes = {
  created: Remyx[];
  connected: Remyx[];
};

export type RemyxPlaylistDeleteResponse = {
  remyx_deleted: boolean;
};
