import { Component, For, createEffect, createSignal } from "solid-js";
import { NavLink, useNavigate } from "@solidjs/router";

import ArrowLeft from "../../assets/allowleft";
import { NotificationContext } from "../../services/notifications/notifications";
import { Pager } from "../../components/Pager/Pager";
import { Playlist } from "../../services/api/models";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import styles from "./Create.module.scss";
import { useApi } from "../../hooks/useApi";
import useContextEnsured from "../../hooks/useContextEnsured";

const PAGE_SIZE = 10;

export const Create: Component = () => {
  const fetch = useApi();
  const nav = useNavigate();
  const { show } = useContextEnsured(NotificationContext);
  const [playlists, setPlaylists] = createSignal<Playlist[]>();
  const [selected, setSelected] = createSignal<string>();
  const [head, setHead] = createSignal<number>(20);
  const [name, setName] = createSignal<string>();
  const [page, setPage] = createSignal(0);

  createEffect(() => {
    fetch((c) => c.playlists(PAGE_SIZE, page() * PAGE_SIZE)).then((r) =>
      setPlaylists(r)
    );
  });

  const _createRemyx = () => {
    const uid = selected();
    if (!uid) return;
    fetch((c) => c.createRemyx(uid, head(), name())).then((r) => {
      show("success", "The Remyx has been created.");
      nav(`/${r.uid}`);
    });
  };

  const _selectPlaylist = (uid: string) => {
    if (selected() === uid) setSelected(undefined);
    else setSelected(uid);
  };

  const _setName = (v: string) => {
    if (!v) setName(undefined);
    else setName(v);
  };

  return (
    <RouteContainer>
      <div class="flex gap">
        <NavLink href="-1" class="navButton">
          <ArrowLeft height="2em" width="2em" />
        </NavLink>
        <h2>Create a Remyx</h2>
      </div>
      <div class={styles.container}>
        <section>
          <label for="iName">
            <h4>Give your Remyx a catchy name!</h4>
          </label>
          <input
            id="iName"
            value={name() ?? ""}
            onInput={(e) => _setName(e.currentTarget.value)}
          />
        </section>
        <section>
          <label for="iHead">
            <h4>How many songs should be used from the source playlists?</h4>
          </label>
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
          <label>
            <h4>Which playlist should be used as source?</h4>
          </label>
          <div class="playlistList">
            <For each={playlists()}>
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
          onClick={_createRemyx}
          title={
            !selected() ? "Please select a playlist to create a remyx." : ""
          }
        >
          Create Remyx
        </button>
      </div>
    </RouteContainer>
  );
};
