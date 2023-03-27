import { Component, For, createEffect, createSignal } from "solid-js";
import { Playlist, Remyx } from "../../services/api/models";
import { useNavigate, useParams } from "@solidjs/router";

import styles from "./Connect.module.scss";
import { useApi } from "../../hooks/useApi";

export const Connect: Component = () => {
  const fetch = useApi();
  const nav = useNavigate();
  const [playlists, setPlaylists] = createSignal<Playlist[]>();
  const { id } = useParams();

  createEffect(() => {
    fetch((c) => c.playlists(100)).then((r) => setPlaylists(r));
  });

  const connectRemyx = (playlistid: string) => {
    if (!id) return;
    fetch((c) => c.connectRemyx(id, playlistid)).then(() => nav("/"));
  };

  return (
    <div class={styles.container}>
      <section>
        <label>Playlist</label>
        <div class={styles.list}>
          <For
            each={playlists()}
            fallback={<>Seems like you have no playlists.</>}
          >
            {(item) => (
              <button onClick={() => connectRemyx(item.uid)}>
                {item.name}
              </button>
            )}
          </For>
        </div>
      </section>
    </div>
  );
};
