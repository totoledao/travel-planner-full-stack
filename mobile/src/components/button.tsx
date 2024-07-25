import clsx from "clsx";
import { createContext, useContext } from "react";
import {
  ActivityIndicator,
  Pressable,
  PressableProps,
  Text,
} from "react-native";
import { TextProps } from "react-native-svg";

type Variants = "primary" | "secondary";
type ButtonProps = PressableProps & {
  variant?: Variants;
  isLoading?: boolean;
};

const ThemeContext = createContext<{ variant?: Variants }>({});

function Button({
  variant = "primary",
  isLoading = false,
  children,
  ...props
}: ButtonProps) {
  return (
    <ThemeContext.Provider value={{ variant }}>
      <Pressable
        className={clsx(
          "w-full h-11 flex-row items-center justify-center rounded-lg gap-2",
          { "bg-lime-300": variant === "primary" },
          { "bg-zinc-800": variant === "secondary" }
        )}
        disabled={isLoading}
        {...props}
      >
        {isLoading ? <ActivityIndicator className="text-lime-950" /> : children}
      </Pressable>
    </ThemeContext.Provider>
  );
}

function Title({ children, ...props }: TextProps) {
  const { variant } = useContext(ThemeContext);
  return (
    <Text
      className={clsx(
        "text-base font-semibold ",
        { "text-lime-950": variant === "primary" },
        { "text-zinc-200": variant === "secondary" }
      )}
      {...props}
    >
      {children}
    </Text>
  );
}

Button.Title = Title;
export { Button };
