import React, { useState } from "react";

import Button from "src/shared/components/Button";
import Input from "src/shared/components/Input";
import { type CreateTemplateInput, type Template, useTemplates } from "src/shared/hooks/useTemplates";

const TemplatePage = () => {
  const {
    templates,
    isLoading,
    isError,
    error,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  } = useTemplates();

  const [newTemplate, setNewTemplate] = useState<CreateTemplateInput>({ name: "", body: "" });
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [editedValues, setEditedValues] = useState<Partial<Template>>({});
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const handleChange = (field: "name" | "body") => (event: React.ChangeEvent<HTMLInputElement>) => {
    setNewTemplate(prevState => ({ ...prevState, [field]: event.target.value }));
  };

  const handleCreate = () => {
    const { name, body } = newTemplate;
    if (!name?.trim() || !body?.trim()) return;

    setIsSubmitting(true);
    createTemplate.mutate(newTemplate, {
      onSuccess: () => setNewTemplate({ name: "", body: "" }),
    });
  };

  const handleEditClick = (tmpl: Template) => {
    setEditingTemplate(tmpl);
    setEditedValues({ name: tmpl.name, body: tmpl.body });
  };

  const handleEditChange = (field: "name" | "body") => (event: React.ChangeEvent<HTMLInputElement>) => {
    setEditedValues(prev => ({ ...prev, [field]: event.target.value }));
  };

  const handleSaveEdit = () => {
    if (!editingTemplate) {
      return;
    }

    updateTemplate.mutate(
      { ...editingTemplate, ...editedValues },
      {
        onSuccess: () => {
          setEditingTemplate(null);
          setEditedValues({});
        },
      },
    );
  };

  const handleCancelEdit = () => {
    setEditingTemplate(null);
    setEditedValues({});
  };

  const handleDelete = (id: number) => {
    deleteTemplate.mutate(id);
  };

  if (isLoading) return <p>Загрузка...</p>;
  if (isError) return <p>Ошибка: {error?.message}</p>;

  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-4">Шаблоны</h2>

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
            {editingTemplate?.id === tmpl.id ? <>
              <Input
                value={editedValues.name || ""}
                onChange={handleEditChange("name")}
                className="mb-2"
              />
              <Input
                value={editedValues.body || ""}
                onChange={handleEditChange("body")}
                className="mb-2"
              />
              <div className="flex space-x-2 mt-2">
                <Button onClick={handleSaveEdit}>Сохранить</Button>
                <Button variant="outline" onClick={handleCancelEdit}>
                  Отмена
                </Button>
              </div>
            </> : <>
              <div className="text-lg font-medium mb-2">{tmpl.name}</div>
              <div className="text-md text-gray-700 mb-4">
                {tmpl.body}
              </div>
              <div className="flex space-x-2">
                <Button onClick={() => handleEditClick(tmpl)}>
                  Изменить
                </Button>
                <Button variant="outline" onClick={() => handleDelete(tmpl.id)}>
                  Удалить
                </Button>
              </div>
            </>}
          </div>
        ))}

        <div
          className="
            bg-gray-50
            outline outline-1 outline-gray-200 dark:outline-gray-600
            rounded-lg p-4
            shadow-sm hover:shadow transition-shadow
            flex flex-col justify-between
          "
        >
          <h3 className="text-lg font-semibold mb-2">Добавить шаблон</h3>
          <Input
            placeholder="Название шаблона"
            value={newTemplate.name}
            onChange={handleChange("name")}
            className="mb-2"
          />
          <Input
            placeholder="Текст шаблона"
            value={newTemplate.body}
            onChange={handleChange("body")}
            className="mb-4"
          />
          <Button
            onClick={handleCreate}
            disabled={isSubmitting}
            className="mt-auto"
          >
            {isSubmitting ? "Создание..." : "Создать"}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default TemplatePage;
