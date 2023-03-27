import { Navigate, Route, Routes } from "@solidjs/router";

import type { Component } from "solid-js";
import { Connect } from "./routes/Connect/Connect";
import { Create } from "./routes/Create/Create";
import { Details } from "./routes/Details/Details";
import { Home } from "./routes/Home/Home";
import { Login } from "./routes/Login/Login";
import styles from "./App.module.scss";

const App: Component = () => {
  return (
    <div class={styles.App}>
      <Routes>
        <Route path="/login" component={Login} />
        <Route path="/create" component={Create} />
        <Route path="/c/:id" component={Connect} />
        <Route path="/:id" component={Details} />
        <Route path="/" component={Home} />
        <Route path="*" element={<Navigate href="/" />} />
      </Routes>
    </div>
  );
};

export default App;
