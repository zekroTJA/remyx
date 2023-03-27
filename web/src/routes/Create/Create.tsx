import { Component, For, createEffect, createSignal } from "solid-js";
import { NavLink, useNavigate } from "@solidjs/router";

import { Playlist } from "../../services/api/models";
import styles from "./Create.module.scss";
import { useApi } from "../../hooks/useApi";

export const Create: Component = () => {
  const fetch = useApi();
  const nav = useNavigate();
  const [playlists, setPlaylists] = createSignal<Playlist[]>();
  const [head, setHead] = createSignal<number>(20);

  createEffect(() => {
    fetch((c) => c.playlists(100)).then((r) => setPlaylists(r));
  });

  const createRemyx = (pl: string) => {
    fetch((c) => c.createRemyx(pl, head())).then((r) => nav(`/${r.uid}`));
  };

  return (
    <div class={styles.container}>
      <section>
        <label for="iHead">Head</label>
        <input
          id="iHead"
          type="number"
          min="0"
          max="50"
          value={head()}
          onInput={(e) => setHead(parseInt(e.currentTarget.value))}
        />
      </section>
      <section>
        <label>Playlist</label>
        <div class={styles.list}>
          <For
            each={playlists()}
            fallback={<>Seems like you have no playlists.</>}
          >
            {(item) => (
              <button onClick={() => createRemyx(item.uid)}>{item.name}</button>
            )}
          </For>
        </div>
      </section>
    </div>
  );
};
