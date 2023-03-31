import { Component, JSX, ParentProps } from "solid-js";

import styles from "./RouteContainer.module.scss";

type Props = ParentProps & JSX.HTMLAttributes<HTMLDivElement>;

export const RouteContainer: Component<Props> = ({ children }) => {
  return (
    <div class={styles.container}>
      <div>{children}</div>
    </div>
  );
};
