import { Component } from "solid-js";
import { loginUrl } from "../../services/api/client";

export const Login: Component = () => {
  return (
    <div>
      <a href={loginUrl()}>Login with Spotify</a>
    </div>
  );
};
