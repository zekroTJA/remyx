import { Component, createEffect, createSignal } from "solid-js";

import { Remyx } from "../../services/api/models";
import styles from "./Details.module.scss";
import { useApi } from "../../hooks/useApi";
import { useParams } from "@solidjs/router";

export const Details: Component = () => {
  const fetch = useApi();
  const [remyx, setRemyx] = createSignal<Remyx>();
  const { id } = useParams();

  createEffect(() => {
    if (!id) return;
    fetch((c) => c.remyx(id)).then((r) => setRemyx(r));
  });

  return (
    <div class={styles.container}>
      <section>
        <label>UID</label>
        <span>{remyx()?.uid}</span>
      </section>
    </div>
  );
};
