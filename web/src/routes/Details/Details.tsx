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
          <button class="button">Save Changes</button>
          <section>
            <For each={remyx()?.playlists}>
              {(item) => <div>{item.name}</div>}
            </For>
          </section>
        </Show>
      </div>
    </RouteContainer>
  );
};
