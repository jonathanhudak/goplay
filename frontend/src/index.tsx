import * as React from "react";
import * as ReactDOM from "react-dom";
import { App } from "./App";

const root: HTMLElement = document.createElement("div");

document.body.appendChild(root);

ReactDOM.render(<App />, root);
