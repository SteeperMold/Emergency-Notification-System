import React, { useState } from "react";

import Button from "src/shared/components/Button";
import Input from "src/shared/components/Input";
import NavButton from "src/shared/components/NavButton";
import { type Contact, type CreateContactInput, useContacts } from "src/shared/hooks/useContacts";

const ContactsPage = () => {
  const {
    contacts,
    isLoading,
    isError,
    error,
    createContact,
    updateContact,
    deleteContact,
  } = useContacts();

  const [newContact, setNewContact] = useState<CreateContactInput>({ name: "", phone: "" });
  const [editingContact, setEditingContact] = useState<Contact | null>(null);
  const [editedValues, setEditedValues] = useState<Partial<Contact>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleNewChange =
    (field: "name" | "phone") => (e: React.ChangeEvent<HTMLInputElement>) => {
      setNewContact((prev) => ({ ...prev, [field]: e.target.value }));
    };

  const handleCreate = () => {
    if (!newContact.name.trim() || !newContact.phone.trim()) {
      setFormError("Имя и телефон обязательны");
      return;
    }

    setIsSubmitting(true);
    createContact.mutate(newContact, {
      onSuccess: () => {
        setNewContact({ name: "", phone: "" });
        setIsSubmitting(false);
      },
      onError: (err: {response: {status: number, data: string}}) => {
        if (err.response?.status === 422) {
          if (err.response?.data.includes("invalid phone")) {
            setFormError("Неверно набран номер телефона");
          } else if (err.response?.data.includes("invalid name")) {
            setFormError("Название должно быть до 32 символов");
          }
        } else {
          setFormError("Что‑то пошло не так");
        }
        setIsSubmitting(false);
      },
    });
  };

  const handleEditClick = (c: Contact) => {
    setEditingContact(c);
    setEditedValues({ name: c.name, phone: c.phone });
  };

  const handleEditChange =
    (field: "name" | "phone") => (e: React.ChangeEvent<HTMLInputElement>) => {
      setEditedValues((prev) => ({ ...prev, [field]: e.target.value }));
    };

  const handleSaveEdit = () => {
    if (!editingContact) return;
    updateContact.mutate(
      { id: editingContact.id, ...editedValues } as Contact,
      {
        onSuccess: () => {
          setEditingContact(null);
          setEditedValues({});
        },
      },
    );
  };

  const handleCancelEdit = () => {
    setEditingContact(null);
    setEditedValues({});
  };

  const handleDelete = (id: number) => {
    deleteContact.mutate(id);
  };

  if (isLoading) return <p>Загрузка...</p>;
  if (isError) return <p>Ошибка: {error?.message}</p>;

  return (
    <div className="p-4">
      <div className="flex flex-col w-1/4 mb-10">
        <NavButton to="/load-contacts" variant="link">
          Загрузить контакты из .csv/.xlsx ➜
        </NavButton>
      </div>

      <h2 className="text-xl font-bold mb-4">Контакты</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {contacts.map((c) => (
          <div
            key={c.id}
            className="
              bg-gray-50
              outline outline-1 outline-gray-200 dark:outline-gray-600
              rounded-lg p-4
              shadow-sm hover:shadow transition-shadow
              flex flex-col justify-between
            "
          >
            {editingContact?.id === c.id ? (
              <>
                <Input
                  value={editedValues.name || ""}
                  onChange={handleEditChange("name")}
                  className="mb-2"
                  placeholder="Имя"
                />
                <Input
                  value={editedValues.phone || ""}
                  onChange={handleEditChange("phone")}
                  className="mb-2"
                  placeholder="Телефон"
                />
                <div className="flex space-x-2 mt-2">
                  <Button onClick={handleSaveEdit}>Сохранить</Button>
                  <Button variant="outline" onClick={handleCancelEdit}>
                    Отмена
                  </Button>
                </div>
              </>
            ) : (
              <>
                <div className="text-lg font-medium mb-1">{c.name}</div>
                <div className="text-md text-gray-700 mb-4">{c.phone}</div>
                <div className="flex space-x-2">
                  <Button onClick={() => handleEditClick(c)}>Изменить</Button>
                  <Button
                    variant="outline"
                    onClick={() => handleDelete(c.id)}
                  >
                    Удалить
                  </Button>
                </div>
              </>
            )}
          </div>
        ))}

        {/* New contact card */}
        <div
          className="
            bg-gray-50
            outline outline-1 outline-gray-200 dark:outline-gray-600
            rounded-lg p-4
            shadow-sm hover:shadow transition-shadow
            flex flex-col justify-between
          "
        >
          <h3 className="text-lg font-semibold mb-2">Добавить контакт</h3>
          <Input
            placeholder="Имя"
            value={newContact.name}
            onChange={handleNewChange("name")}
            className="mb-2"
          />
          <Input
            placeholder="Телефон"
            value={newContact.phone}
            onChange={handleNewChange("phone")}
            className="mb-4"
          />

          {formError && (
            <p className="text-sm text-red-600 mb-2">{formError}</p>
          )}

          <Button
            onClick={handleCreate}
            disabled={isSubmitting}
            className="mt-auto"
          >
            {isSubmitting ? "Сохранение..." : "Создать"}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ContactsPage;
