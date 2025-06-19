import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Api from "src/api";

export interface Contact {
  id: number;
  userId: number;
  name: string;
  phone: string;
  creationTime: Date;
  updateTime: Date;
}

export type CreateContactInput = Pick<Contact, "name" | "phone">;
export type UpdateContactInput = CreateContactInput & { id: number };

export const useContacts = () => {
  const qc = useQueryClient();

  const {
    data: contacts = [],
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["contacts"],
    queryFn: () => Api.get<Contact[]>("/contacts").then(res => res.data),
    select: rawContacts => {
      return rawContacts
        .map(c => ({
          ...c,
          creationTime: new Date(c.creationTime as unknown as string),
          updateTime: new Date(c.updateTime as unknown as string),
        }))
        .sort((a, b) =>
          a.creationTime.getTime() - b.creationTime.getTime(),
        );
    },
  });

  const createContact = useMutation({
    mutationFn: (newContact: CreateContactInput) =>
      Api.post("/contacts", {
        name: newContact.name,
        phone: newContact.phone,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["contacts"] }),
  });

  const updateContact = useMutation({
    mutationFn: (contact: UpdateContactInput) =>
      Api.put(`/contacts/${contact.id}`, {
        name: contact.name,
        phone: contact.phone,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["contacts"] }),
  });

  const deleteContact = useMutation({
    mutationFn: (id: number) => Api.delete(`/contacts/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["contacts"] }),
  });

  return {
    contacts,
    isLoading,
    isError,
    error,
    createContact,
    updateContact,
    deleteContact,
  };
};
