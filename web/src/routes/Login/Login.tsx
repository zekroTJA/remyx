import { Component } from "solid-js";
import Spotify from "../../assets/spotify";
import Warning from "../../assets/warning";
import { loginUrl } from "../../services/api/client";
import styles from "./Login.module.scss";

export const Login: Component = () => {
  return (
    <div class={styles.container}>
      <div class={styles.login}>
        <h1>REMYX</h1>
        <div class="mb2 card warning" style={{ "margin-bottom": "1.8em" }}>
          <div class="heading">
            <Warning />
            <strong>Warning</strong>
          </div>
          Because the Spotify API application is currently in the process of
          being reviewed, the login to the app is invite-only at the moment.
          Feel free to <a href="https://www.zekro.de/contact">contact</a> me so
          I can add you to the testers list.
        </div>
        <a class="button" href={loginUrl()}>
          <Spotify />
          Login with Spotify
        </a>
      </div>
    </div>
  );
};
