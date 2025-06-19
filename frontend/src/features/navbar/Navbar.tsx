import { Link } from "react-router-dom";
import NavButton from "src/shared/components/NavButton";
import { useUser } from "src/shared/hooks/useUser";

import ProfileButton from "./ProfileButton";

const Navbar = () => {
  const { user } = useUser();

  return (
    <nav className="flex justify-between items-center py-[2.5vh]">
      <div>
        <Link to="/" className="text-[3vh] mx-10">Главная</Link>
        <NavButton to="/template" className="text-lg mr-6">Шаблоны сообщений</NavButton>
        <NavButton to="/contacts" className="text-lg">Контакты</NavButton>
      </div>

      <div className="mr-2 flex items-center">
        {user ? (
          <ProfileButton className="text-[3vh] mr-10"/>
        ) : (
          <div className="mr-6">
            <NavButton to="/login" className="mr-6">Войти в аккаунт</NavButton>
            <NavButton to="/signup">Зарегистрироваться</NavButton>
          </div>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
