import React, { useEffect, useState } from "react";

import { useNavigate, useParams } from "react-router-dom";
import Api from "src/api";
import Button from "src/shared/components/Button";

const SendCooldownPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();

  const [secondsLeft, setSecondsLeft] = useState(5);
  const [cancelled, setCancelled] = useState(false);
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (cancelled || secondsLeft <= 0) return;

    const timer = setTimeout(() => {
      setSecondsLeft(prev => prev - 1);
    }, 1000);

    return () => clearTimeout(timer);
  }, [secondsLeft, cancelled]);

  useEffect(() => {
    if (cancelled || secondsLeft !== 0) return;

    setSending(true);
    Api.post(`/send-notification/${id}`)
      .then(() => setSent(true))
      .catch(() => setError("Не удалось отправить уведомление"))
      .finally(() => setSending(false));
  }, [secondsLeft, cancelled, id]);

  const handleCancel = () => {
    setCancelled(true);
    navigate(-1);
  };

  return (
    <div className="p-4 flex flex-col items-center">
      <h2 className="text-xl font-bold mb-4">Отправка уведомления</h2>

      {error && <p className="text-red-600 mb-4">Ошибка: {error}</p>}

      {!cancelled && !sent ? (
        <>
          <p className="mb-4">
            Уведомление будет отправлено через{" "}
            <span className="font-semibold">{secondsLeft}</span>{" "}
            {secondsLeft === 1 ? "секунду" : "секунд"}.
          </p>
          <Button variant="outline" onClick={handleCancel} disabled={sending}>
            {sending ? "Отправка..." : "Отменить отправку"}
          </Button>
        </>
      ) : null}

      {sent && <p className="text-green-600 mb-4">Уведомление успешно отправлено!</p>}
      {sent && (
        <Button onClick={() => navigate("/contacts")}>Вернуться к контактам</Button>
      )}

      {cancelled && <p className="text-red-600">Отправка отменена.</p>}
    </div>
  );
};

export default SendCooldownPage;
