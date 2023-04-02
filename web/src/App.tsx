import { Navigate, Route, Routes } from "@solidjs/router";

import type { Component } from "solid-js";
import { Connect } from "./routes/Connect/Connect";
import { Create } from "./routes/Create/Create";
import { Details } from "./routes/Details/Details";
import { Home } from "./routes/Home/Home";
import { Login } from "./routes/Login/Login";
import { NotificationProvider } from "./services/notifications/notifications";
import { Notifications } from "./components/Notifications/Notifications";
import styles from "./App.module.scss";

const App: Component = () => {
  return (
    <div class={styles.App}>
      <NotificationProvider>
        <Notifications />
        <Routes>
          <Route path="/login" component={Login} />
          <Route path="/create" component={Create} />
          <Route path="/c/:id" component={Connect} />
          <Route path="/:id" component={Details} />
          <Route path="/" component={Home} />
          <Route path="*" element={<Navigate href="/" />} />
        </Routes>
      </NotificationProvider>
    </div>
  );
};

export default App;
