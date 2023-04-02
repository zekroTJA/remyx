import { Component, For, createEffect, createSignal } from "solid-js";
import { MyRemyxes, Remyx } from "../../services/api/models";

import Add from "../../assets/add";
import { NavLink } from "@solidjs/router";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import Trashcan from "../../assets/trashcan";
import styles from "./Home.module.scss";
import { useApi } from "../../hooks/useApi";

export const Home: Component = () => {
  const fetch = useApi();
  const [remyxes, setRemyxes] = createSignal<MyRemyxes>();

  createEffect(() => {
    fetch((c) => c.remyxes()).then((r) => setRemyxes(r));
  });

  const _deleteRemyx = (e: MouseEvent, id: string) => {
    e.preventDefault();
    const rmx = remyxes();
    if (!rmx) return;
    fetch((c) => c.deleteRemyx(id)).then(() =>
      setRemyxes({
        connected: rmx.connected.filter((r) => r.uid != id),
        created: rmx.created.filter((r) => r.uid !== id),
      })
    );
  };

  return (
    <RouteContainer>
      <h2>Your Remyxes</h2>
      <div class="playlistList">
        <For each={remyxes()?.created} fallback={<>No items</>}>
          {(item) => <ListItem item={item} onDelete={_deleteRemyx} />}
        </For>
      </div>

      <h2>Participating Remyxes</h2>
      <div class="playlistList">
        <For each={remyxes()?.connected} fallback={<>No items</>}>
          {(item) => <ListItem item={item} onDelete={_deleteRemyx} />}
        </For>
      </div>

      <NavLink class={"button " + styles.createButton} href="/create">
        <Add />
        Create a new Remyx
      </NavLink>
    </RouteContainer>
  );
};

const ListItem: Component<{
  item: Remyx;
  onDelete: (e: MouseEvent, id: string) => void;
}> = ({ item, onDelete }) => {
  return (
    <NavLink href={`/${item.uid}`}>
      {item.name ? (
        <span>{item.name}</span>
      ) : (
        <span class={styles.idPlaceholder}>{item.uid}</span>
      )}
      <div class="controls">
        <button onClick={(e) => onDelete(e, item.uid)}>
          <Trashcan />
        </button>
      </div>
    </NavLink>
  );
};
