import { Component, For, Show, createEffect, createSignal } from "solid-js";
import { NavLink, useNavigate, useParams } from "@solidjs/router";
import { Playlist, Remyx } from "../../services/api/models";

import { APIError } from "../../services/api/errors";
import ArrowLeft from "../../assets/allowleft";
import Clipboard from "../../assets/clipboard";
import { NotificationContext } from "../../services/notifications/notifications";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import Trashcan from "../../assets/trashcan";
import styles from "./Details.module.scss";
import { useApi } from "../../hooks/useApi";
import useContextEnsured from "../../hooks/useContextEnsured";

export const Details: Component = () => {
  const fetch = useApi();
  const nav = useNavigate();
  const { show } = useContextEnsured(NotificationContext);
  const [remyx, setRemyx] = createSignal<Remyx>();
  const { id } = useParams();

  createEffect(() => {
    if (!id) return;
    fetch((c) => c.remyx(id))
      .then((r) => setRemyx(r))
      .catch((e) => {
        console.log("TESt");
        if (e instanceof APIError && e.code === 404) {
          show("error", "There is no Remyx with this ID.");
          nav("/");
        }
      });
  });

  const _setRemyx = (v: Partial<Remyx>) => {
    const n = { ...remyx()!, ...v };
    setRemyx(n);
  };

  const _update = () => {
    if (!id) return;
    fetch((c) => c.updateRemyx(id, remyx()?.name, remyx()?.head)).then(() =>
      show("success", "Remyx has been updated.")
    );
  };

  const _deletePlaylist = (pl: Playlist) => {
    const rmx = remyx();
    if (!rmx) return;
    fetch((c) => c.deleteRemyxPlaylist(id, pl.uid)).then((r) => {
      if (r.remyx_deleted) {
        show(
          "warn",
          "The Remyx has been deleted because all source playlists have been removed."
        );
        nav("/");
      } else {
        show("success", "The source playlist has been removed.");
        _setRemyx({
          playlists: rmx.playlists?.filter(
            (p) => !(p.uid === pl.uid && p.owner_id === pl.owner_id)
          ),
        });
      }
    });
  };

  const _copySharelink = () => {
    navigator.clipboard
      .writeText(`${window.location.origin}/c/${remyx()?.uid}`)
      .then(() =>
        show("success", "Share link has been copied to your clipboard.")
      )
      .catch(() =>
        show("error", "Failed copying share link to your clipboard.")
      );
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
            <span>
              {`${window.location.origin}/c/${remyx()?.uid}`}
              <Clipboard onClick={_copySharelink} />
            </span>
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
                    {remyx()?.mine && (
                      <div class="controls">
                        <button onClick={() => _deletePlaylist(item)}>
                          <Trashcan />
                        </button>
                      </div>
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
