import { cn } from "@/lib/utils";
import type { LabelHTMLAttributes } from "react";
import { forwardRef } from "react";

const Label = forwardRef<
  HTMLLabelElement,
  LabelHTMLAttributes<HTMLLabelElement>
>(({ className, ...props }, ref) => {
  return (
    <label
      ref={ref}
      className={cn(
        "text-xs font-medium uppercase tracking-wide text-muted-foreground peer-disabled:cursor-not-allowed peer-disabled:opacity-40",
        className,
      )}
      {...props}
    />
  );
});
Label.displayName = "Label";

export { Label };
