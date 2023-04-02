import { Component, For, createEffect, createSignal } from "solid-js";
import { Playlist, Remyx } from "../../services/api/models";
import { useNavigate, useParams } from "@solidjs/router";

import { NotificationContext } from "../../services/notifications/notifications";
import { Pager } from "../../components/Pager/Pager";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import styles from "./Connect.module.scss";
import { useApi } from "../../hooks/useApi";
import useContextEnsured from "../../hooks/useContextEnsured";

const PAGE_SIZE = 10;

export const Connect: Component = () => {
  const fetch = useApi();
  const nav = useNavigate();
  const { show } = useContextEnsured(NotificationContext);
  const { id } = useParams();
  const [playlists, setPlaylists] = createSignal<Playlist[]>();
  const [selected, setSelected] = createSignal<string>();
  const [page, setPage] = createSignal(0);

  createEffect(() => {
    fetch((c) => c.remyx(id))
      .catch(() => {
        show("error", "There is no Remyx with the given ID.");
        nav("/");
      })
      .then(() => {
        fetch((c) => c.playlists(PAGE_SIZE, page() * PAGE_SIZE)).then((r) =>
          setPlaylists(r)
        );
      });
  });

  const _connectRemyx = () => {
    const playlist_id = selected();
    if (!playlist_id) return;
    fetch((c) => c.connectRemyx(id, playlist_id)).then(() => {
      show("success", "You are now connected to the Remyx.");
      nav(`/${id}`);
    });
  };

  const _selectPlaylist = (uid: string) => {
    if (selected() === uid) setSelected(undefined);
    else setSelected(uid);
  };

  return (
    <RouteContainer>
      <div class={styles.container}>
        <h2>Connect</h2>
        <section>
          <div class="playlistList">
            <For
              each={playlists()}
              fallback={<>Seems like you have no playlists.</>}
            >
              {(item) => (
                <button
                  class={selected() === item.uid ? "selected" : ""}
                  onClick={() => _selectPlaylist(item.uid)}
                >
                  {item.name}
                </button>
              )}
            </For>
          </div>
          <Pager
            page={page()}
            setPage={setPage}
            hasNext={(playlists()?.length ?? 0) >= PAGE_SIZE}
          />
        </section>
        <button
          class="button"
          disabled={!selected()}
          onClick={_connectRemyx}
          title={
            !selected() ? "Please select a playlist to create a remyx." : ""
          }
        >
          Connect
        </button>
      </div>
    </RouteContainer>
  );
};
