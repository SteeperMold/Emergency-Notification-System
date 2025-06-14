import NavButton from "src/shared/components/NavButton";
import { useUser } from "src/shared/hooks/useUser";

import NotAuthenticated from "./NotAuthenticated";

const MainPage = () => {
  const { user } = useUser();

  if (!user) {
    return <NotAuthenticated/>;
  }

  return (
    <div className="flex flex-col items-center">
      <h1 className="text-xl">
        Для того чтобы начать, вы должны
      </h1>

      <div className="flex flex-row justify-between items-center 2xl:w-1/2 lg:w-2/3 md:w-5/6 sm:w-11/12 mt-20">
        <NavButton
          to="/template"
          variant="secondary"
        >
          Создать шаблон сообщения
        </NavButton>
        <p className="text-2xl">и</p>
        <NavButton
          to="/contacts"
          variant="secondary"
        >
          Загрузить список контактов
        </NavButton>
      </div>
    </div>
  );
};

export default MainPage;
