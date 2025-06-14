import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import App from "./App";
import { UserProvider } from "./shared/hooks/useUser";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Root element with ID 'root' not found");
}

const root = ReactDOM.createRoot(rootElement);

root.render(
  <UserProvider>
    <BrowserRouter>
      <App/>
    </BrowserRouter>
  </UserProvider>,
);
