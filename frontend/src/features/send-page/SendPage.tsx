import { useNavigate } from "react-router-dom";
import Button from "src/shared/components/Button";
import { type Template, useTemplates } from "src/shared/hooks/useTemplates";

const SendPage = () => {
  const { templates, isLoading, isError, error } = useTemplates();
  const navigate = useNavigate();

  if (isLoading) return <p>Загрузка...</p>;
  if (isError) return <p>Ошибка: {error?.message}</p>;

  const handleSelect = (tmpl: Template) => {
    navigate(`/send/${tmpl.id}`);
  };

  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-4 text-center">Отправить уведомление</h2>
      <p className="mb-6 text-center">Выберите шаблон уведомления для отправки вашим контактам:</p>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {templates.map((tmpl) => (
          <div
            key={tmpl.id}
            className="
              bg-gray-50
              outline outline-1 outline-gray-200 dark:outline-gray-600
              rounded-lg p-4
              shadow-sm hover:shadow transition-shadow
              flex flex-col justify-between
            "
          >
            <div>
              <div className="text-lg font-medium mb-2">{tmpl.name}</div>
              <div className="text-md text-gray-700 mb-4">{tmpl.body}</div>
            </div>
            <Button onClick={() => handleSelect(tmpl)}>Отправить</Button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default SendPage;
