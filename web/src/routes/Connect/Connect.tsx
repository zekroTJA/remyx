import { Component, For, createEffect, createSignal } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";

import { NotificationContext } from "../../services/notifications/notifications";
import { Pager } from "../../components/Pager/Pager";
import { Playlist } from "../../services/api/models";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import { isLibraryPlaylist } from "../../util";
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

  const _selectPlaylist = (e: MouseEvent, pl: Playlist) => {
    if (e.shiftKey && !isLibraryPlaylist(pl.uid)) {
      window.open(pl.url);
      return;
    }
    if (selected() === pl.uid) setSelected(undefined);
    else setSelected(pl.uid);
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
                  onClick={(e) => _selectPlaylist(e, item)}
                >
                  <span>{item.name}</span>
                  <span class="controls">
                    {!isLibraryPlaylist(item.uid) && (
                      <span class="ois">Shift + Click to open in Spotify</span>
                    )}
                  </span>
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
