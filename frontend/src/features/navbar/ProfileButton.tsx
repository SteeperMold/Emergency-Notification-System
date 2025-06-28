import { useState } from "react";

import NavButton from "src/shared/components/NavButton";
import { useUser } from "src/shared/hooks/useUser";

import ProfileIcon from "./profile_icon.svg?react";

interface ProfileButtonProps {
  className: string;
}

const ProfileButton = ({ className }: ProfileButtonProps) => {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const { user } = useUser();

  if (!user) {
    return <></>;
  }

  return (
    <div className={`${className} flex flex-row justify-end`}>
      <div className="flex flex-row justify-end items-center w-1/2 hover:bg-gray-100 rounded-xl">
        <ProfileIcon className="w-full h-1/2 bg-gray-100 rounded-xl cursor-pointer"/>
        <button
          className="text-lg cursor-pointer"
          onClick={() => setIsDropdownOpen(!isDropdownOpen)}
        >
          {user.email}
        </button>
      </div>

      {isDropdownOpen && (
        <div
          onClick={() => setIsDropdownOpen(false)}
          className="absolute z-10 top-20 flex flex-col items-center p-2"
        >
          <NavButton to="/logout" className="text-lg">Выйти из аккаунта</NavButton>
        </div>
      )}
    </div>
  );
};

export default ProfileButton;
