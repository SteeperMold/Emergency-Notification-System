import { Link, type LinkProps } from "react-router-dom";

const baseLinkClass = "rounded-md font-medium transition-colors " +
  "focus:outline-none focus:ring-2 focus:ring-offset-2 cursor-pointer text-center";

interface NavLinkStyledProps extends LinkProps {
  className?: string;
  variant?: "link" | "secondary";
}

const variants = {
  link: "px-2 py-1 text-gray-600 hover:text-gray-800 hover:bg-gray-100",
  secondary: "px-4 py-2 bg-black text-white hover:bg-gray-800 hover:text-gray-200 focus:ring-gray-400 text-xl",
};

const NavButton = ({
  variant = "link",
  className,
  ...props
}: NavLinkStyledProps) => {
  return (
    <Link
      className={`${baseLinkClass} ${variants[variant]} ${className}`}
      {...props}
    />
  );
};

export default NavButton;
