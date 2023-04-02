import {
  NotificationContext,
  NotificationState,
  NotificationType,
} from "../../services/notifications/notifications";

import { Component } from "solid-js";
import styles from "./Notifications.module.scss";
import useContextEnsured from "../../hooks/useContextEnsured";

const getBgColor = (t: NotificationType) => {
  switch (t) {
    case "error":
      return "#d13446";
    case "success":
      return "#45b534";
    case "warn":
      return "#d19734";
    default:
      return "#3478d1";
  }
};

export const Notifications: Component = () => {
  const ctx = useContextEnsured(NotificationContext);

  return (
    <div
      class={[styles.container, ctx.state?.visible ? styles.show : null]
        .filter((s) => !!s)
        .join(" ")}
    >
      <div style={{ "background-color": getBgColor(ctx.state.type) }}>
        {ctx.state.content}
      </div>
    </div>
  );
};
