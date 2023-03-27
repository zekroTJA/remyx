import { Component, For, createEffect, createSignal } from "solid-js";

import { NavLink } from "@solidjs/router";
import { Remyx } from "../../services/api/models";
import styles from "./Home.module.scss";
import { useApi } from "../../hooks/useApi";

export const Home: Component = () => {
  const fetch = useApi();
  const [remyxes, setRemyxes] = createSignal<Remyx[]>();

  createEffect(() => {
    fetch((c) => c.remyxes()).then((r) => setRemyxes(r));
  });

  return (
    <div class={styles.list}>
      <For each={remyxes()} fallback={<>No items</>}>
        {(item) => <div>{item.uid}</div>}
      </For>
      <NavLink href="/create">Create a new Remyx</NavLink>
    </div>
  );
};
