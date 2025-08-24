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

interface TemplatesResponse {
  templates: Template[];
  total: number;
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
    queryKey: ["templates"],
    queryFn: () => Api.get<TemplatesResponse>("/templates").then(res => res.data.templates),
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
      Api.post("/templates", {
        name: newTmpl.name,
        body: newTmpl.body,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });

  const updateTemplate = useMutation({
    mutationFn: (tmpl: UpdateTemplateInput) =>
      Api.put(`/templates/${tmpl.id}`, {
        name: tmpl.name,
        body: tmpl.body,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });

  const deleteTemplate = useMutation({
    mutationFn: (id: number) => Api.delete(`/templates/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
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
