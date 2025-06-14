import React from "react";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  className?: string;
}

const baseStyles = "w-full outline-none bg-inherit border-b-2 py-2 dark:placeholder:text-dark-text-primary";

const Input: React.FC<InputProps> = ({ className, ...props }) => {
  return (
    <input
      className={`${baseStyles} ${className}`}
      {...props}
    />
  );
};

export default Input;
