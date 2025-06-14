import React, { useState } from "react";

import { StatusCodes } from "http-status-codes";
import { Navigate, useNavigate } from "react-router-dom";
import Api from "src/api";
import Button from "src/shared/components/Button";
import Input from "src/shared/components/Input";
import { useUser } from "src/shared/hooks/useUser";

const LoginPage = () => {
  const [error, setError] = useState<string | null>(null);
  const { user, updateUser } = useUser();
  const navigate = useNavigate();

  if (user) {
    return <Navigate to="/"/>;
  }

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const email = formData.get("email");
    const password = formData.get("password");

    Api.post("/login", { email, password })
      .then(response => {
        localStorage.setItem("accessToken", response.data.accessToken);
        localStorage.setItem("refreshToken", response.data.refreshToken);
        updateUser();
        navigate("/");
      })
      .catch(error => {
        if (error.status === StatusCodes.NOT_FOUND) {
          setError("Пользователь с указанным адресом электронной почты не существует");
          return;
        }
        if (error.status === StatusCodes.UNAUTHORIZED) {
          setError("Неверный пароль или адрес электронной почты");
          return;
        }
        setError("Не удалось войти в аккаунт");
      });
  };

  return <>
    <h1 className="text-3xl text-center mt-10">Вход в аккаунт</h1>

    <form onSubmit={onSubmit} className="mx-auto w-1/4 mt-16 flex flex-col items-center">
      {error && <h2 className="self-start mb-8 text-xl text-red-600">{error}</h2>}

      <Input
        type="email"
        autoComplete="email"
        required={true}
        name="email"
        placeholder="Email"
      />

      <Input
        type="password"
        autoComplete="current-password"
        required={true}
        name="password"
        placeholder="Пароль"
        className="my-10"
      />

      <Button
        type="submit"
        variant="primary"
      >
        Войти в аккаунт →
      </Button>
    </form>
  </>;
};

export default LoginPage;
