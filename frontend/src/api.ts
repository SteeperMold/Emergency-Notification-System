import axios from "axios";

const Api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080",
  headers: {
    "Content-Type": "application/json",
  },
});

Api.interceptors.request.use(
  config => {
    const token = localStorage.getItem("accessToken");

    if (token) {
      config.headers["Authorization"] = `Bearer ${token}`;
    }

    return config;
  },
  error => Promise.reject(error),
);

Api.interceptors.response.use(
  response => response,
  async err => {
    const originalConfig = err.config;

    if (
      err.response &&
      err.response.status == 401 &&
      !originalConfig._retry &&
      !["/login", "/signup", "/refresh-token"].includes(originalConfig.url)
    ) {
      if (err.response.status === 401 && !originalConfig._retry) {
        originalConfig._retry = true;

        try {
          const response = await Api.post("/refresh-token", {
            refreshToken: localStorage.getItem("refreshToken"),
          });

          const { accessToken } = response.data;
          localStorage.setItem("accessToken", accessToken);

          return Api(originalConfig);
        } catch (_error) {
          localStorage.removeItem("accessToken");
          localStorage.removeItem("refreshToken");

          return Promise.reject(_error);
        }
      }
    }

    return Promise.reject(err);
  },
);

export default Api;
