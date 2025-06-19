import { Navigate } from "react-router-dom";
import NavButton from "src/shared/components/NavButton";
import { useContacts } from "src/shared/hooks/useContacts";
import { useTemplates } from "src/shared/hooks/useTemplates";
import { useUser } from "src/shared/hooks/useUser";

import NotAuthenticated from "./NotAuthenticated";

const MainPage = () => {
  const { user } = useUser();

  const {
    contacts,
    isLoading: contactsIsLoading,
    isError: contactsIsError,
  } = useContacts();

  const {
    templates,
    isLoading: templatesIsLoading,
    isError: templatesIsError,
  } = useTemplates();

  if (!user) {
    return <NotAuthenticated/>;
  }

  if (contactsIsLoading || templatesIsLoading) {
    return <p>Загрузка...</p>;
  }

  if (
    contacts.length !== 0 &&
    templates.length !== 0 &&
    !contactsIsError &&
    !templatesIsError
  ) {
    return <Navigate to="/send"/>;
  }

  return (
    <div className="flex flex-col items-center">
      <h1 className="text-xl">
        Для того чтобы начать, вы должны
      </h1>

      <div className="flex flex-row justify-between items-center 2xl:w-1/2 lg:w-2/3 md:w-5/6 sm:w-11/12 mt-20">
        <NavButton
          to="/template"
          variant={(templates.length === 0 || templatesIsError) ? "secondary" : "link"}
        >
          Создать шаблон сообщения
        </NavButton>

        <p className="text-2xl">и</p>

        <NavButton
          to="/contacts"
          variant={(contacts.length === 0 || contactsIsError) ? "secondary" : "link"}
        >
          Загрузить список контактов
        </NavButton>
      </div>
    </div>
  );
};

export default MainPage;
