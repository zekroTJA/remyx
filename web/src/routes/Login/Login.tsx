import { Component } from "solid-js";
import Spotify from "../../assets/spotify";
import { loginUrl } from "../../services/api/client";
import styles from "./Login.module.scss";

export const Login: Component = () => {
  return (
    <div class={styles.container}>
      <div class={styles.login}>
        <h1>REMYX</h1>
        <a class="button" href={loginUrl()}>
          <Spotify />
          Login with Spotify
        </a>
      </div>
    </div>
  );
};
