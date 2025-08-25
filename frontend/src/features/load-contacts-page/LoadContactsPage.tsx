import React, { useRef, useState } from "react";

import { useQueryClient } from "@tanstack/react-query";
import { HttpStatusCode } from "axios";
import { useNavigate } from "react-router-dom";
import Api from "src/api";
import Button from "src/shared/components/Button";
import NavButton from "src/shared/components/NavButton";

import tableExampleUrl from "./table_example.png";

const LoadContactsPage = () => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();
  const qc = useQueryClient();

  const onButtonClick = () => {
    setError(null);
    setSuccess(false);
    inputRef.current?.click();
  };

  const onFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    setError(null);
    setSuccess(false);

    const form = new FormData();
    form.append("file", file);

    Api.post("/load-contacts", form, { headers: { "Content-Type": "multipart/form-data" } })
      .then(async res => {
        if (res.status === HttpStatusCode.Accepted) {
          await qc.invalidateQueries({ queryKey: ["contacts"] });
          navigate("/");
        } else {
          setError("Не удалось загрузить файл");
          if (inputRef.current) inputRef.current.value = "";
        }
      })
      .catch(() => {
        setError("Не удалось загрузить файл");
        if (inputRef.current) inputRef.current.value = "";
      });
  };

  return (
    <div className="flex flex-col items-center">
      <h1 className="text-2xl">Загрузка контактов из таблицы</h1>

      <p className="my-10">
        Для успешной загрузки контактов, таблица должна быть
        отформатирована в таком формате:
      </p>

      <img
        src={tableExampleUrl}
        alt="Не удалось загрузить изображение"
        className="mb-10"
      />

      <p className="mb-10">
        Все строчки с корректным содержимым будут добавлены в Ваш
        список контактов, а остальные будут пропущены
      </p>

      <Button
        onClick={onButtonClick}
        disabled={uploading}
      >
        {uploading ? "Загрузка..." : "Загрузить csv/xlsx"}
      </Button>

      {error && <p className="text-red-600 mt-2">Ошибка: {error}</p>}
      {success && <>
        <p className="text-green-600 mt-2">Файл загружен успешно!</p>
        <NavButton to="/contacts">Перейти к контактам</NavButton>
      </>}

      <input
        ref={inputRef}
        onChange={onFileChange}
        className="collapse"
        type="file"
        accept="
          .csv,
          application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,
          application/vnd.ms-excel
        "
      />
    </div>
  );
};

export default LoadContactsPage;
