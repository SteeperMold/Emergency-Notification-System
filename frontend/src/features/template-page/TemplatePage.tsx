import React, { useState } from "react";

import { StatusCodes } from "http-status-codes";
import { Navigate } from "react-router-dom";
import Api from "src/api";
import Button from "src/shared/components/Button";
import Input from "src/shared/components/Input";
import { useUser } from "src/shared/hooks/useUser";

const TemplatePage = () => {
  const [error, setError] = useState<string | null>(null);
  const { user } = useUser();

  if (!user) {
    return <Navigate to="/"/>;
  }

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const body = formData.get("body");

    Api.post("/template", { body })
      .catch(error => {
        if (error.status === StatusCodes.UNPROCESSABLE_ENTITY) {
          setError("Указанный шаблон слишком короткий или слишком длинный");
          return;
        }
        setError("Не удалось сохранить шаблон");
      });
  };

  return <>
    <h1 className="text-2xl text-center">Редактирование шаблона</h1>

    <form onSubmit={onSubmit} className="mx-auto w-1/4 mt-10 flex flex-col items-center">
      {error && <h2 className="self-start mb-8 text-xl text-red-600">{error}</h2>}

      <Input
        required={true}
        name="body"
        placeholder="Введие шаблон"
      />

      <Button
        className="mt-10"
      >
        Сохранить
      </Button>
    </form>
  </>;
};

export default TemplatePage;
