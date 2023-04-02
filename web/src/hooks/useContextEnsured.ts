import { Context, useContext } from "solid-js";

export default function useContextEnsured<T>(context: Context<T>) {
  const ctx = useContext(context);
  return ctx!;
}
