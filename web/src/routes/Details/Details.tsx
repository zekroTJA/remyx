import { Component, For, Show, createEffect, createSignal } from "solid-js";
import { NavLink, useParams } from "@solidjs/router";

import ArrowLeft from "../../assets/allowleft";
import { Remyx } from "../../services/api/models";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import styles from "./Details.module.scss";
import { useApi } from "../../hooks/useApi";

export const Details: Component = () => {
  const fetch = useApi();
  const [remyx, setRemyx] = createSignal<Remyx>();
  const { id } = useParams();

  createEffect(() => {
    if (!id) return;
    fetch((c) => c.remyx(id)).then((r) => setRemyx(r));
  });

  const _setRemyx = (v: Partial<Remyx>) => {
    const n = { ...remyx()!, ...v };
    setRemyx(n);
  };

  const _update = () => {
    if (!id) return;
    fetch((c) => c.updateRemyx(id, remyx()?.name, remyx()?.head));
  };

  return (
    <RouteContainer>
      <div class={styles.container}>
        <Show when={remyx()}>
          <div class="flex gap">
            <NavLink href="-1" class="navButton">
              <ArrowLeft height="2em" width="2em" />
            </NavLink>
            <h2>{remyx()?.name ?? <i>Unnamed remyx</i>}</h2>
          </div>
          <div class={styles.shareLink}>
            <span>
              You can use this link to share it with your friends so they can
              collaborate on this Remyx.
            </span>
            <span>{`${window.location.origin}/c/${remyx()?.uid}`}</span>
          </div>
          <section>
            <label for="iName">Name</label>
            <input
              id="iName"
              value={remyx()?.name ?? ""}
              onInput={(e) =>
                _setRemyx({
                  name: !!e.currentTarget.value
                    ? e.currentTarget.value
                    : undefined,
                })
              }
            />
          </section>
          <section>
            <label for="iHead">Song Count (per Source Playlist)</label>
            <input
              id="iHead"
              type="number"
              min="0"
              max="50"
              value={remyx()?.head ?? 20}
              onInput={(e) =>
                _setRemyx({ head: parseInt(e.currentTarget.value) })
              }
            />
          </section>
          <button class="button" onClick={_update}>
            Save Changes
          </button>
          <section>
            <h3>Source Playlists</h3>
            <div class="playlistList">
              <For each={remyx()?.playlists}>
                {(item) => (
                  <div>
                    <span>{item.name}</span>
                    {item.owner_name && (
                      <span class="owner">{item.owner_name}</span>
                    )}
                  </div>
                )}
              </For>
            </div>
          </section>
        </Show>
      </div>
    </RouteContainer>
  );
};
