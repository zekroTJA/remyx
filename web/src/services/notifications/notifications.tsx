import { Component, JSX, ParentProps, createContext } from "solid-js";

import { createStore } from "solid-js/store";

export type NotificationType = "info" | "error" | "success" | "warn";

export type NotificationState = {
  type: NotificationType;
  content: string | JSX.Element;
  visible: boolean;
};

export type NotificationContextProps = {
  state: NotificationState;
  show: (
    type: NotificationType,
    content: string | JSX.Element,
    duration?: number
  ) => void;
  hide: () => void;
};

export const NotificationContext = createContext<NotificationContextProps>();

export const NotificationProvider: Component<ParentProps> = (props) => {
  const [state, setState] = createStore<NotificationState>({
    type: "info",
    content: "",
    visible: false,
  });

  const show = (
    type: NotificationType,
    content: string | JSX.Element,
    duration?: number
  ) => {
    setState({ type, content, visible: true });
    setTimeout(() => hide(), duration ?? 6000);
  };

  const hide = () => {
    setState({ visible: false });
  };

  const notifications = {
    state,
    show,
    hide,
  };

  return (
    <NotificationContext.Provider value={notifications}>
      {props.children}
    </NotificationContext.Provider>
  );
};
