import NavButton from "src/shared/components/NavButton";

const NotAuthenticated = () => {
  return (
    <div className="flex flex-col items-center">
      <h1 className="text-xl">
        Чтобы начать, необходимо
      </h1>

      <div className="flex flex-row justify-between items-center w-1/2 mt-20">
        <NavButton
          to="/login"
          variant="secondary"
          className="w-1/3"
        >
          Войти в аккаунт
        </NavButton>
        <p className="text-2xl">или</p>
        <NavButton
          to="/signup"
          variant="secondary"
          className="w-1/3"
        >
          Зарегистрироваться
        </NavButton>
      </div>
    </div>
  );
};

export default NotAuthenticated;
