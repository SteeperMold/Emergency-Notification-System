import React from "react";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline";
}

const baseStyles = "rounded-md font-medium text-xl " +
  "transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 cursor-pointer";

const variants = {
  primary: "px-7 py-2 bg-[#0070d7] text-white hover:bg-[#005bb5] focus:ring-blue-500",
  secondary: "px-4 py-2 bg-black text-white hover:bg-gray-800 hover:text-gray-200 focus:ring-gray-400",
  outline:
    "px-3 py-1 border border-gray-400 text-gray-700 bg-gray-100 hover:bg-gray-50 focus:ring-gray-400",
};

const Button: React.FC<ButtonProps> = ({
  variant = "primary",
  className,
  children,
  ...props
}) => {
  return (
    <button
      className={`${baseStyles} ${variants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
};

export default Button;
