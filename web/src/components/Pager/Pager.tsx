import ArrowLeft from "../../assets/allowleft";
import ArrowRight from "../../assets/arrowright";
import { Component } from "solid-js";
import styles from "./Pager.module.scss";

type Props = {
  page: number;
  hasNext: boolean;
  setPage: (v: number) => void;
};

export const Pager: Component<Props> = (props) => {
  const _set = (v: number) => {
    props.setPage(props.page + v);
  };

  return (
    <div class={styles.pager}>
      <button disabled={props.page === 0} onClick={() => _set(-1)}>
        <ArrowLeft />
        <span>Last</span>
      </button>
      <span>Page {props.page + 1}</span>
      <button disabled={!props.hasNext} onClick={() => _set(1)}>
        <span>Next</span>
        <ArrowRight />
      </button>
    </div>
  );
};
