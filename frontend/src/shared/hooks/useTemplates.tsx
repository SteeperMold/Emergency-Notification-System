import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Api from "src/api";

export interface Template {
  id: number;
  userId: number;
  name: string;
  body: string;
  creationTime: Date;
  updateTime: Date;
}

export type CreateTemplateInput = Pick<Template, "name" | "body">;
export type UpdateTemplateInput = CreateTemplateInput & { id: number };

export const useTemplates = () => {
  const qc = useQueryClient();

  const {
    data: templates = [],
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["template"],
    queryFn: () => Api.get<Template[]>("/template").then(res => res.data),
    select: rawTemplates => {
      return rawTemplates
        .map(t => ({
          ...t,
          creationTime: new Date(t.creationTime as unknown as string),
          updateTime: new Date(t.creationTime as unknown as string),
        }))
        .sort((a, b) =>
          a.creationTime.getTime() - b.creationTime.getTime(),
        );
    },
  });

  const createTemplate = useMutation({
    mutationFn: (newTmpl: CreateTemplateInput) =>
      Api.post("/template", {
        name: newTmpl.name,
        body: newTmpl.body,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["template"] }),
  });

  const updateTemplate = useMutation({
    mutationFn: (tmpl: UpdateTemplateInput) =>
      Api.put(`/template/${tmpl.id}`, {
        name: tmpl.name,
        body: tmpl.body,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["template"] }),
  });

  const deleteTemplate = useMutation({
    mutationFn: (id: number) => Api.delete(`/template/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["template"] }),
  });

  return {
    templates,
    isLoading,
    isError,
    error,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  };
};
