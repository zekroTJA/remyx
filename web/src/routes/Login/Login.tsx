import { Component } from "solid-js";
import Logo from "../../assets/remyx-portrait.svg";
import Spotify from "../../assets/spotify";
import { loginUrl } from "../../services/api/client";
import styles from "./Login.module.scss";
import { useLocation } from "@solidjs/router";

export const Login: Component = () => {
  const { state } = useLocation();

  return (
    <div class={styles.container}>
      <div class={styles.login}>
        <img src={Logo} />
        <a class="button" href={loginUrl((state as any)?.redirect)}>
          <Spotify />
          Login with Spotify
        </a>
      </div>
    </div>
  );
};
