"use client";

import { useState, forwardRef } from "react";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export interface InputPasswordProps extends React.InputHTMLAttributes<HTMLInputElement> {}

const InputPassword = forwardRef<HTMLInputElement, InputPasswordProps>(({ className, ...props }, ref) => {
  const [isVisible, setIsVisible] = useState(false);

  return (
    <div className="relative w-full">
      <Input type={isVisible ? "text" : "password"} className={`pr-10 ${className}`} ref={ref} {...props} />
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => setIsVisible((prevState) => !prevState)}
        className="text-muted-foreground absolute inset-y-0 right-0 h-full w-10 hover:bg-transparent active:scale-100 transition-none"
      >
        {isVisible ? <EyeOffIcon className="size-4" /> : <EyeIcon className="size-4" />}
        <span className="sr-only">{isVisible ? "Hide password" : "Show password"}</span>
      </Button>
    </div>
  );
});

InputPassword.displayName = "InputPassword";

export { InputPassword };
