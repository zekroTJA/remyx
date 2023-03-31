import { Component, For, createEffect, createSignal } from "solid-js";
import { MyRemyxes, Remyx } from "../../services/api/models";

import Add from "../../assets/add";
import { NavLink } from "@solidjs/router";
import { RouteContainer } from "../../components/RouteContainer/RouteContainer";
import styles from "./Home.module.scss";
import { useApi } from "../../hooks/useApi";

export const Home: Component = () => {
  const fetch = useApi();
  const [remyxes, setRemyxes] = createSignal<MyRemyxes>();

  createEffect(() => {
    fetch((c) => c.remyxes()).then((r) => setRemyxes(r));
  });

  return (
    <RouteContainer>
      <h2>Your Remyxes</h2>
      <div class={styles.list}>
        <For each={remyxes()?.created} fallback={<>No items</>}>
          {(item) => <ListItem item={item} />}
        </For>
      </div>

      <h2>Participating Remyxes</h2>
      <div class={styles.list}>
        <For each={remyxes()?.connected} fallback={<>No items</>}>
          {(item) => <ListItem item={item} />}
        </For>
      </div>

      <NavLink class={"button " + styles.createButton} href="/create">
        <Add />
        Create a new Remyx
      </NavLink>
    </RouteContainer>
  );
};

const ListItem: Component<{ item: Remyx }> = ({ item }) => {
  return (
    <NavLink href={`/${item.uid}`}>
      {item.name ? (
        <span>{item.name}</span>
      ) : (
        <span class={styles.idPlaceholder}>{item.uid}</span>
      )}
    </NavLink>
  );
};
